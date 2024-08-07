package cron

import (
	"fmt"
	"go_code/myselfgo/inits"
	"go_code/myselfgo/notice"
	"go_code/myselfgo/tradgo"
	"go_code/myselfgo/wsclient"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/open-binance/logger"
	"github.com/shopspring/decimal"
)

// 创建超级止损单
// 每一分钟的开始时间判断是否需要创建超级止损单
func syncKlinesMacd(interval int, wg *sync.WaitGroup) {
	defer wg.Done()
	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			for _, symbol := range inits.Config.Symbol.SymbolWhiteList {
				//如果有订单就查仓位
				_, ok := inits.SafeOrderInfo3.GetOrderClientOrderID(symbol)
				if ok {
					go SyncKlinesMacdOnce(symbol)
				}
			}
		}
	}()
}

func SyncKlinesMacdOnce(symbol string) {

	//未开启此功能
	if !inits.Config.Symbol.RootStopMacd {
		return
	}
	var mackState int
	macd, isBool := inits.SafeMacdInfo.GetBySymbol(inits.Config.Symbol.RootSymbol)
	if !isBool {
		logger.Infof("msg=超级止损单没有找到macd数据，symbol=%s,rootsymbol=%s,macd=%v", symbol, inits.Config.Symbol.RootSymbol, macd)
		return
	} else {
		//比较macd.UpdateTime和当前时间，如果时间不对就停止下面代码的执行
		start, end := getCurrentMinuteStartAndEndAsInt64(100)
		logger.Infof("msg=start and end,start=%d,end=%d", start, end)
		if macd.MacdUpdateTime*100 < start {
			logger.Infof("msg=超级止损单没有找到macd的新数据，symbol=%s,rootsymbol=%s,macd=%v", symbol, inits.Config.Symbol.RootSymbol, macd)
			return
		}

	}
	if macd.Macd[3].GreaterThan(decimal.Zero) {
		mackState = 1 //绿柱
	} else {
		mackState = 2 //红柱
	}
	//设置A1成交后对应的订单ID
	cid, ok := inits.SafeOrderInfo3.GetOrderClientOrderID(symbol)
	if !ok {
		return
	}
	rootOrderInfo, isbool := inits.SafeOrderInfo1.GetValue(symbol + cid)
	if !isbool {
		return
	}
	// //buy 是绿柱符合策略，sell是红柱符合策略就不往下执行
	if (rootOrderInfo.OrderLimitSideType == "BUY" && mackState == 1) || (rootOrderInfo.OrderLimitSideType == "SELL" && mackState == 2) {
		logger.Infof("msg=超级止损单不符合创建策略，symbol=%s,siteType=%s,mackState=%d", symbol, rootOrderInfo.OrderLimitSideType, mackState)
		return
	}
	//不符合策略就撤消和清仓
	// [
	// {1721558101775368A1  LIMIT BUY 0.13622 73 FILLED}
	// {1721558101775368B1  STOP SELL 0.1358340639925 146 NEW}
	// {1721558101775368C1  TAKE_PROFIT SELL 0.1366 73 NEW}
	// {1721558101775368A2  LIMIT BUY 0.13618 73 FILLED}
	// {1721558101775368C2  TAKE_PROFIT SELL 0.13656 73 NEW}
	// ]
	var orderBefore string
	if orderinfo, isExists := inits.SafeOrderInfo2.GetValue(symbol); isExists && inits.Config.Order.Open_mb {
		logger.Infof("msg=check orderinfo=%v", orderinfo)
		orderBefore = orderinfo[0].ClientOrderID[0 : len(orderinfo[0].ClientOrderID)-2]
		orderLastCode := orderinfo[0].ClientOrderID[len(orderinfo[0].ClientOrderID)-2:]
		var qty decimal.Decimal
		var superIsStopOrder bool
		superCid := orderBefore + "B0"
		for _, v := range orderinfo {
			logger.Infof("msg=show orderinfo=%v,mackState=%d", v, mackState)
			if v.ClientOrderID == superCid {
				superIsStopOrder = true
			}
			if v.OrderType == "STOP" {
				qty = v.Quantity
			}
		}
		//如果已存在就不创建超级止损单了
		if superIsStopOrder {
			return
		}
		//创建止盈单
		var orderMsg string
		var stopPrice decimal.Decimal
		bestprice, _ := inits.BestPriceInfo.GetBySymbol(symbol)
		if futures.SideType(rootOrderInfo.TakeProfitSideType) == "BUY" {
			stopPrice = decimal.NewFromFloat(bestprice.BestAskPrice)
		} else {
			stopPrice = decimal.NewFromFloat(bestprice.BestBidPrice)
		}
		res3, err := tradgo.CreateStopMarkerOrder(symbol, futures.SideType(rootOrderInfo.OrderStopSideType), qty, superCid)
		if err != nil {
			orderMsg = symbol + "超级止损单创建失败:" + fmt.Sprintf("%v", err.Error())
		} else {
			//撤消所有非超级单且没有成交的
			for _, v := range orderinfo {
				logger.Infof("msg=show orderinfo=%v,mackState=%d", v, mackState)
				if v.OrderStatus == "NEW" && orderLastCode != "B0" {
					tradgo.CanceleOrder(symbol, v.ClientOrderID)
				}
			}
			wsclient.SetOrderInfo(symbol, orderBefore, "MARKET", rootOrderInfo.OrderStopSideType, stopPrice, qty, "NEW")
			orderMsg = symbol + "超级止损单创建创建成功:" + fmt.Sprintf("%v", res3)
			logger.Infof("msg=超级止损单创建成功,symbol=%s,res1:%v", symbol, res3)
		}
		notice.SendDingTalk("[超级止损单预警]" + orderMsg)
		inits.SafeOrderInfo1.Delete(symbol + orderBefore + "A1")
		//删除这个key
	}
}

// 分批止盈
// 两笔进，四次出：
// 1、做空，红柱变短一次，跑一笔，绿住出现跑最后一笔。
// 2、做多，绿柱变短一次，跑一笑，红住出现跑最后一笔。
func SyncStepProfit(interval int, wg *sync.WaitGroup) {
	defer wg.Done()
	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			for _, symbol := range inits.Config.Symbol.SymbolWhiteList {
				_, ok := inits.SafeOrderInfo3.GetOrderClientOrderID(symbol)
				if ok {
					go SyncStepProfitOnce(symbol)
				}
			}
		}
	}()
}

func SyncStepProfitOnce(symbol string) {

	//未开启此功能
	if !inits.Config.Symbol.RootStopMacd {
		return
	}
	var mackState int
	macd, isBool := inits.SafeMacdInfo.GetBySymbol(inits.Config.Symbol.RootSymbol)
	if !isBool {
		logger.Infof("msg=分批止盈单没有找到macd数据，symbol=%s,rootsymbol=%s,macd=%v", symbol, inits.Config.Symbol.RootSymbol, macd)
		return
	} else {
		//比较macd.UpdateTime和当前时间，如果时间不对就停止下面代码的执行
		start, end := getCurrentMinuteStartAndEndAsInt64(100)
		logger.Infof("msg=start and end,start=%d,end=%d", start, end)
		if macd.MacdUpdateTime*100 >= start {
			logger.Infof("msg=分批止盈单没有找到最新macd数据，symbol=%s,rootsymbol=%s,macd=%v", symbol, inits.Config.Symbol.RootSymbol, macd)
			return
		}

	}
	//设置A1成交后对应的订单ID
	cid, ok := inits.SafeOrderInfo3.GetOrderClientOrderID(symbol)
	if !ok {
		return
	}
	rootOrderInfo, isbool := inits.SafeOrderInfo1.GetValue(symbol + cid)
	if !isbool {
		return
	}
	// //buy 是绿柱符合策略，sell是红柱符合策略就不往下执行
	if (rootOrderInfo.OrderLimitSideType == "BUY" && mackState == 1) || (rootOrderInfo.OrderLimitSideType == "SELL" && mackState == 2) {
		logger.Infof("msg=超级止损单不符合创建策略，symbol=%s,siteType=%s,mackState=%d", symbol, rootOrderInfo.OrderLimitSideType, mackState)
		return
	}
	//不符合策略就撤消和清仓
	// [
	// {1721558101775368A1  LIMIT BUY 0.13622 73 FILLED}
	// {1721558101775368B1  STOP BUY 0.1358340639925 146 NEW}
	// {1721558101775368C1  TAKE_PROFIT SELL 0.1366 73 NEW}
	// {1721558101775368A2  LIMIT BUY 0.13618 73 FILLED}
	// {1721558101775368C2  TAKE_PROFIT SELL 0.13656 73 NEW}
	// ]
	var orderBefore string
	if orderinfo, isExists := inits.SafeOrderInfo2.GetValue(symbol); isExists && inits.Config.Order.Open_mb {
		logger.Infof("msg=check orderinfo=%v", orderinfo)
		orderBefore = orderinfo[0].ClientOrderID[0 : len(orderinfo[0].ClientOrderID)-2]
		orderLastCode := orderinfo[0].ClientOrderID[len(orderinfo[0].ClientOrderID)-2:]
		var qty decimal.Decimal
		var superIsStopOrder bool
		superCid := orderBefore + "B0"
		for _, v := range orderinfo {
			logger.Infof("msg=show orderinfo=%v,mackState=%d", v, mackState)
			if v.ClientOrderID == superCid {
				superIsStopOrder = true
			}
			if v.OrderType == "STOP" {
				qty = v.Quantity
			}
		}
		//如果已存在就不创建超级止损单了
		if superIsStopOrder {
			return
		}
		//创建止盈单
		var orderMsg string
		var stopPrice decimal.Decimal
		bestprice, _ := inits.BestPriceInfo.GetBySymbol(symbol)
		if futures.SideType(rootOrderInfo.TakeProfitSideType) == "BUY" {
			stopPrice = decimal.NewFromFloat(bestprice.BestAskPrice)
		} else {
			stopPrice = decimal.NewFromFloat(bestprice.BestBidPrice)
		}
		res3, err := tradgo.CreateStopMarkerOrder(symbol, futures.SideType(rootOrderInfo.OrderStopSideType), qty, superCid)
		if err != nil {
			orderMsg = symbol + "超级止损单创建失败:" + fmt.Sprintf("%v", err.Error())
		} else {
			//撤消所有非超级单且没有成交的
			for _, v := range orderinfo {
				logger.Infof("msg=show orderinfo=%v,mackState=%d", v, mackState)
				if v.OrderStatus == "NEW" && orderLastCode != "B0" {
					tradgo.CanceleOrder(symbol, v.ClientOrderID)
				}
			}
			wsclient.SetOrderInfo(symbol, orderBefore, "MARKET", rootOrderInfo.OrderStopSideType, stopPrice, qty, "NEW")
			orderMsg = symbol + "超级止损单创建创建成功:" + fmt.Sprintf("%v", res3)
			logger.Infof("msg=超级止损单创建成功,symbol=%s,res1:%v", symbol, res3)
		}
		notice.SendDingTalk("[超级止损单预警]" + orderMsg)
		inits.SafeOrderInfo1.Delete(symbol + orderBefore + "A1")
		//删除这个key
	}
}

func getCurrentMinuteStartAndEndAsInt64(precision int) (int64, int64) {
	now := time.Now()

	// 获取当前分钟的开始时间（0秒）
	minuteStart := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())
	minuteStartUnix := minuteStart.Unix()

	// 计算当前分钟开始时的int64表示（百分之一秒或更高精度）
	startAsInt64 := minuteStartUnix * int64(100) // 百分之一秒精度，如果是千分之一秒则乘以1000

	// 计算当前分钟结束时的int64表示（59秒时的值）
	endAsInt64 := startAsInt64 + 59*int64(precision) // 加上59*精度

	return startAsInt64, endAsInt64
}
