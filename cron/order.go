package cron

import (
	"go_code/myselfgo/inits"
	"go_code/myselfgo/notice"
	"sync"
	"time"

	"github.com/open-binance/logger"
)

// 查委托订单
func syncWtOrder(interval int, wg *sync.WaitGroup) {
	defer wg.Done()
	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		for range ticker.C {
			for _, symbol := range inits.Config.Symbol.SymbolWhiteList {
				//如果有订单就查持仓
				ClientOrderID, ok := inits.SafeOrderInfo3.GetOrderClientOrderID(symbol)
				if ok {
					go SyncWtOrderOnce(symbol, ClientOrderID)
				}
			}
		}
	}()
}

func SyncWtOrderOnce(symbol, clientOrderID string) {

	//不符合策略就撤消和清仓
	// [
	// {1721558101775368A1  LIMIT BUY 0.13622 73 FILLED}
	// {1721558101775368B1  STOP BUY 0.1358340639925 146 NEW}
	// {1721558101775368C1  TAKE_PROFIT SELL 0.1366 73 NEW}
	// {1721558101775368A2  LIMIT BUY 0.13618 73 FILLED}
	// {1721558101775368C2  TAKE_PROFIT SELL 0.13656 73 NEW}
	// ]
	var orderMsg string
	if orderinfo, isExists := inits.SafeOrderInfo2.GetValue(symbol); isExists {
		logger.Infof("msg=check orderinfo=%v", orderinfo)
		for _, v := range orderinfo {
			if (v.OrderType == "LIMIT" || v.OrderType == "STOP" || v.OrderType == "TAKE_PROFIT") && v.OrderStatus == "NEW" {
				orderMsg = orderMsg + "交易对:" + symbol + " 委托单号:" + v.ClientOrderID + " 类型:" + getOrderType(v.OrderType) + " 方向:" + v.SideType + " 价格:" + v.Price.String() + " 数量:" + v.Quantity.String() + "\n"
			}
		}

		notice.SendDingTalk("[当前委托预警]\n" + orderMsg)
		//删除这个key
	}
}

func getOrderType(orderType string) string {
	var orderTypeString string
	switch orderType {
	case "LIMIT":
		orderTypeString = "开仓挂单"
	case "MARKET":
		orderTypeString = "市价单"
	case "STOP":
		orderTypeString = "止损单"
	case "TAKE_PROFIT":
		orderTypeString = "止盈单"
	default:
		orderTypeString = "未知类型"
	}
	return orderTypeString
}
