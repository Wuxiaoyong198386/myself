package cron

import (
	"context"
	"go_code/myselfgo/client"
	"go_code/myselfgo/define"
	"go_code/myselfgo/inits"
	"go_code/myselfgo/sqlite"
	"sync"
	"time"

	"github.com/open-binance/logger"
	"github.com/shopspring/decimal"
)

type MarketAmplitude struct {
	Symbol    string `json:"symbol"`     // 合约名称
	StartTime string `json:"start_time"` // 开始时间
	EndTime   string `json:"end_time"`   // 结束时间
	Amplitude bool   `json:"amplitude"`  // 涨跌
}

var marketAmplitudeL = make(map[string][]MarketAmplitude)

func syncMarketRate(interval int, wg *sync.WaitGroup, ch chan error) {
	defer wg.Done()
	logger.Infof("msg=The interval for GetMarketRate filter information||interval=%ds", interval)
	if err := GetMarketRate(); err != nil {
		ch <- err
		return
	}
	go func() {
		//创建一个周期性的定时器
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		//从定时器中获取时间
		for range ticker.C {
			GetMarketRate()
		}
	}()
}

func GetMarketRate() error {

	acount := inits.Config.Account
	apiKey := acount.ApiKey
	secretKey := acount.SecretKey
	ctx, cancel := context.WithTimeout(context.Background(), define.TimeoutBinanceAPI)
	defer cancel()
	exchangeInfo, err := client.GetFutureClient(apiKey, secretKey).NewExchangeInfoService().Do(ctx)
	if err != nil {
		logger.Errorf("获取合约交易对失败", err.Error())
	}
	symbols := []string{}
	for i := 0; i < len(exchangeInfo.Symbols); i++ {
		symbol := exchangeInfo.Symbols[i].Symbol
		if exchangeInfo.Symbols[i].Status != define.StatusTrading { //状态是否为trading
			continue
		}
		lastFour := symbol[len(symbol)-4:]
		if lastFour != "USDT" { // symbol后四位不等于"USDT"跳过
			continue
		}
		symbols = append(symbols, symbol)
	}
	// symbols := []string{"BTCUSDT", "ETHUSDT", "NOTUSDT", "BNBUSDT", "XRPUSDT", "LTCUSDT", "BCHUSDT", "EOSUSDT", "ETCUSDT", "LINKUSDT"}
	var wg sync.WaitGroup
	for _, val := range symbols {
		wg.Add(1)
		go beforMilKline(val, &wg)
	}
	wg.Wait()
	// 计算概率
	// fmt.Println(marketAmplitudeL)
	for i := 0; i < 2; i++ {
		var StartTime, EndTime string
		var equallyNum, notEquallyNum int
		if i == 1 {
			for _, kline := range marketAmplitudeL {
				if kline[i].Amplitude {
					equallyNum++
				} else {
					notEquallyNum++
				}
				StartTime = kline[i].StartTime
				EndTime = kline[i].EndTime
			}
			//写入数据库
			rate_up := float64(equallyNum) / float64(equallyNum+notEquallyNum)
			rate_down := float64(notEquallyNum) / float64(equallyNum+notEquallyNum)
			// fmt.Println(StartTime, EndTime, equallyNum, notEquallyNum, rate_up, rate_down)
			sqlite.MarketInsert(StartTime, EndTime, equallyNum, notEquallyNum, rate_up, rate_down)
		}
	}
	return nil
}
func beforMilKline(symbol string, wg *sync.WaitGroup) error {
	defer wg.Done()
	acount := inits.Config.Account
	apiKey := acount.ApiKey
	secretKey := acount.SecretKey
	ctx, cancel := context.WithTimeout(context.Background(), define.TimeoutBinanceAPI)
	defer cancel()
	// 获取当前时间
	now := time.Now()
	// 截断到分钟（去掉秒和纳秒部分）
	truncated := now.Truncate(time.Minute)
	// 开始时间
	startTime := truncated.Add(-2*time.Minute).Unix() * 1000
	// 前一分钟的结束时间（即当前分钟的第0秒，但通常用于表示一分钟的结束）
	endTime := truncated.Add(-1*time.Minute).Unix() * 1000
	kList, err := client.GetFutureClient(apiKey, secretKey).NewKlinesService().Symbol(symbol).
		Interval("1m").StartTime(startTime).EndTime(endTime).Do(ctx)
	if err != nil {
		logger.Errorf("获取数据失败：%v,error:%v", symbol, err)
	}
	mu.Lock()
	defer mu.Unlock()
	var marketAmplitude []MarketAmplitude
	for _, kline := range kList { // 不需要索引index
		klineOpen, _ := decimal.NewFromString(kline.Open)
		klineClose, _ := decimal.NewFromString(kline.Close)
		diff := klineClose.Sub(klineOpen)
		amplitude := true
		if diff.LessThan(decimal.Zero) {
			amplitude = false
		}
		marketAmplitude = append(marketAmplitude, MarketAmplitude{
			StartTime: time.Unix(kline.OpenTime/1000, 0).Format("2006-01-02 15:04:05"),
			EndTime:   time.Unix(kline.CloseTime/1000, 0).Format("2006-01-02 15:04:05"),
			Symbol:    symbol,    // 使用函数参数中的 symbol
			Amplitude: amplitude, //ture 涨
		})
	}
	// fmt.Println("K线结果：", len(klineAmplitude))
	marketAmplitudeL[symbol] = []MarketAmplitude{}
	marketAmplitudeL[symbol] = append(marketAmplitudeL[symbol], marketAmplitude...)
	return err
}
