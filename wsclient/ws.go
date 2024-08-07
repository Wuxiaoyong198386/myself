package wsclient

import (
	"context"
	"fmt"
	"time"

	"go_code/myselfgo/client"
	"go_code/myselfgo/define"
	"go_code/myselfgo/inits"
	"go_code/myselfgo/modules/datastruct"
	"go_code/myselfgo/tradgo"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"

	"github.com/open-binance/logger"
)

// 用于错误处理的通道
var errChan = make(chan error, 10)
var KlineIsFirstGet = make(map[string]bool, 30)
var KlineIsFirstGetxh = true

// 开始所有的WS主流程
func StartWSProcess() {
	// 获取所有的symbol信息对
	symbolInfoPairs := inits.Config.Symbol.SymbolWhiteList
	// fmt.Println("symbolInfoPairs", symbolInfoPairs)
	// 创建一个空的字符串切片，用于存储symbol
	// 遍历symbol信息对，将key（symbol）添加到symbols切片中

	for _, v := range symbolInfoPairs {
		go startWsCombinedBookTicker(v)
	}
	// 启动goroutine更新账户基本信息

	go startWsUserDataStream()

	//订阅最新K线数据，计算macd
	go startWsKlineDataStreamXh()

	//订阅合约最新K线数据，计算macd，并做交易
	go startWsKlineDataStream()

	// 处理错误
	for err := range errChan {
		fmt.Printf("Error: %v\n", err)
	}
}

/*
 * @description: ws请求订单薄  <symbol>@bookTicker   此方法只执行了三次
 * @fileName: ws.go
 * @author: vip120@126.com
 * @date: 2024-03-25 15:18:27
 */
func startWsCombinedBookTicker(symbol string) {
	//在wsHandler中获取最优价格并进行交易
	wsHandler := func(event *futures.WsBookTickerEvent) {
		// 此交易对已过滤过
		if event == nil {
			logger.Errorf("msg=bookTicker is empty")
			return
		}
		// 更新最优价格和价值,如果最优价格更新失败，则不进行交易
		inits.BestPriceInfo.Update(event)
	}
	errHandler := func(err error) {
		if err != nil {
			logger.Errorf("msg=bookTicker update fail||err=%s", err.Error())
			time.Sleep(5 * time.Second)
			startWsCombinedBookTicker(symbol)
		}
	}
	doneC, stopC, err := futures.WsBookTickerServe(symbol, wsHandler, errHandler)
	//如果失败，5秒中再次请求
	if err != nil {
		logger.Errorf("msg=%s||symbol=%s||err=%s", "WsCombinedBookTickerServe get fail", symbol, err.Error())
		time.Sleep(5 * time.Second)
		startWsCombinedBookTicker(symbol)
		return
	}
	logger.Infof("msg=%s||symbol=%s", "WsCombinedBookTickerServe get success", symbol)
	_ = doneC
	_ = stopC
}

//
/*
 * @description: 启动ws以使用侦听密钥的流命名值
 * @fileName: ws.go
 * @author: vip120@126.com
 * @date: 2024-03-25 15:39:30
 */
func startWsUserDataStream() {
	wsHandler := func(event *futures.WsUserDataEvent) {
		if event == nil {
			logger.Errorf("WsUserDataEvent 订阅事件发生出现异常")
			return
		} else {
			dealWithOrderUpdate(event)
			time.Sleep(3 * time.Second)
			//存款，提款，交易，下单或取消
			dealWithAccountUpdate(event)
		}
	}
	//如果失败，5秒后再次执行
	errHandler := func(err error) {
		if err != nil {
			logger.Errorf("订单状态更新出现异常，错误原因:%s", err.Error())
			time.Sleep(5 * time.Second)
			startWsUserDataStream()
		}
	}
	doneC, stopC, err := futures.WsUserDataServe(inits.ListenKey.Get(), wsHandler, errHandler)
	if err != nil {
		logger.Errorf("msg=%s||err=%s", "startWsUserDataStream fail", err.Error())
		time.Sleep(5 * time.Second)
		startWsUserDataStream()
		return
	} else {
		logger.Infof("订单状态订阅成功")
	}
	_ = doneC
	_ = stopC

}

// 只获取最新K线数据，计算出macd后 不做交易
func startWsKlineDataStreamXh() {
	wsHandler := func(event *binance.WsKlineEvent) {
		if event == nil {
			logger.Errorf("msg=kline is empty")
			return
		} else {
			//当前K线已结束
			if event.Kline.IsFinal {
				klinedata, ok := inits.KlineInfo.GetBySymbol(event.Kline.Symbol)
				if !ok {
					logger.Errorf("msg=kline is not exist")
					return
				}
				//如果是第一次，就更新k1的数据
				if KlineIsFirstGetxh {
					KlineIsFirstGetxh = false
					klinedata[inits.Config.Kline.Kine_count-1].Kline = datastruct.WsKline{
						StartTime:   event.Kline.StartTime,
						EndTime:     event.Kline.EndTime,
						Symbol:      event.Kline.Symbol, // 使用函数参数中的 symbol
						Open:        event.Kline.Open,
						High:        event.Kline.High,
						Low:         event.Kline.Low,
						Close:       event.Kline.Close,
						Volume:      event.Kline.Volume,
						QuoteVolume: event.Kline.QuoteVolume,
						TradeNum:    event.Kline.TradeNum,
					}
				} else {
					//删除第一个元素的值，同时将新的值添加到数组的最后
					klinedata = append(klinedata[1:], datastruct.WsKlineEvent{
						Kline: datastruct.WsKline{
							StartTime:   event.Kline.StartTime,
							EndTime:     event.Kline.EndTime,
							Symbol:      event.Kline.Symbol, // 使用函数参数中的 symbol
							Open:        event.Kline.Open,
							High:        event.Kline.High,
							Low:         event.Kline.Low,
							Close:       event.Kline.Close,
							Volume:      event.Kline.Volume,
							QuoteVolume: event.Kline.QuoteVolume,
							TradeNum:    event.Kline.TradeNum,
						},
					})
				}
				//要更新到全局变量中
				inits.KlineInfo.ReInit(event.Kline.Symbol, klinedata)

				if event.Kline.Symbol == inits.Config.Symbol.RootSymbol && inits.Config.Kline.Macd_open {
					diff, eda, macd := tradgo.GetMacdBySymbol(event.Kline.Symbol, klinedata)
					//更新到全局变量中
					macdresult := datastruct.KlineMacd{
						MacdUpdateTime: event.Kline.StartTime,
						Diff:           diff,
						Eda:            eda,
						Macd:           macd,
					}
					//动态检查root_symbol k1 kline,
					// 直接使用 time.UnixMilli 来创建 time.Time 对象（Go 1.17+）
					t := time.UnixMilli(event.Kline.StartTime)
					inits.SafeMacdInfo.ReInit(inits.Config.Symbol.RootSymbol, macdresult)
					logger.Infof("MacdUpdateTime:%s,%s macd: %+v", t.Format("2006-01-02 15:04:05"), event.Kline.Symbol, macdresult)
				}

			}

		}
	}
	//如果失败，5秒后再次执行
	errHandler := func(err error) {
		if err != nil {
			logger.Errorf("msg=%s||err=%s", "startWsKlineDataStream fail", err.Error())
			time.Sleep(5 * time.Second)
			startWsKlineDataStreamXh()
		}
	}
	var symbols = make(map[string]string)

	symbols[inits.Config.Symbol.RootSymbol] = "1m"

	doneC, stopC, err := binance.WsCombinedKlineServe(symbols, wsHandler, errHandler)
	if err != nil {
		logger.Errorf("msg=%s||err=%s", "startWsKlineDataStream fail", err.Error())
		time.Sleep(5 * time.Second)
		startWsKlineDataStreamXh()
	}
	_ = doneC
	_ = stopC

}

// 获取最新K线数据，计算出macd后 并做交易
func startWsKlineDataStream() {

	wsHandler := func(event *futures.WsKlineEvent) {
		if event == nil {
			return
		} else {
			//当前K线已结束
			if event.Kline.IsFinal {
				klinedata, ok := inits.KlineInfo.GetBySymbol(event.Kline.Symbol)
				if !ok {
					return
				}
				//如果是第一次，就更新k1的数据
				if !KlineIsFirstGet[event.Kline.Symbol] {
					KlineIsFirstGet[event.Kline.Symbol] = true
					klinedata[inits.Config.Kline.Kine_count-1].Kline = datastruct.WsKline{
						StartTime:   event.Kline.StartTime,
						EndTime:     event.Kline.EndTime,
						Symbol:      event.Kline.Symbol, // 使用函数参数中的 symbol
						Open:        event.Kline.Open,
						High:        event.Kline.High,
						Low:         event.Kline.Low,
						Close:       event.Kline.Close,
						Volume:      event.Kline.Volume,
						QuoteVolume: event.Kline.QuoteVolume,
						TradeNum:    event.Kline.TradeNum,
					}
				} else {
					//删除第一个元素的值，同时将新的值添加到数组的最后
					klinedata = append(klinedata[1:], datastruct.WsKlineEvent{
						Kline: datastruct.WsKline{
							StartTime:   event.Kline.StartTime,
							EndTime:     event.Kline.EndTime,
							Symbol:      event.Kline.Symbol, // 使用函数参数中的 symbol
							Open:        event.Kline.Open,
							High:        event.Kline.High,
							Low:         event.Kline.Low,
							Close:       event.Kline.Close,
							Volume:      event.Kline.Volume,
							QuoteVolume: event.Kline.QuoteVolume,
							TradeNum:    event.Kline.TradeNum,
						},
					})
				}
				//要更新到全局变量中
				inits.KlineInfo.ReInit(event.Kline.Symbol, klinedata)

				// if event.Kline.Symbol == inits.Config.Symbol.RootSymbol && inits.Config.Kline.Macd_open {
				// 	diff, eda, macd := tradgo.GetMacdBySymbol(event.Kline.Symbol, klinedata)
				// 	//更新到全局变量中
				// 	macdresult := datastruct.KlineMacd{
				// 		MacdUpdateTime: time.Now().Unix(),
				// 		Diff:           diff,
				// 		Eda:            eda,
				// 		Macd:           macd,
				// 	}
				// 	inits.SafeMacdInfo.ReInit(inits.Config.Symbol.RootSymbol, macdresult)
				// 	logger.Infof("%s macd: %+v", inits.Config.Symbol.RootSymbol, macdresult)
				// }
				//go tradgo.GetRsiBySymbol(event.Kline.Symbol, klinedata)
				//尝试交易
				go tradgo.TryTrade(event.Kline.Symbol, klinedata[79:])
				//超级止损单
				go tradgo.CreateSuperStopOrder(event.Kline.Symbol)
			}
		}
	}
	//如果失败，5秒后再次执行
	errHandler := func(err error) {
		if err != nil {
			logger.Errorf("msg=%s||err=%s", "startWsKlineDataStream fail", err.Error())
			time.Sleep(5 * time.Second)
			startWsKlineDataStream()
		}
	}
	var symbols = make(map[string]string)
	for _, v := range inits.Config.Symbol.SymbolWhiteList {
		symbols[v] = "1m"
	}
	doneC, stopC, err := futures.WsCombinedKlineServe(symbols, wsHandler, errHandler)
	if err != nil {
		logger.Errorf("msg=%s||err=%s", "startWsKlineDataStream fail", err.Error())
		time.Sleep(5 * time.Second)
		startWsKlineDataStream()
	}
	_ = doneC
	_ = stopC

}

func ClientGetKineData(symbol string) error {
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
	startTime := truncated.Add(-21 * time.Minute)
	// 前一分钟的结束时间（即当前分钟的第0秒，但通常用于表示一分钟的结束）
	// 注意：由于Unix时间戳是基于整点的，我们通常使用当前分钟的开始来代表“前一分钟的结束”
	endTime := truncated.Add(-1 * time.Minute)

	// 获取前一分钟结束时间的Unix时间戳（秒为单位）
	endTimeUnix := truncated.Unix()
	//endTimeUnix := getLastMinuteEndTimeMillis()
	ftime := formatTime()

	fmt.Println(startTime.Format("2006ç-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"), ftime)

	_, err := client.GetFutureClient(apiKey, secretKey).NewKlinesService().Symbol(symbol).
		Interval(inits.Config.Kline.Kine_type).StartTime(startTime.Unix() * 1000).EndTime(endTimeUnix*1000 - 1).Do(ctx)
	if err != nil {
		inits.ErrorMsg(inits.ErrorBestPrice, err)
		return err
	}
	return nil
}

func formatTime() string {
	now := time.Now()
	formatted := now.Format("2006-01-02 15:04:05")
	return formatted
}
