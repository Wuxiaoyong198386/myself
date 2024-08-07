package wsclient

import (
	"context"
	"errors"
	"fmt"
	"go_code/myselfgo/client"
	"go_code/myselfgo/inits"
	"go_code/myselfgo/modules/datastruct"
	"go_code/myselfgo/notice"
	"go_code/myselfgo/sqlite"
	"go_code/myselfgo/tradgo"
	"go_code/myselfgo/utils"
	"strconv"
	"time"

	"github.com/open-binance/logger"
	"github.com/shopspring/decimal"

	"github.com/adshao/go-binance/v2/futures"
)

var MoneySum, orderFail, dorderFail, korderFail, dorderSuccess, korderSuccess, orderSuccess decimal.Decimal

// reference: https://binance-docs.github.io/apidocs/spot/cn/#payload-3
/*
 * @description:执行类型 还有一个状态
 * NEW - 新订单已被引擎接受。
 * CANCELED - 订单被用户取消。
 * REJECTED - 新订单被拒绝 （这信息只会在撤消挂单再下单中发生，下新订单被拒绝但撤消挂单请求成功）。
 * TRADE - 订单有新成交。
 * EXPIRED - 订单已根据 Time In Force 参数的规则取消（e.g. 没有成交的 LIMIT FOK 订单或部分成交的 LIMIT IOC 订单）或者被交易所取消（e.g. 强平或维护期间取消的订单）。
 * TRADE_PREVENTION - 订单因 STP 触发而过期。
 * @fileName: user_data.go
 * @author: vip120@126.com
 * @date: 2024-03-27 14:17:37
 */

func dealWithOrderUpdate(event *futures.WsUserDataEvent) {
	if event.Event != futures.UserDataEventTypeOrderTradeUpdate {
		return
	}

	symbol := event.OrderTradeUpdate.Symbol
	//订阅挂单成交
	// 获取ClientOrderID的最后两个字符
	orderBefore := event.OrderTradeUpdate.ClientOrderID[0 : len(event.OrderTradeUpdate.ClientOrderID)-2]
	orderLastCode := event.OrderTradeUpdate.ClientOrderID[len(event.OrderTradeUpdate.ClientOrderID)-2:]
	//止盈或止损动作
	var sideType, orderMsg string
	var side, profit int
	logger.Infof("logid=%s,msg=%s,Symbol=%s,ClientOrderID=%s,Side=%s,Type=%s,OriginalType=%s,ExecutionType=%s,OriginalQty=%s,AccumulatedFilledQty=%s,Status=%s,RealizedPnL=%s,OriginalPrice=%s,StopPrice=%s,CommissionAsset=%s,Commission=%s",
		orderBefore,
		"订单状态订阅",
		event.OrderTradeUpdate.Symbol,               //交易对
		event.OrderTradeUpdate.ClientOrderID,        //订单ID
		event.OrderTradeUpdate.Side,                 //持仓方向
		event.OrderTradeUpdate.Type,                 //订单类型
		event.OrderTradeUpdate.OriginalType,         //原始订单类型
		event.OrderTradeUpdate.ExecutionType,        //执行类型
		event.OrderTradeUpdate.OriginalQty,          //订单原始数量
		event.OrderTradeUpdate.AccumulatedFilledQty, //订单已完成累计数量
		event.OrderTradeUpdate.Status,               //状态
		event.OrderTradeUpdate.RealizedPnL,          //实现利润
		event.OrderTradeUpdate.OriginalPrice,        //原价
		event.OrderTradeUpdate.StopPrice,
		event.OrderTradeUpdate.CommissionAsset, //手续费资产类型
		event.OrderTradeUpdate.Commission,      //手续费数量
	)

	//超级止损单
	if event.OrderTradeUpdate.ExecutionType == "TRADE" && event.OrderTradeUpdate.Status == "FILLED" && event.OrderTradeUpdate.OriginalType == "MARKET" {

		//如果两个止盈单都执行了，才撤消止损单
		OriginalPrice, _ := decimal.NewFromString(event.OrderTradeUpdate.OriginalPrice)
		AccumulatedFilledQty, _ := decimal.NewFromString(event.OrderTradeUpdate.AccumulatedFilledQty)

		logger.Infof("logid=%s,msg=超级止盈单被成交,symbol=%s,ClientOrderID=%s,OriginalType=%s,Side=%s,price=%s,AccumulatedFilledQty=%s",
			orderBefore, symbol, event.OrderTradeUpdate.ClientOrderID,
			string(event.OrderTradeUpdate.OriginalType),
			string(event.OrderTradeUpdate.Side),
			OriginalPrice.String(),
			AccumulatedFilledQty.String())
		time.Sleep(300 * time.Millisecond)
		RealizedPnL, _ := decimal.NewFromString(event.OrderTradeUpdate.RealizedPnL)

		//计算手续费
		uCommissionSum, Commission := Commission(event.OrderTradeUpdate.Commission, event.OrderTradeUpdate.CommissionAsset)
		//统计笔数
		var actionType string
		if RealizedPnL.GreaterThanOrEqual(decimal.Zero) {
			orderSuccess = inits.SafeOrderTjInfo.WOrderSuccess()
			if event.OrderTradeUpdate.Side == "BUY" {
				side = 2
				sideType = "做空"
				korderSuccess = inits.SafeOrderTjInfo.WKorderSuccess()
			} else {
				side = 1
				sideType = "做多"
				dorderSuccess = inits.SafeOrderTjInfo.WDorderSuccess()
			}
			profit = 1
			actionType = "止盈"
		} else {
			orderFail = inits.SafeOrderTjInfo.WOrderFail()
			if event.OrderTradeUpdate.Side == "BUY" {
				side = 2
				sideType = "做空"
				korderFail = inits.SafeOrderTjInfo.WKorderFail()
			} else {
				side = 1
				sideType = "做多"
				dorderFail = inits.SafeOrderTjInfo.WDorderFail()
			}
			profit = 2
			actionType = "止损"
		}
		RealizedPnL_C := RealizedPnL.Sub(Commission)
		MoneySum = inits.SafeOrderTjInfo.WMoneySum(RealizedPnL)
		inits.SafeOrderInfo3.DeleteOrderClientOrderID(symbol)
		logger.Infof("撤单成功，symbol=%s,ClientOrderID=%s", symbol, event.OrderTradeUpdate.ClientOrderID)
		notice.SendDingTalk("[" + sideType + "--" + actionType + "预警]\n交易对:" + event.OrderTradeUpdate.Symbol + ":超级止损单触发。" +
			"\n做空亏损笔数:" + inits.SafeOrderTjInfo.GetKorderFail().String() + "  做空盈利笔数:" + korderSuccess.String() +
			"\n做多亏损笔数:" + inits.SafeOrderTjInfo.GetDorderFail().String() + "  做多盈利笔数:" + dorderSuccess.String() +
			"\n亏损笔数:" + inits.SafeOrderTjInfo.GetOrderFail().String() +
			"\n盈利笔数:" + orderSuccess.String() +
			"\n此订单手续费U:" + Commission.String() +
			"\n此订单盈亏:" + RealizedPnL_C.String() + "(已减手续费)" +
			"\n此订单盈亏:" + RealizedPnL.String() + "(未减手续费)" +
			"\n总手续费U:" + uCommissionSum.String() +
			"\n总盈亏金额:" + MoneySum.String() +
			"\n触发时间:" + utils.FormatTime())
		//写入数据库
		sqlite.StatisticsInsert(orderBefore, event.OrderTradeUpdate.Symbol, side, profit, RealizedPnL, utils.FormatTime())
	}

	//止盈动作
	//1、如果最后一个止盈单被执行，要撤消止损单，如果第一个被执行，不要撤消止损单
	//2、数据统计和消息通知
	if event.OrderTradeUpdate.ExecutionType == "TRADE" && event.OrderTradeUpdate.Status == "FILLED" && event.OrderTradeUpdate.OriginalType == "TAKE_PROFIT" {

		//如果两个止盈单都执行了，才撤消止损单
		OriginalPrice, _ := decimal.NewFromString(event.OrderTradeUpdate.OriginalPrice)
		AccumulatedFilledQty, _ := decimal.NewFromString(event.OrderTradeUpdate.AccumulatedFilledQty)
		SetOrderInfo(symbol, event.OrderTradeUpdate.ClientOrderID,
			string(event.OrderTradeUpdate.OriginalType),
			string(event.OrderTradeUpdate.Side),
			OriginalPrice, AccumulatedFilledQty,
			"FILLED")

		logger.Infof("logid=%s,msg=止盈单被成交,symbol=%s,ClientOrderID=%s,OriginalType=%s,Side=%s,price=%s,AccumulatedFilledQty=%s",
			orderBefore, symbol, event.OrderTradeUpdate.ClientOrderID,
			string(event.OrderTradeUpdate.OriginalType),
			string(event.OrderTradeUpdate.Side),
			OriginalPrice.String(),
			AccumulatedFilledQty.String())

		time.Sleep(300 * time.Millisecond)

		//判断A2是有执行
		a2_ok := inits.SafeOrderInfo1.GetValueA2(event.OrderTradeUpdate.Symbol + orderBefore + "A1")
		if !a2_ok {
			//撤消过单
			var msg string
			a2err := tradgo.CanceleOrder(event.OrderTradeUpdate.Symbol, orderBefore+"A2")
			if a2err != nil {
				msg = symbol + ":A2挂单未成交，撤单失败，错误原因:" + a2err.Error()
			}
			//撤消止损单
			stopa2 := tradgo.CanceleOrder(event.OrderTradeUpdate.Symbol, orderBefore+"B1")
			if stopa2 != nil {
				msg = msg + "\n" + symbol + "止盈后，止损单撤单失败，错误原因:" + a2err.Error()
			}
			if msg != "" {
				notice.SendDingTalk("[异常预警]" + msg)
			}
			//删除全局变量
			inits.SafeOrderInfo1.Delete(event.OrderTradeUpdate.Symbol + orderBefore + "A1")
			inits.SafeOrderInfo2.DeleteValue(symbol)
			inits.SafeOrderInfo3.DeleteOrderClientOrderID(symbol)
		} else {
			//检查执行过几次止盈单2
			exc_count := checkTakeProfit(event.OrderTradeUpdate.Symbol, "TAKE_PROFIT")
			logger.Infof("logid=%s,msg=检查执行过几次止盈单,symbol=%s,count=%d", orderBefore, symbol, exc_count)

			//如果止盈单执行完了，就撤消止损单并清楚全局数据
			if exc_count == 2 {
				tradgo.CanceleOrder(event.OrderTradeUpdate.Symbol, orderBefore+"B1")
				inits.SafeOrderInfo1.Delete(event.OrderTradeUpdate.Symbol + orderBefore + "A1")
				inits.SafeOrderInfo2.DeleteValue(symbol)
				inits.SafeOrderInfo3.DeleteOrderClientOrderID(symbol)
			}
			// else {
			// 	//如果第一笔止盈执行了，就要更改止损单的数量
			// 	//选取消止损单
			// 	tradgo.CanceleOrder(event.OrderTradeUpdate.Symbol, orderBefore+"B1")
			// 	//再重新挂一个止损单
			// 	//要先更改止损单的数量
			// 	filledQty, _ := decimal.NewFromString(event.OrderTradeUpdate.AccumulatedFilledQty)
			// 	safeOrderInfo1, ok := inits.SafeOrderInfo1.GetValue(event.OrderTradeUpdate.Symbol + orderBefore + "A1")
			// 	if !ok {
			// 		logger.Infof("logid=%s,msg=未找到原始订单信息，symbol=%s", orderBefore, symbol)
			// 		return
			// 	}
			// 	stopqty := safeOrderInfo1.OrderStopAllQuantity.Sub(filledQty)
			// 	//创建止损单
			// 	_, err1 := tradgo.CreateStopLimitOrder(symbol,
			// 		futures.SideType(safeOrderInfo1.OrderStopSideType),
			// 		safeOrderInfo1.OrderStopPrice,
			// 		stopqty,
			// 		safeOrderInfo1.OrderClientOrderID)
			// 	if err1 != nil {
			// 		logger.Errorf("logid=%s,msg=止盈一笔订单后更新止损单数失败，但原止损单已撤消，请留言止损。symbol=%s", orderBefore, symbol)
			// 		notice.SendDingTalk("[异常预警]\n" + symbol + "止盈一笔订单后更新止损单数量失败，但原止损单已撤消，请留意止损")
			// 	} else {
			// 		logger.Infof("logid=%s,msg=止盈后更新止损单数成功。symbol=%s", orderBefore, symbol)
			// 	}

			// }

		}

		RealizedPnL, _ := decimal.NewFromString(event.OrderTradeUpdate.RealizedPnL)

		//计算手续费
		uCommissionSum, Commission := Commission(event.OrderTradeUpdate.Commission, event.OrderTradeUpdate.CommissionAsset)
		//统计笔数
		var actionType string
		if RealizedPnL.GreaterThanOrEqual(decimal.Zero) {
			orderSuccess = inits.SafeOrderTjInfo.WOrderSuccess()
			if event.OrderTradeUpdate.Side == "BUY" {
				side = 2
				sideType = "做空"
				korderSuccess = inits.SafeOrderTjInfo.WKorderSuccess()
			} else {
				side = 1
				sideType = "做多"
				dorderSuccess = inits.SafeOrderTjInfo.WDorderSuccess()
			}
			profit = 1
			actionType = "止盈"
		} else {
			orderFail = inits.SafeOrderTjInfo.WOrderFail()
			if event.OrderTradeUpdate.Side == "BUY" {
				side = 2
				sideType = "做空"
				korderFail = inits.SafeOrderTjInfo.WKorderFail()
			} else {
				side = 1
				sideType = "做多"
				dorderFail = inits.SafeOrderTjInfo.WDorderFail()
			}
			profit = 2
			actionType = "止损"
		}
		RealizedPnL_C := RealizedPnL.Sub(Commission)
		MoneySum = inits.SafeOrderTjInfo.WMoneySum(RealizedPnL)
		logger.Infof("撤单成功，symbol=%s,ClientOrderID=%s", symbol, event.OrderTradeUpdate.ClientOrderID)
		notice.SendDingTalk("[" + sideType + "--" + actionType + "预警]\n交易对:" + event.OrderTradeUpdate.Symbol + ":止盈单触发成功,止损单撤消成功。" +
			"\n做空亏损笔数:" + inits.SafeOrderTjInfo.GetKorderFail().String() + "  做空盈利笔数:" + korderSuccess.String() +
			"\n做多亏损笔数:" + inits.SafeOrderTjInfo.GetDorderFail().String() + "  做多盈利笔数:" + dorderSuccess.String() +
			"\n亏损笔数:" + inits.SafeOrderTjInfo.GetOrderFail().String() +
			"\n盈利笔数:" + orderSuccess.String() +
			"\n此订单手续费U:" + Commission.String() +
			"\n此订单盈亏:" + RealizedPnL_C.String() + "(已减手续费)" +
			"\n此订单盈亏:" + RealizedPnL.String() + "(未减手续费)" +
			"\n总手续费U:" + uCommissionSum.String() +
			"\n总盈亏金额:" + MoneySum.String() +
			"\n触发时间:" + utils.FormatTime())
		//写入数据库
		sqlite.StatisticsInsert(orderBefore, event.OrderTradeUpdate.Symbol, side, profit, RealizedPnL, utils.FormatTime())
	}

	//止损动作
	//1、撤消c1和c2，有可能有一个止盈单已这被执行
	//2、判断A2有没有成交，如果没有成交，也要撤消掉
	//3、止损数据统计和消息通知
	if event.OrderTradeUpdate.ExecutionType == "TRADE" && event.OrderTradeUpdate.Status == "FILLED" && event.OrderTradeUpdate.OriginalType == "STOP" {
		//撤单
		FilledPrice, _ := decimal.NewFromString(event.OrderTradeUpdate.LastFilledPrice)
		FilledQty, _ := decimal.NewFromString(event.OrderTradeUpdate.AccumulatedFilledQty)
		SetOrderInfo(symbol, event.OrderTradeUpdate.ClientOrderID,
			string(event.OrderTradeUpdate.OriginalType),
			string(event.OrderTradeUpdate.Side),
			FilledPrice,
			FilledQty, "FILLED")
		logger.Infof("msg=----STOP order update,symbol=%s,ClientOrderID=%s,OriginalType=%s,Side=%s",
			symbol, event.OrderTradeUpdate.ClientOrderID, string(event.OrderTradeUpdate.OriginalType), string(event.OrderTradeUpdate.Side))

		time.Sleep(300 * time.Millisecond)

		//判断A2是有执行
		a2_ok := inits.SafeOrderInfo1.GetValueA2(event.OrderTradeUpdate.Symbol + orderBefore + "A1")
		if !a2_ok {
			//撤消A2挂单
			tradgo.CanceleOrder(event.OrderTradeUpdate.Symbol, orderBefore+"A2")
			//撤消止盈
			tradgo.CanceleOrder(event.OrderTradeUpdate.Symbol, orderBefore+"C1")
			//删除全局变量
			inits.SafeOrderInfo1.Delete(event.OrderTradeUpdate.Symbol + orderBefore + "A1")
			inits.SafeOrderInfo2.DeleteValue(symbol)
			inits.SafeOrderInfo3.DeleteOrderClientOrderID(symbol)
		} else {
			//只有一个止损单，被执行后，因为A2被执行了，就有两个止盈单撤消
			tradgo.CanceleOrder(symbol, orderBefore+"C1")
			tradgo.CanceleOrder(symbol, orderBefore+"C2")
			logger.Infof("msg=止损已执行完，撤消全部止盈单，symbol=%s", symbol)
			//删除全局变量
			inits.SafeOrderInfo1.Delete(event.OrderTradeUpdate.Symbol + orderBefore + "A1")
			inits.SafeOrderInfo2.DeleteValue(symbol)
			inits.SafeOrderInfo3.DeleteOrderClientOrderID(symbol)
		}

		RealizedPnL, _ := decimal.NewFromString(event.OrderTradeUpdate.RealizedPnL)
		//计算手续费
		var actionType string
		uCommissionSum, Commission := Commission(event.OrderTradeUpdate.Commission, event.OrderTradeUpdate.CommissionAsset)
		if RealizedPnL.LessThan(decimal.Zero) {
			orderFail = inits.SafeOrderTjInfo.WOrderFail()
			if event.OrderTradeUpdate.Side == "BUY" {
				side = 2
				sideType = "做空"
				korderFail = inits.SafeOrderTjInfo.WKorderFail()
			} else {
				side = 1
				sideType = "做多"
				dorderFail = inits.SafeOrderTjInfo.WDorderFail()
			}
			profit = 2
			actionType = "止损"
		} else {
			orderSuccess = inits.SafeOrderTjInfo.WOrderSuccess()
			if event.OrderTradeUpdate.Side == "BUY" {
				side = 2
				sideType = "做空"
				korderSuccess = inits.SafeOrderTjInfo.WKorderSuccess()
			} else {
				side = 1
				sideType = "做多"
				dorderSuccess = inits.SafeOrderTjInfo.WDorderSuccess()
			}
			profit = 1
			actionType = "止盈"
		}
		RealizedPnL_C := RealizedPnL.Sub(Commission)
		MoneySum = inits.SafeOrderTjInfo.WMoneySum(RealizedPnL)
		logger.Infof("撤单成功，symbol=%s,ClientOrderID=%s", symbol, event.OrderTradeUpdate.ClientOrderID)
		notice.SendDingTalk("[" + sideType + "--" + actionType + "预警]\n交易对:" + event.OrderTradeUpdate.Symbol + ":止损单触发成功,止盈单撤消成功。" +
			"\n做空亏损笔数:" + korderFail.String() + "  做空盈利笔数:" + inits.SafeOrderTjInfo.GetKorderSuccess().String() +
			"\n做多亏损笔数:" + dorderFail.String() + "  做多盈利笔数:" + inits.SafeOrderTjInfo.GetDorderSuccess().String() +
			"\n亏损笔数:" + orderFail.String() +
			"\n盈利笔数:" + inits.SafeOrderTjInfo.GetOrderSuccess().String() +
			"\n此订单手续费:" + Commission.String() +
			"\n此订单盈亏:" + RealizedPnL_C.String() + "(已减手续费)" +
			"\n此订单盈亏:" + RealizedPnL.String() + "(未减手续费)" +
			"\n总手续费U:" + uCommissionSum.String() +
			"\n总盈亏金额:" + MoneySum.String() +
			"\n触发时间:" + utils.FormatTime())
		//写入数据库
		sqlite.StatisticsInsert(orderBefore, event.OrderTradeUpdate.Symbol, side, profit, RealizedPnL, utils.FormatTime())
	}

	//A1或A2成交后，动作
	//A1成交，创建一个止盈单和一个止损单
	//A2成交，再挂一个中轨止盈单
	if event.OrderTradeUpdate.ExecutionType == "TRADE" && event.OrderTradeUpdate.Status == "FILLED" && event.OrderTradeUpdate.OriginalType == "LIMIT" {
		//防止提交订单还没有返回结果，订阅消息就提前来，所以延时300毫秒
		time.Sleep(300 * time.Millisecond)
		//计算手续费
		Commission(event.OrderTradeUpdate.Commission, event.OrderTradeUpdate.CommissionAsset)
		//A1成交后，第一次创建止盈单
		if orderLastCode == "A1" {

			//重新获取数量，并更新实际交数量
			filledQty, _ := decimal.NewFromString(event.OrderTradeUpdate.AccumulatedFilledQty)
			// inits.SafeOrderInfo1.SetValueQtyA1(symbol+orderBefore+"A1", filledQty)

			LastFilledPrice, _ := decimal.NewFromString(event.OrderTradeUpdate.LastFilledPrice)
			go SetOrderInfo(symbol, event.OrderTradeUpdate.ClientOrderID, "LIMIT", string(event.OrderTradeUpdate.Side), LastFilledPrice, filledQty, "FILLED")
			time.Sleep(300 * time.Microsecond)

			orderinfo, ok := inits.SafeOrderInfo1.GetValue(symbol + event.OrderTradeUpdate.ClientOrderID)
			logger.Warnf("挂单详情%s,", orderinfo)
			if !ok {
				logger.Warnf("logid=%s,msg=未找到A1%s的挂单信息,暂停300毫秒,ClientOrderID=%s", orderBefore, symbol, event.OrderTradeUpdate.ClientOrderID)
				time.Sleep(300 * time.Millisecond)
			}
			logger.Infof("logid=%s,msg=A1成交后,创建止盈单,symbol=%s,orderinfo=%v", orderBefore, symbol, orderinfo)
			if orderinfo.OrderLimitSideType == "SELL" {
				orderMsg = "开空挂单已成交"
			} else if orderinfo.OrderLimitSideType == "BUY" {
				orderMsg = "开多挂单已成功"
			} else {
				orderMsg = "未知挂单类型"
				logger.Warnf("logid=%s,msg=未知挂单类型,symbol=%s,orderinfo=%v", orderBefore, symbol, orderinfo)
				notice.SendDingTalk("[未知挂单类型预警]\n交易对：" + symbol + ",严重错误，未知挂单类型，有可能未找到挂单信息，请检查日志。")
				return
			}
			//创建止损单
			res1, err1 := tradgo.CreateStopLimitOrder(symbol,
				futures.SideType(orderinfo.OrderStopSideType),
				orderinfo.OrderStopPrice,
				orderinfo.OrderStopAllQuantity,
				// filledQty,
				orderinfo.OrderClientOrderID)

			if err1 != nil {
				orderMsg = orderMsg + "\n止损单创建失败:" + fmt.Sprintf("%v", err1.Error())
				logger.Warnf("logid=%s,msg=A1成交后,止损单创建失败,symbol=%s,err=%v", orderBefore, symbol, err1.Error())
				//创建市价止损单
				// market, marketError := tradgo.CreateStopMarkerOrder(symbol, futures.SideType(orderinfo.OrderStopSideType), orderinfo.OrderStopAllQuantity, orderinfo.OrderClientOrderID)
				// if marketError != nil {
				// 	logger.Warnf("logid=%s,msg=A1成交后,止损单市价创建失败,symbol=%s,err=%v", orderBefore, symbol, marketError.Error())
				// } else {
				// 	logger.Warnf("logid=%s,msg=A1成交后,止损单市价创建成功,symbol=%s,res=%v", orderBefore, symbol, market)
				// }
			} else {
				go SetOrderInfo(symbol, orderinfo.OrderClientOrderID, "STOP", orderinfo.OrderStopSideType, orderinfo.OrderStopPrice, filledQty, "NEW")
				orderMsg = orderMsg + "\n止损单创建成功:" + fmt.Sprintf("%v", res1)
				logger.Warnf("logid=%s,msg=A1成交后,止损单创建成功,symbol=%s", orderBefore, symbol)
			}
			//创建止盈单
			res2, err2 := tradgo.CreateTakeProfitOrder(symbol,
				futures.SideType(orderinfo.TakeProfitSideType),
				orderinfo.TakeProfitPrice,
				filledQty,
				orderinfo.TakeProfitClientOrderID)

			if err2 != nil {
				orderMsg = orderMsg + "\n止盈单创建失败:" + fmt.Sprintf("%v", err2.Error())
				// //创建市价止盈单
				// market1, marketError1 := tradgo.CreateStopMarkerOrder(symbol, futures.SideType(orderinfo.OrderStopSideType), filledQty, orderinfo.TakeProfitClientOrderID)
				// if marketError1 != nil {
				// 	logger.Warnf("logid=%s,msg=止盈单市价创建失败,symbol=%s,err=%v", orderBefore, symbol, marketError1.Error())
				// } else {
				// 	logger.Warnf("logid=%s,msg=止盈单市价创建成功,symbol=%s,res=%v", orderBefore, symbol, market1)
				// }
			} else {
				SetOrderInfo(symbol, orderinfo.TakeProfitClientOrderID, "TAKE_PROFIT", orderinfo.TakeProfitSideType, orderinfo.TakeProfitPrice, filledQty, "NEW")
				orderMsg = orderMsg + "\n止盈单创建成功:" + fmt.Sprintf("%v", res2)
			}

			logger.Infof("msg=%s,orderMsg:%v,symbol:%s", orderBefore, orderMsg, event.OrderTradeUpdate.Symbol)
			notice.SendDingTalk("[挂单成交预警]" + orderMsg)
			//写入数据库

			return
		}
		//第二个挂成已成交后，修改止损、止盈单和新建一个止损单到中规价
		if orderLastCode == "A2" {
			// 重新获取数量，并更新实际交数量
			filledQty, _ := decimal.NewFromString(event.OrderTradeUpdate.AccumulatedFilledQty)
			inits.SafeOrderInfo1.SetValueQtyA2(symbol+orderBefore+"A1", filledQty)
			LastFilledPrice, _ := decimal.NewFromString(event.OrderTradeUpdate.LastFilledPrice)
			SetOrderInfo(symbol, event.OrderTradeUpdate.ClientOrderID, "LIMIT", string(event.OrderTradeUpdate.Side), LastFilledPrice, filledQty, "FILLED")
			//设置A2已成交
			inits.SafeOrderInfo1.SetValueA2(symbol+orderBefore+"A1", true)
			time.Sleep(200 * time.Millisecond)
			//重新挂个新的止损单
			if orderinfo, ok := inits.SafeOrderInfo1.GetValue(event.OrderTradeUpdate.Symbol + orderBefore + "A1"); !ok {
				logger.Infof("创建A2时，未找到A1%s的挂单信息,ClientOrderID=%s", event.OrderTradeUpdate.Symbol, event.OrderTradeUpdate.ClientOrderID)
				return
			} else {
				//只要一个止损单
				// stopClientOrderID := orderBefore + "B1"
				// tradgo.CanceleOrder(event.OrderTradeUpdate.Symbol, orderBefore+"B1")
				// res1, err1 := tradgo.CreateStopLimitOrder(event.OrderTradeUpdate.Symbol, futures.SideType(orderinfo.OrderStopSideType), orderinfo.OrderStopPrice, orderinfo.OrderStopAllQuantity, stopClientOrderID)
				// if err1 != nil {
				// 	//如果止损单二次创建失败，一般是原来的价格不存在了，创建失败，直接用现价创建
				// 	stopPriceMap, ok := inits.BestPriceInfo.GetBySymbol(symbol)
				// 	var stopPrice decimal.Decimal
				// 	if !ok {
				// 		logger.Warnf("logid=%s,msg=止损单二次未找到最优价格，symbol=%s,orderinfo=%v", orderBefore, symbol, orderinfo)
				// 	} else {
				// 		if orderinfo.OrderStopSideType == "BUY" {
				// 			stopPrice = decimal.NewFromFloat(stopPriceMap.BestAskPrice)
				// 		} else {
				// 			stopPrice = decimal.NewFromFloat(stopPriceMap.BestBidPrice)
				// 		}

				// 	}
				// 	//直接市价单跑掉
				// 	_, err3 := tradgo.CreateStopMarkerOrder(event.OrderTradeUpdate.Symbol, futures.SideType(orderinfo.OrderStopSideType), orderinfo.OrderStopAllQuantity, stopClientOrderID)
				// 	if err3 != nil {
				// 		orderMsg = "止损单二次创建失败:" + fmt.Sprintf("%v", err1.Error())
				// 	} else {
				// 		SetOrderInfo(symbol, stopClientOrderID, "STOP", orderinfo.OrderStopSideType, stopPrice, orderinfo.OrderStopAllQuantity, "NEW")
				// 		SetOrderInfo(symbol, stopClientOrderID, "STOP", orderinfo.OrderStopSideType, stopPrice, orderinfo.OrderStopAllQuantity, "NEW")
				// 	}

				// } else {
				// 	SetOrderInfo(symbol, stopClientOrderID, "STOP", orderinfo.OrderStopSideType, orderinfo.A2_stop_price, orderinfo.OrderStopAllQuantity, "NEW")
				// 	orderMsg = orderMsg + "\n止损单二次创建成功:" + fmt.Sprintf("%v", res1)
				// 	logger.Infof("msg=止损单二次创建成功,symbol=%s,res1:%v", event.OrderTradeUpdate.Symbol, res1)
				// }
				//再挂个中轨止盈单
				TakeProfitClientOrderID := orderBefore + "C2"
				TakeProfitPrice, ok := tradgo.Bollma(event.OrderTradeUpdate.Symbol, orderinfo.OrderLimitSideType)
				if !ok {
					logger.Infof("msg=无法中轨挂单，symbol=%s,err=%s", event.OrderTradeUpdate.Symbol, "布林数据不全，需等待两分钟。")
				} else {
					res3, err3 := tradgo.CreateTakeProfitOrder(event.OrderTradeUpdate.Symbol, futures.SideType(orderinfo.TakeProfitSideType), TakeProfitPrice, filledQty, TakeProfitClientOrderID)
					if err3 != nil {
						orderMsg = orderMsg + "\n止盈单中轨价创建失败:" + fmt.Sprintf("%v", err3.Error())
					} else {
						SetOrderInfo(symbol, TakeProfitClientOrderID, "TAKE_PROFIT", orderinfo.TakeProfitSideType, TakeProfitPrice, filledQty, "NEW")
						orderMsg = orderMsg + "止盈单中轨价创建成功:" + fmt.Sprintf("%v", res3)
						logger.Infof("msg=止盈单中轨价创建成功,symbol=%s,res1:%v", event.OrderTradeUpdate.Symbol, res3)
					}
				}
			}
			notice.SendDingTalk("[创建止盈单和止损单预警]" + orderMsg)
			return
		}

	}

	//挂单成交
	if event.OrderTradeUpdate.ExecutionType == "TRADE" && event.OrderTradeUpdate.Status == "NEW" && event.OrderTradeUpdate.OriginalType == "LIMIT" {
		OriginalPrice, _ := decimal.NewFromString(event.OrderTradeUpdate.OriginalPrice)    //订单原始价格
		filledQty, _ := decimal.NewFromString(event.OrderTradeUpdate.AccumulatedFilledQty) //累计数量
		SetOrderInfo(symbol, event.OrderTradeUpdate.ClientOrderID, "LIMIT", string(event.OrderTradeUpdate.Side),
			OriginalPrice, filledQty, "NEW")
	}
}

// 账户信息
func dealWithAccountUpdate(event *futures.WsUserDataEvent) {
	if event.Event != futures.UserDataEventTypeAccountUpdate {
		return
	}
	acount := inits.Config.Account
	apiKey := acount.ApiKey
	secretKey := acount.SecretKey
	res, _ := client.GetFutureClient(apiKey, secretKey).NewGetAccountService().Do(context.Background())
	//账户余额信息
	walletmsg := "[余额预警]\n"
	for _, v := range res.Assets {
		balance, _ := strconv.ParseFloat(v.WalletBalance, 64)
		if balance > 0 {
			walletmsg = walletmsg + "资产:" + v.Asset + ",总余额:" + v.WalletBalance + ",可用余额:" + v.AvailableBalance + "\n"
		}
	}
	//持仓信息
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
			retainMsg = retainMsg + "交易对:" + v.Symbol + ",方向:" + side + ",数量:" + v.PositionAmt + ",盈亏:" + v.UnrealizedProfit + "\n"
		}
	}
	notice.SendDingTalk(walletmsg + retainMsg)
	logger.Infof(walletmsg + retainMsg)
}
func GetOrderCount(symobl, state, OrderType string) (int, error) {
	var orderlist []datastruct.OrderInfo2
	var ok bool
	var count int
	if orderlist, ok = inits.SafeOrderInfo2.GetValue(symobl); !ok {
		logger.Errorf("msg=%s,symbol=%s", "在SafeOrderInfo2 Map中未找到定单信息", symobl)
		return 0, errors.New("在SafeOrderInfo2 Map中未找到定单信息")
	}
	for _, v := range orderlist {
		if v.OrderStatus == state && v.OrderType == OrderType {
			count = count + 1
		}
	}
	return count, nil
}

func SetOrderInfo(symbol, clientOrderID, orderType, sideType string, price, quantity decimal.Decimal, orderStatus string) {
	inits.SafeOrderInfo2.SetValue(symbol, datastruct.OrderInfo2{
		ClientOrderID: clientOrderID,
		OrderType:     orderType,
		SideType:      sideType,
		Price:         price,
		Quantity:      quantity,
		OrderStatus:   orderStatus,
	})
	logger.Infof("symbol=setOrderInfo 记录成交后的订单数据，%s,ClientOrderID=%s,OrderType=%s,SideType=%s,OrderStatus=%s，Price=%s,Quantity=%s",
		symbol, clientOrderID, orderType, sideType, orderStatus, price, quantity)
}

// 计算手续费,返回总手续费和此笔的手续费
func Commission(Commission string, CommissionAsset string) (decimal.Decimal, decimal.Decimal) {
	Commission1, _ := decimal.NewFromString(Commission)
	var uCommissionSum decimal.Decimal
	if CommissionAsset == "USDT" {
		uCommissionSum = inits.SafeOrderTjInfo.UsdtCommissionSum(Commission1)
	} else {
		//NBN
		bestPriceMap, _ := inits.BestPriceInfo.GetBySymbol("BNBUSDT")
		bestPrice := decimal.NewFromFloat(bestPriceMap.BestAskPrice)
		Commission1 = Commission1.Mul(bestPrice)
		uCommissionSum = inits.SafeOrderTjInfo.UsdtCommissionSum(Commission1)
	}
	return uCommissionSum, Commission1
}

func checkTakeProfit(symbol string, orderType string) int {
	info, isExists := inits.SafeOrderInfo2.GetValue(symbol)
	var count int
	if isExists {
		for _, v := range info {
			if v.OrderType == orderType && v.OrderStatus == "FILLED" {
				count = count + 1
			}
		}
		return count
	}
	return 0
}
