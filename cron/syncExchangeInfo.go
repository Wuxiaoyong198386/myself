package cron

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go_code/myselfgo/client"
	"go_code/myselfgo/define"
	"go_code/myselfgo/inits"
	"go_code/myselfgo/modules/datastruct"

	"github.com/open-binance/logger"
	"github.com/shopspring/decimal"
)

var (
	TotalValue095  decimal.Decimal
	TotalValueLoss float64
)

/*
 * @description: 函数用于同步价格筛选器信息
 * @fileName: syncExchangeInfo.go
 * @doc:https://binance-docs.github.io/apidocs/spot/cn/#3f1907847c
 * @author: vip120@126.com
 * @date: 2024-03-20 10:09:16
 * @param: interval 同步间隔，单位秒
 * @param: wg *sync.WaitGroup  用于等待同步任务完成
 * @param: ch chan error  错误信息通道，用于返回同步过程中出现的错误信息
 */
func syncPriceFilterInfo(interval int, wg *sync.WaitGroup, ch chan error) {
	defer wg.Done()
	logger.Infof("msg=The interval for synchronizing price filter information||interval=%ds", interval)
	if err := syncExchangeFilterInfo(); err != nil {
		ch <- err
		return
	}
	go func() {
		//创建一个周期性的定时器
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		//从定时器中获取时间
		for range ticker.C {
			syncExchangeFilterInfo()
		}
	}()
}

func syncExchangeFilterInfo() error {
	//获取接口需要用到的签名数据
	acount := inits.Config.Account
	apiKey := acount.ApiKey
	secretKey := acount.SecretKey

	start := time.Now()
	//WithTimeout是Go标准库context中的函数，它接受一个Context和一个超时时间作为参数，返回一个子Context和一个取消函数CancelFunc
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(define.TimeoutBinanceAPI)*time.Second)
	//调用取消函数CancelFunc
	defer cancel()
	//获取数据

	exchangeInfo, err := client.GetFutureClient(apiKey, secretKey).NewExchangeInfoService().Do(ctx)
	if err != nil {
		logger.Errorf("msg=%s||api_key=%s||secret_key=%s||err=%s",
			"Failed to obtain exchange information", apiKey, secretKey, err.Error())
		return err
	}

	//如果没有获取到数据，抛出异常
	if exchangeInfo == nil {
		logger.Errorf("msg=exchange info is empty")
		return errors.New("msg=exchange info is empty")
	}
	//重新设置速率限制信息
	tradingSymbols := make(map[string]bool) // key: symbol
	//定义每个交易对的过滤器的map
	m := make(map[string]datastruct.SpotFilterInfo) // key: symbol
	for i := 0; i < len(exchangeInfo.Symbols); i++ {
		symbol := exchangeInfo.Symbols[i].Symbol
		//如果Status==trading 就设置为true
		if exchangeInfo.Symbols[i].Status == define.StatusTrading {
			tradingSymbols[symbol] = true
		} else {
			continue
		}
		filterInfo := datastruct.SpotFilterInfo{}
		for j := 0; j < len(exchangeInfo.Symbols[i].Filters); j++ {
			filter := exchangeInfo.Symbols[i].Filters[j]
			//PRICE_FILTER 价格过滤器,价格过滤器 用于检测订单中 price 参数的合法性。包含以下三个部分:
			//minPrice 定义了 price/stopPrice 允许的最小值。
			//maxPrice 定义了 price/stopPrice 允许的最大值。
			//tickSize 定义了 price/stopPrice 的步进间隔，即price必须等于minPrice+(tickSize的整数倍)
			//以上每一项均可为0，为0时代表这一项不再做限制。
			//price >= minPrice
			//price <= maxPrice
			//price % tickSize == 0
			if filter["filterType"] == "PRICE_FILTER" {
				filterInfo.Price = datastruct.PriceFilterInfo{
					MinPrice: fmt.Sprintf("%v", filter["minPrice"]),
					MaxPrice: fmt.Sprintf("%v", filter["maxPrice"]),
					TickSize: fmt.Sprintf("%v", filter["tickSize"]),
				}
			}
			//LOT_SIZE 订单尺寸，LOT_SIZE 过滤器对订单中的 quantity 也就是数量参数进行合法性检查。包含三个部分:
			//minQty 表示 quantity/icebergQty 允许的最小值。
			//maxQty 表示 quantity/icebergQty 允许的最大值。
			//stepSize 表示 quantity/icebergQty 允许的步进值。
			if filter["filterType"] == "LOT_SIZE" {
				filterInfo.LotSize = datastruct.LotSizeInfo{
					MinQty:   fmt.Sprintf("%v", filter["minQty"]),
					MaxQty:   fmt.Sprintf("%v", filter["maxQty"]),
					StepSize: fmt.Sprintf("%v", filter["stepSize"]),
				}
			}
			//MARKET_LOT_SIZE过滤器为交易对上的MARKET订单定义了数量(即拍卖中的"手数")规则。 共有3部分：
			//minQty定义了允许的最小quantity。maxQty定义了允许的最大数量。stepSize定义了可以增加/减少数量的间隔。
			//为了通过market lot size，quantity必须满足以下条件：quantity >= minQty and quantity <= maxQty and quantity % stepSize == 0
			if filter["filterType"] == "MARKET_LOT_SIZE" {
				filterInfo.MarketLotSize = datastruct.MarketLotSizeInfo{
					MinQty:   fmt.Sprintf("%v", filter["minQty"]),
					MaxQty:   fmt.Sprintf("%v", filter["maxQty"]),
					StepSize: fmt.Sprintf("%v", filter["stepSize"]),
				}
			}

			if filter["filterType"] == "NOTIONAL" {
				filterInfo.Notional = datastruct.NotionalInfo{
					MinNotional:      fmt.Sprintf("%v", filter["minNotional"]),
					ApplyMinToMarket: fmt.Sprintf("%v", filter["applyMinToMarket"]),
					MaxNotional:      fmt.Sprintf("%v", filter["maxNotional"]),
				}
			}

		}
		m[symbol] = filterInfo
	}

	//确保线程安全的前提下，更新 SafePriceFilterInfoMap 的时间戳和数据，并将其重置为传入的 m 映射。
	inits.SpotPriceFilterInfo.ReInit(m)

	cost := time.Since(start).Seconds() * 1000 // unit: ms
	mlen := len(m)
	logger.Infof(inits.SuccessFindFilter, mlen, mlen, cost)
	return nil
}
