package cron

import (
	"context"
	"fmt"
	"go_code/myselfgo/client"
	"go_code/myselfgo/define"
	"go_code/myselfgo/inits"
	"go_code/myselfgo/notice"
	"strconv"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/open-binance/logger"
	"github.com/shopspring/decimal"
)

func syncRetain(interval int, wg *sync.WaitGroup, ch chan error) {
	defer wg.Done()
	logger.Infof("msg=The interval for syncRetain filter information||interval=%ds", interval)
	if err := SyncRetainServer(); err != nil {
		ch <- err
		return
	}
	go func() {
		//创建一个周期性的定时器
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		//从定时器中获取时间
		for range ticker.C {
			SyncRetainServer()
		}
	}()
}

// 查看当前持仓
func SyncRetainServer() error {
	acount := inits.Config.Account
	apiKey := acount.ApiKey
	secretKey := acount.SecretKey
	//获取当前持仓
	res, _ := client.GetFutureClient(apiKey, secretKey).NewGetAccountService().Do(context.Background())
	retainMsg := "[当前持仓预警]\n"
	for _, v := range res.Positions {
		num, _ := strconv.ParseFloat(v.PositionAmt, 64)
		if num != 0 {
			var side string
			if num < 0 {
				side = "SELL"
			} else {
				side = "BUY"
			}
			retainMsg = retainMsg + "交易对:" + v.Symbol + ",方向:" + side + ",数量:" + v.PositionAmt + ",盈亏:" + v.UnrealizedProfit + "\n"
		}
	}
	notice.SendDingTalk(retainMsg)
	logger.Infof(retainMsg)
	//更改持仓  false单向  true双向
	// res := client.GetFutureClient(apiKey, secretKey).NewChangePositionModeService().DualSide(false).Do(context.Background())
	// if res != nil {
	// 	logger.Errorf("set position mode error: %s", res)
	// } else {
	// 	logger.Infof("set position mode success")
	// }
	return nil
}

// 清仓
func ClearOrderServer() {
	acount := inits.Config.Account
	apiKey := acount.ApiKey
	secretKey := acount.SecretKey
	//获取当前持仓
	res, _ := client.GetFutureClient(apiKey, secretKey).NewGetAccountService().Do(context.Background())
	var clearMsg string
	for _, v := range res.Positions {
		num, _ := strconv.ParseFloat(v.PositionAmt, 64)
		if num != 0 {
			var side, clear string
			if num < 0 {
				side = "BUY"
			} else {
				side = "SELL"
			}
			num1, _ := decimal.NewFromString(v.PositionAmt)
			PositionAmt := num1.Abs()
			_, err := client.GetFutureClient(apiKey, secretKey).NewCreateOrderService().
				Symbol(v.Symbol).
				Side(futures.SideType(side)).
				Type("MARKET").
				Quantity(PositionAmt.String()).
				Do(context.Background())
			if err != nil {
				clear = "失败" + err.Error()
			} else {
				clear = "成功"
			}
			//取消所有订单
			err1 := client.GetFutureClient(apiKey, secretKey).NewCancelAllOpenOrdersService().Symbol(v.Symbol).Do(context.Background())
			if err1 != nil {
				clear = "清仓失败" + err1.Error()
			}
			clearMsg = clearMsg + "交易对:" + v.Symbol + ",方向:" + side + ",数量:" + v.PositionAmt + ",盈亏:" + v.UnrealizedProfit + ",结果:" + clear + "\n"
			fmt.Println("res:", clearMsg)
		}
	}
}

// 持仓
func ShowOrderServer() {
	acount := inits.Config.Account
	apiKey := acount.ApiKey
	secretKey := acount.SecretKey
	//获取当前持仓
	res, err := client.GetFutureClient(apiKey, secretKey).NewGetAccountService().Do(context.Background())
	if err != nil {
		fmt.Printf("get account error: %s", err)
	}
	retainMsg := "[当前持仓]\n"
	for _, v := range res.Positions {
		num, _ := strconv.ParseFloat(v.PositionAmt, 64)
		if num != 0 {
			var side string
			if num < 0 {
				side = "SELL"
			} else {
				side = "BUY"
			}
			fmt.Printf("val:%+v", v)
			retainMsg = retainMsg + "交易对:" + v.Symbol + ",方向:" + side + ",数量:" + v.PositionAmt + ",盈亏:" + v.UnrealizedProfit + "\n"
		}
	}
	fmt.Println(retainMsg)
	//账户余额信息
	walletmsg := "[当前余额]\n"
	for _, v := range res.Assets {
		balance, _ := strconv.ParseFloat(v.WalletBalance, 64)
		if balance > 0 {
			walletmsg = walletmsg + "资产:" + v.Asset + ",总余额:" + v.WalletBalance + ",可用余额:" + v.AvailableBalance + "\n"
		}
	}
	fmt.Println(walletmsg)
}

// 测试
func Test() {
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
	startTime := truncated.Add(-1475*time.Minute).Unix() * 1000
	// 前一分钟的结束时间（即当前分钟的第0秒，但通常用于表示一分钟的结束）
	endTime := truncated.Add(-5*time.Minute).Unix() * 1000
	kList, err := client.GetClient(apiKey, secretKey).NewKlinesService().Symbol("BTCFDUSD").
		Interval("5m").StartTime(startTime).EndTime(endTime).Do(ctx)
	if err != nil {
		fmt.Printf("获取数据失败,error:%+v", err)
	}
	for k, kline := range kList { // 不需要索引index
		key := k + 1

		if key > 6 {
			time := getFormatTime(kline.OpenTime) //开盘时间
			// high := StrToFloat64(kline.High)      //最高价
			// low := StrToFloat64(kline.Low)        //最低价
			// open := StrToFloat64(kline.Open)      //开盘价
			close := StrToFloat64(kline.Close) //收盘价
			//MA7
			var madef float64
			var ma7Str string
			for i := 1; i < 8; i++ {
				madef = madef + StrToFloat64(kList[key-i].Close)
				// fmt.Printf("madef:%v,%v\n", getFormatTime(kList[key-i].CloseTime), kList[key-i].Close)
			}
			ma7 := fmt.Sprintf("%.2f", madef/float64(7))
			if StrToFloat64(ma7) > StrToFloat64(kline.Close) {
				ma7Str = "是"
			} else {
				ma7Str = "否"
			}
			fmt.Printf("key&%v,时间&%v,收盘价&%v,ma&%v,大于收盘价&%v\n", key, time, close, ma7, ma7Str)
			// fmt.Printf("key:%v,时间:%v,最高价：%v,最低价：%v,开盘价：%v,收盘价：%v,ma7:%v,大于收盘价：%v\n", key, time, high, low, open, close, ma7, ma7Str)
		}
	}
}
