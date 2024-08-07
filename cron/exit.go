package cron

import (
	"encoding/json"
	"sync"
	"time"

	"go_code/myselfgo/define"
	"go_code/myselfgo/inits"
	"go_code/myselfgo/utils/spot"

	"github.com/open-binance/logger"
)

func checkExit(interval int) {
	if !inits.Config.Exit.Enable {
		logger.Infof("msg=exit next day is disabled")
		return
	}

	nowTs := time.Now().Unix()
	exitTs := getExitTimestamp()
	logger.Infof("msg=%s||interval=%ds||exit_ts=%d||now_ts=%d||left=%d",
		"interval of checking exit", interval, exitTs, nowTs, exitTs-nowTs)
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for range ticker.C {
		nowTs := time.Now().Unix()
		left := exitTs - nowTs
		if left > 0 {
			logger.Infof("msg=%s||now_ts=%d||exit_ts=%d||left=%d",
				"succeed to check exiting time", nowTs, exitTs, left)
			continue
		}
		logger.Infof("msg=try to exit gracefully because of reaching exiting time")
		spot.GracefulExit(define.ExitMsgReachingExitingTime)
	}
}

func getExitTimestamp() int64 {
	now := time.Now()
	nowTs := now.Unix()
	if inits.Config.Exit.LeftSeconds > 0 {
		return nowTs + inits.Config.Exit.LeftSeconds
	}

	todayLeftSeconds := 86400 - int64(now.Hour()*3600+now.Minute()*60+now.Second())
	tomorrowSeconds := int64(7*3600 + 59*60)
	return nowTs + todayLeftSeconds + tomorrowSeconds - int64(3600*8)
}

// updateReturnRate 函数更新返回率
// 参数：
//
//	wg *sync.WaitGroup：等待组对象，用于同步函数执行
//	interval int：更新间隔，单位秒
//	ch chan error：错误通道，用于传递更新返回率时出现的错误
//
// 返回值：无
func updateReturnRate(wg *sync.WaitGroup, interval int, ch chan error) {
	defer wg.Done()

	// 检查配置文件是否启用了更新返回率的功能
	if !inits.Config.ReturnRateUpdate.Enable {
		logger.Infof("msg=updating return rate is disabled by config file")
		return
	}

	// 执行一次更新返回率的操作
	// update once
	if err := updateReturnRateOnce(getNextTimestamp()); err != nil {
		ch <- err
		return
	}

	go func() {
		// 创建一个定时器，每隔 interval 秒触发一次
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		for range ticker.C {
			nextTs := getNextTimestamp()
			// 忽略错误，持续更新返回率
			updateReturnRateOnce(nextTs) // ignore error is ok
		}
	}()
}

// updateReturnRateOnce 根据给定的时间戳更新收益率
// 参数： nextTs int64 - 下一个时间戳
// 返回值：error - 如果更新过程中发生错误则返回相应的错误信息，否则返回nil
func updateReturnRateOnce(nextTs int64) error {
	nowTs := time.Now().Unix()
	deltaQuoteValue := inits.QuoteValueInfo.GetDeltaQuoteValue()
	delta := nextTs - nowTs
	before := inits.TradeValue.Get()
	beforeRR := inits.ReturnRate.Get()
	var tradeValue float64
	var deltaQuoteValueThrd float64
	var returnRateUpdateInfo inits.ReturnRateConfig

	d0 := inits.Config.ReturnRateUpdate.Info0.DeltaTime
	d1 := inits.Config.ReturnRateUpdate.Info1.DeltaTime
	d2 := inits.Config.ReturnRateUpdate.Info2.DeltaTime
	d3 := inits.Config.ReturnRateUpdate.Info3.DeltaTime
	d4 := inits.Config.ReturnRateUpdate.Info4.DeltaTime
	if delta <= d0 { // 2h
		deltaQuoteValueThrd = inits.Config.ReturnRateUpdate.Info0.DeltaQuoteValue
		tradeValue = inits.Config.ReturnRateUpdate.Info0.TradeValue
		returnRateUpdateInfo = inits.Config.ReturnRateUpdate.Info0.ReturnRate
	} else if delta <= d1 { // 4h
		deltaQuoteValueThrd = inits.Config.ReturnRateUpdate.Info1.DeltaQuoteValue
		tradeValue = inits.Config.ReturnRateUpdate.Info1.TradeValue
		returnRateUpdateInfo = inits.Config.ReturnRateUpdate.Info1.ReturnRate
	} else if delta <= d2 { // 6h
		deltaQuoteValueThrd = inits.Config.ReturnRateUpdate.Info2.DeltaQuoteValue
		tradeValue = inits.Config.ReturnRateUpdate.Info2.TradeValue
		returnRateUpdateInfo = inits.Config.ReturnRateUpdate.Info2.ReturnRate
	} else if delta <= d3 { // 8h
		deltaQuoteValueThrd = inits.Config.ReturnRateUpdate.Info3.DeltaQuoteValue
		tradeValue = inits.Config.ReturnRateUpdate.Info3.TradeValue
		returnRateUpdateInfo = inits.Config.ReturnRateUpdate.Info3.ReturnRate
	} else if delta <= d4 { // 10h
		deltaQuoteValueThrd = inits.Config.ReturnRateUpdate.Info4.DeltaQuoteValue
		tradeValue = inits.Config.ReturnRateUpdate.Info4.TradeValue
		returnRateUpdateInfo = inits.Config.ReturnRateUpdate.Info4.ReturnRate
	} else {
		logger.Infof("msg=stop to update return rate because of delta time||delta_time=%d", delta)
		return nil
	}

	if deltaQuoteValue <= deltaQuoteValueThrd {
		logger.Infof("msg=%s||delta_time=%d||delta_quote_value=%v||delta_quote_value_thrd=%v",
			"stop to update return rate because of delta quote value", delta, deltaQuoteValue, deltaQuoteValueThrd)
		return nil
	}

	logger.Infof("msg=%s||delta_time=%d||delta_quote_value=%v||delta_quote_value_thrd=%v||d0=%d||d1=%d||d2=%d||d3=%d||d4=%d",
		"it will update trade value and return rate", delta, deltaQuoteValue, deltaQuoteValueThrd, d0, d1, d2, d3, d4)

	inits.TradeValue.Set(tradeValue)
	logger.Warnf("msg=succeed to update trade value||before=%v||after=%v", before, inits.TradeValue.Get())

	returnRateInfo := beforeRR // deep copy
	returnRateInfo.Free0.Min = returnRateUpdateInfo.Free0.Min
	returnRateInfo.Free1.Min = returnRateUpdateInfo.Free1.Min
	returnRateInfo.Free2.Min = returnRateUpdateInfo.Free2.Min
	inits.ReturnRate.Set(returnRateInfo)
	beforeRRJSON, _ := json.Marshal(beforeRR)
	afterJSON, _ := json.Marshal(inits.ReturnRate.Get())
	logger.Warnf("msg=succeed to update return rate||before=%v||after=%v", string(beforeRRJSON), string(afterJSON))
	return nil
}

func getNextTimestamp() int64 {
	now := time.Now()
	nowTs := now.Unix()
	todayLeftSeconds := 86400 - int64(now.Hour()*3600+now.Minute()*60+now.Second())
	tomorrowSeconds := int64(8 * 3600)
	return nowTs + todayLeftSeconds + tomorrowSeconds - int64(3600*8)
}
