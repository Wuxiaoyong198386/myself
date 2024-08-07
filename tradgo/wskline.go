package tradgo

import (
	"fmt"
	"go_code/myselfgo/inits"
	"time"

	"github.com/adshao/go-binance/v2"
	json "github.com/json-iterator/go"
	"github.com/open-binance/logger"
)

func StartWsKlinesServe() {
	whiteSymbol := inits.GetSymbolWhiteList()
	if len(whiteSymbol) == 0 {
		logger.Errorf("msg=StartWsKlinesServe no get while symbol")
		return
	}

	// 获取Kline类型并存储在变量中，避免在循环中重复引用
	kineType := inits.Config.Kline.Kine_type

	// 初始化input映射
	input := make(map[string]string, len(whiteSymbol))

	// 遍历whiteSymbol切片并填充input映射
	for _, symbol := range whiteSymbol {
		input[symbol] = kineType
	}

	handler := func(event *binance.WsKlineEvent) {
		WsKlineEvent := binance.WsKlineEvent{
			Event:  event.Event,
			Time:   event.Time,
			Symbol: event.Symbol,
			Kline: binance.WsKline{
				StartTime:            event.Kline.StartTime,            //这根K线开始时间
				EndTime:              event.Kline.EndTime,              //这根K线结束时间
				Symbol:               event.Kline.Symbol,               //交易对
				Interval:             event.Kline.Interval,             //K线间隔
				FirstTradeID:         event.Kline.FirstTradeID,         //这根K线期间第一笔成交ID
				LastTradeID:          event.Kline.LastTradeID,          //这根K线期间末一笔成交ID
				Open:                 event.Kline.Open,                 //开盘价
				High:                 event.Kline.High,                 //根K线期间最高成交价
				Low:                  event.Kline.Low,                  //根K线期间最低成交价
				Close:                event.Kline.Close,                //这根K线期间末一笔成交价
				Volume:               event.Kline.Volume,               //这根K线期间成交量
				IsFinal:              event.Kline.IsFinal,              //这根K线是否完结(是否已经开始下一根K线)
				QuoteVolume:          event.Kline.QuoteVolume,          //这根K线期间成交额
				ActiveBuyVolume:      event.Kline.ActiveBuyVolume,      //主动买入的成交量
				ActiveBuyQuoteVolume: event.Kline.ActiveBuyQuoteVolume, //主动买入的成交额
			},
		}
		jsonstr, err := json.Marshal(WsKlineEvent)
		if err != nil {
			logger.Errorf("msg=StartWsKlinesServe fail||err=%s", err.Error())
			return
		}
		fmt.Println("-------------", string(jsonstr))
	}
	errHandler := func(err error) {
		logger.Errorf("msg=StartWsKlinesServe fail||err=%s", err.Error())
		time.Sleep(5 * time.Second)
		StartWsKlinesServe()
	}

	doneC, stopC, err := binance.WsCombinedKlineServe(input, handler, errHandler)
	//如果订阅失败，停5秒后则重试
	if err != nil {
		logger.Errorf("msg=StartWsKlinesServe fail||err=%s", err.Error())
		time.Sleep(1 * time.Second)
		StartWsKlinesServe()
		return
	}
	logger.Infof("msg=succeed to StartWsKlinesServe")
	_ = doneC
	_ = stopC
}
