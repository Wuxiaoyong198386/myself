package tradgo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go_code/myselfgo/client"
	"go_code/myselfgo/define"
	"go_code/myselfgo/inits"
	"go_code/myselfgo/modules/datastruct"
	"go_code/myselfgo/notice"
	"go_code/myselfgo/sqlite"
	"go_code/myselfgo/utils"
	"go_code/myselfgo/utils/spot"
	"math"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/open-binance/logger"
	"github.com/shopspring/decimal"
)

type OrderInfo struct {
	Symbol   string
	SideType futures.SideType
	K1_high  decimal.Decimal
	K1_low   decimal.Decimal
	K1_open  decimal.Decimal
	K1_close decimal.Decimal
	K2_open  decimal.Decimal
	K2_close decimal.Decimal
	K2_high  decimal.Decimal
	K2_low   decimal.Decimal
	Ma       decimal.Decimal
	Up       decimal.Decimal
	Dn       decimal.Decimal
}

func newContext() context.Context {
	return context.Background()
}

var OldBollSymbol = make(map[string]OrderInfo)
var Oldboll OrderInfo

func TryTrade(symbol string, klineEvents []datastruct.WsKlineEvent) error {
	if symbol == "BNBUSDT" {
		return nil
	}
	var closeSum, mdSum decimal.Decimal
	for _, k := range klineEvents { // 不需要索引index
		// 收盘价之和
		close, _ := decimal.NewFromString(k.Kline.Close)
		closeSum = closeSum.Add(close)
	}
	// 21日移动平均线 中轨线 要四舍五入
	klineEventsLen := int64(len(klineEvents))
	// 将整数转换为decimal.Decimal
	decimalValue := decimal.NewFromInt(klineEventsLen)
	ma := closeSum.Div(decimalValue)

	// 计算每个价格与平均值的差的平方，
	klineEvents_len := len(klineEvents)
	for index := 0; index < klineEvents_len; index++ {
		close, _ := decimal.NewFromString(klineEvents[index].Kline.Close)
		diff := close.Sub(ma)
		mdSum = mdSum.Add(diff.Mul(diff))
	}
	//然后求其平均
	md_avg := mdSum.Div(decimalValue)
	floatVal, _ := md_avg.Float64()
	//求md的平方根
	md := math.Sqrt(floatVal)
	md_decimal := decimal.NewFromFloat(md)
	// 布林带宽度
	places, _ := GetPlaces(symbol) //计算有几位小数
	kFactor := decimal.NewFromInt(2)
	//上轨线
	up := ma.Add(kFactor.Mul(md_decimal))
	up = utils.RoundDecimal(up, places)
	//中轨线
	ma = utils.RoundDecimal(ma, places)
	//下轨线
	dn := ma.Sub(kFactor.Mul(md_decimal))
	dn = utils.RoundDecimal(dn, places)

	k1_high := utils.StrToDecimal(klineEvents[klineEvents_len-1].Kline.High)
	k1_low := utils.StrToDecimal(klineEvents[klineEvents_len-1].Kline.Low)
	k1_open := utils.StrToDecimal(klineEvents[klineEvents_len-1].Kline.Open)
	k1_close := utils.StrToDecimal(klineEvents[klineEvents_len-1].Kline.Close)

	k2_open := utils.StrToDecimal(klineEvents[klineEvents_len-2].Kline.Open)
	k2_close := utils.StrToDecimal(klineEvents[klineEvents_len-2].Kline.Close)
	k2_high := utils.StrToDecimal(klineEvents[klineEvents_len-2].Kline.High)
	k2_low := utils.StrToDecimal(klineEvents[klineEvents_len-2].Kline.Low)

	OrderInfo := &OrderInfo{
		Symbol:   symbol,
		SideType: futures.SideTypeSell,
		K1_high:  k1_high,
		K1_low:   k1_low,
		K1_open:  k1_open,
		K1_close: k1_close,
		K2_open:  k2_open,
		K2_close: k2_close,
		K2_high:  k2_high,
		K2_low:   k2_low,
		Ma:       ma,
		Dn:       dn,
		Up:       up,
	}
	inits.SafeBollInfo.SetBoll(OrderInfo.Symbol, datastruct.BollInfo{Ma: ma, Dn: dn, Up: up})

	//获取k2的boll
	Oldboll = OldBollSymbol[OrderInfo.Symbol]
	//保存到旧值变量中
	OldBollSymbol[OrderInfo.Symbol] = *OrderInfo
	//系统刚启动时第1根k线，oldboll.Dn为0，不计算直接返回
	if Oldboll.Dn.Equal(define.Decimal0) {
		return nil
	}
	// 方向判断是做多还是做空
	if OrderInfo.K1_close.GreaterThan(OrderInfo.K2_high) && OrderInfo.K2_open.GreaterThan(OrderInfo.K2_close) && OrderInfo.K1_close.GreaterThan(OrderInfo.K1_open) {
		// 买多
		// logger.Infof("重要指标：%s 21日移动平均线 %s", symbol, ma.String())
		go BuyOrder(OrderInfo)
	} else if (OrderInfo.K1_close.LessThan(OrderInfo.K2_low) && OrderInfo.K1_open.GreaterThan(OrderInfo.K1_close)) && (OrderInfo.K2_open.LessThan(OrderInfo.K2_close)) {
		// 买空
		// logger.Infof("重要指标：%s 21日移动平均线 %s", symbol, ma.String())
		go SellOrder(OrderInfo)
	}
	var orderinfo []datastruct.OrderInfo2
	var isExists bool
	//判断有没有中轨的止盈单，就不用创建动态止盈单，也就是下面的C2
	cid, ok := inits.SafeOrderInfo3.GetOrderClientOrderID(symbol)
	if !ok {
		//logger.Errorf("msg=%v inits.SafeOrderInfo3.GetOrderClientOrderID，symbol=%s", "未找到A1的订单号", symbol)
		return nil
	}
	if orderinfo, isExists = inits.SafeOrderInfo2.GetValue(symbol); isExists && inits.Config.Order.Open_mb {
		if a2_ok := inits.SafeOrderInfo1.GetValueA2(symbol + cid); a2_ok {
			logger.Infof("msg=check %s,orderinfo=%v", symbol, orderinfo)
			for _, v := range orderinfo {
				orderBefore := v.ClientOrderID[0 : len(v.ClientOrderID)-2]
				orderLastCode := v.ClientOrderID[len(v.ClientOrderID)-2:]
				if orderLastCode == "C2" && v.OrderStatus == "NEW" {
					CanceleOrder(symbol, orderBefore+"C2")
					TakeProfitPrice, _ := Bollma(symbol, v.SideType)
					res3, err3 := CreateTakeProfitOrder(
						symbol,
						futures.SideType(v.SideType),
						TakeProfitPrice,
						v.Quantity,
						orderBefore+"C2")
					if err3 != nil {
						logger.Errorf("msg=动态中轨止盈单创建失败，symbol=%s,err=%s", symbol, err3.Error())
					} else {
						logger.Infof("msg=动态中轨止盈单创建成功，symbol=%s,ClientOrderID=%s,res=%v", symbol, orderBefore+"C2", res3)
					}

				}
			}

		}
	}

	return nil

}

// 开多
func BuyOrder(OrderInfo *OrderInfo) error {
	newClientOrderID := define.NewClientOrderID()
	var macdMsg, orderMsg string
	mulvalue := decimal.NewFromInt(100)
	dntj1, dntj2, dlessma, doUpTj3_bool, bigma, dmacdBool, Bollb := false, false, false, false, false, false, false
	if OrderInfo.K1_close.GreaterThan(OrderInfo.K2_high) && OrderInfo.K2_open.GreaterThan(OrderInfo.K2_close) && OrderInfo.K1_close.GreaterThan(OrderInfo.K1_open) {
		dntj1 = true
	} else {
		return nil
	}

	//k1和 k2必须一条在下轨线上
	if (OrderInfo.K1_low.LessThan(OrderInfo.Dn) && OrderInfo.K1_high.GreaterThan(OrderInfo.Dn)) || (OrderInfo.K2_low.LessThan(Oldboll.Dn) && OrderInfo.K2_high.GreaterThan(Oldboll.Dn)) {
		dntj2 = true
	} else {
		return nil
	}
	//做多：k1的收盘价小于ma，可以做，否侧放弃
	if OrderInfo.K1_close.LessThan(OrderInfo.Ma) {
		dlessma = true
	} else {
		return nil
	}
	//k1的最高价大于中轨和最低价小于中轨价，如果在中轨线上，放弃
	if OrderInfo.K1_high.GreaterThan(OrderInfo.Ma) && OrderInfo.K1_low.LessThan(OrderInfo.Ma) {
		logger.Infof("msg=当前K线在中轨线上,不满足条件,symbol=%s,k1_high=%s,k1_low=%s,ma=%s", OrderInfo.Symbol, OrderInfo.K1_high.String(), OrderInfo.K1_low.String(), OrderInfo.Ma.String())
		return nil
	} else {
		bigma = true
	}
	//做多：上轨价-预警价(收盘价)/预警价＞千三
	doUpTj3Value := OrderInfo.Up.Sub(OrderInfo.K1_close).Div(OrderInfo.K1_close).Mul(decimal.NewFromInt(1000))
	if doUpTj3Value.GreaterThan(decimal.NewFromInt(int64(inits.Config.Custom.DoUpTj3))) {
		doUpTj3_bool = true
	} else {
		logger.Infof("logid=%s,msg=上轨价-预警价(收盘价)/预警价＞千三,symbol=%s", newClientOrderID, OrderInfo.Symbol)
		return nil
	}
	//做多：要求btc永续合约macd当前红柱相对于上一根变长 或 绿柱变红柱
	//Macd:[-6.22 -10.96 -10.09 -10.13]}
	btcmacd, ok := inits.SafeMacdInfo.GetBySymbol(inits.Config.Symbol.RootSymbol)
	if ok && inits.Config.Kline.Macd_open {
		macdMsg = "macd 已开启"
		//做多：要求btc永续合约macd当前绿柱相对于上一根变长 或 红柱变绿柱
		//这里只关注Macd[3]和Macd[2]两根柱状图的比较
		//当前柱状图（Macd[3]）为红柱，且长度大于上一根柱状图（Macd[2]）时，认为做多信号成立
		if btcmacd.Macd[3].GreaterThan(decimal.Zero) && btcmacd.Macd[3].GreaterThan(btcmacd.Macd[2]) {
			dmacdBool = true
		}
		if btcmacd.Macd[3].Equal(decimal.Zero) && btcmacd.Macd[3].GreaterThanOrEqual(btcmacd.Macd[2]) {
			dmacdBool = true
		}
	} else {
		dmacdBool = true
	}

	//布林线差值比=（上轨-下轨）/ 中规,如果小于0.8就放弃
	up_dn := OrderInfo.Up.Sub(OrderInfo.Dn).Div(OrderInfo.Ma).Mul(mulvalue)
	if up_dn.LessThan(decimal.NewFromFloat(inits.Config.Order.Bollb)) {
		logger.Infof("logid=%s,msg=布林线差值比小于%f,symbol=%s,up_dn=%s", newClientOrderID, inits.Config.Order.Bollb, OrderInfo.Symbol, up_dn.String())
		return nil
	} else {
		Bollb = true
	}

	logger.Infof("logid=%s,msg=检查策略所有条件结果,symbol=%s,dntj1=%t,dntj2=%t,bigma=%t,dlessma=%t,doUpTj3_bool=%t,dmacdBool=%t,Bollb=%t",
		newClientOrderID, OrderInfo.Symbol, dntj1, dntj2, bigma, dlessma, doUpTj3_bool, dmacdBool, Bollb)

	if dntj1 && dntj2 && dlessma && bigma && doUpTj3_bool && dmacdBool && Bollb && (inits.Config.Order.Side == 1 || inits.Config.Order.Side == 3) {
		logger.Infof("logid=%s,msg=%s,symbol=%s", "下轨反包做多有信号,开始挂单", newClientOrderID, OrderInfo.Symbol)
		// stopPrice := GetStopPrice(OrderInfo.Symbol, futures.SideTypeSell, OrderInfo.K1_close, OrderInfo.K1_high)
		stopPrice := GetStopPrice(futures.SideTypeBuy, OrderInfo.K1_high, OrderInfo.K2_high, OrderInfo.K1_low, OrderInfo.K2_low)
		//获取交易对的价格
		bestPriceMap, ok := inits.BestPriceInfo.GetBySymbol(OrderInfo.Symbol)
		bestPrice := decimal.NewFromFloat(bestPriceMap.BestAskPrice)
		stopPriceBool := true
		if !ok {
			stopPriceBool = false
			logger.Infof("logid=%s,msg=%s,symbol=%s", newClientOrderID, "best price not found", OrderInfo.Symbol)
			return nil
		} else if stopPrice.GreaterThan(bestPrice) {
			//开多单，止损价大于订单价格，说明往下跳水了，放弃做多
			stopPriceBool = false
			logger.Infof("logid=%s,msg=%s,symbol=%s,stop_price=%s,best_price=%s", newClientOrderID, "止损价大于订单价格，往下跳水了，放弃做多", OrderInfo.Symbol, stopPrice, bestPrice)
			return nil
		}
		//如果止损大于配置值，就直接设置为配置值
		//做多，买价大于卖价，所以要用买价减去卖价
		//亏损百分比 = ((买价 - 卖价) / 卖价) * 100=0.5
		// if bestPrice.Sub(stopPrice).Div(stopPrice).Mul(mulvalue).GreaterThan(mulvalue.Mul(decimal.NewFromFloat(inits.Config.Order.Max_loss))) {
		// 	stopPrice = bestPrice.Mul(define.Decimal1.Sub(decimal.NewFromFloat(inits.Config.Order.Max_loss)))
		// 	logger.Infof("logid=%s,msg=止损大于%s，设置为最大%f,symbol=%f,stopprice=%s", newClientOrderID, inits.Config.Order.Max_loss, inits.Config.Order.Max_loss, OrderInfo.Symbol, stopPrice.String())
		// }

		if inits.Config.Order.Enable {
			if stopPriceBool {
				//市场情绪
				reta := sqlite.MarketQuery()
				rateStr := "市场情绪:[涨" + fmt.Sprintf("%.2f", reta.UpRate) + "],跌[" + fmt.Sprintf("%.2f", reta.DownRate) + "]"
				if reta.DownRate > inits.Config.Order.Market_rate || reta.UpRate > inits.Config.Order.Market_rate {
					orderMsg = rateStr + ",放弃做多"
				} else {
					//创建订单,判断持仓模式 1:独仓模式 2:多仓模式
					if inits.Config.Order.Warehouse_mode == 1 {
						price := spot.PriceCorrection(OrderInfo.Symbol, bestPrice.String())
						// price := spot.PriceCorrection(OrderInfo.Symbol, OrderInfo.K1_low.Add(OrderInfo.K1_close).Div(decimal.NewFromInt(2)).String())
						//下单数量校验
						quantity := decimal.NewFromInt(int64(inits.Config.Order.Default_quantity))
						quantity = spot.QuantityCorrection(OrderInfo.Symbol, quantity.Div(price).String())
						_, quantity, err := CreateLimitOrder(OrderInfo.Symbol, futures.SideTypeBuy, price, quantity, newClientOrderID+"A1")
						if err != nil {
							orderMsg = "开多单创建失败:" + OrderInfo.Symbol + "失败原因" + err.Error()
						} else if quantity.GreaterThan(decimal.Zero) {
							TakeProfitPrice := GetTakeProfitPrice(futures.SideTypeBuy, OrderInfo.K1_close)
							orderInfo := &datastruct.OrderLimitInfo{
								Symbol:             OrderInfo.Symbol,
								ClientOrderID:      newClientOrderID,
								OrderStopSideType:  string(futures.SideTypeSell),
								OrderStopPrice:     stopPrice,
								OrderStopQuantity:  quantity,
								OrderClientOrderID: newClientOrderID + "B1",

								OrderLimitSideType:      string(futures.SideTypeBuy),
								TakeProfitSideType:      string(futures.SideTypeSell),
								TakeProfitPrice:         TakeProfitPrice,
								TakeProfitQuantity:      quantity,
								TakeProfitClientOrderID: newClientOrderID + "C1",
							}
							inits.SafeOrderInfo1.SetValue(OrderInfo.Symbol+newClientOrderID+"A1", orderInfo)
							//设置A1成交后对应的订单ID
							inits.SafeOrderInfo3.SetOrderClientOrderID(OrderInfo.Symbol, newClientOrderID+"A1")
							orderMsg = "开多挂单已提交，待成交:ClientOrderID:" + newClientOrderID + "A1"
						}
					} else {
						//一仓价格 预警价（收盘价）
						price1 := spot.PriceCorrection(OrderInfo.Symbol, bestPrice.String())
						//二仓价格  最低价和预警价和的二分之一
						price2 := spot.PriceCorrection(OrderInfo.Symbol, OrderInfo.K1_low.Add(OrderInfo.K1_close).Div(decimal.NewFromInt(2)).String())
						//下单数量校验
						quantity_config, _ := decimal.NewFromString(inits.Config.Order.Warehouse_cnt[0])
						quantity_safe := spot.QuantityCorrection(OrderInfo.Symbol, quantity_config.Div(price1).String())
						_, quantity_order, err1 := CreateLimitOrder(OrderInfo.Symbol, futures.SideTypeBuy, price1, quantity_safe, newClientOrderID+"A1")
						if err1 != nil {
							orderMsg = "开多一单创建失败:" + OrderInfo.Symbol + "失败原因:" + err1.Error()
							logger.Errorf("logid=%s,error=%s", OrderInfo.Symbol, orderMsg)
							notice.SendDingTalk("[开空单创建失败预警]" + orderMsg)
							return nil
						}
						quantity2_config, _ := decimal.NewFromString(inits.Config.Order.Warehouse_cnt[1])
						quantity2_safe := spot.QuantityCorrection(OrderInfo.Symbol, quantity2_config.Div(price2).String())
						_, quantity2_order, err2 := CreateLimitOrder(OrderInfo.Symbol, futures.SideTypeBuy, price2, quantity2_safe, newClientOrderID+"A2")
						if err2 != nil {
							quantity2_order = decimal.NewFromFloat(0)
							orderMsg = "开多二单创建失败:" + OrderInfo.Symbol + "失败原因:" + err2.Error()
							logger.Errorf("logid=%s,error=%s", newClientOrderID, orderMsg)
							notice.SendDingTalk("[开空单创建失败预警]" + orderMsg)
						}
						//如果stopprice大于千五，就直接设置为千五
						//盈利百分比 = ((卖价 - 买价) / 买价) * 100=0.5
						if stopPrice.Sub(bestPrice).Div(bestPrice).Mul(decimal.NewFromFloat(100)).GreaterThan(decimal.NewFromFloat(100).Mul(decimal.NewFromFloat(inits.Config.Order.Max_loss))) {
							stopPrice = bestPrice.Mul(define.Decimal1.Sub(decimal.NewFromFloat(inits.Config.Order.Stop_loss)))
							logger.Infof("logid=%s,msg=止损大于%f，设置为最大%f,symbol=%s,stopprice=%s", newClientOrderID, inits.Config.Order.Max_loss, inits.Config.Order.Max_loss, OrderInfo.Symbol, stopPrice.String())
						}

						if quantity_order.GreaterThan(decimal.Zero) || quantity2_order.GreaterThan(decimal.Zero) {
							takeProfitPrice := GetTakeProfitPrice(futures.SideTypeBuy, OrderInfo.K1_close)
							//如果止盈价不到中轨，就用中轨价
							if takeProfitPrice.LessThan(OrderInfo.Ma) {
								takeProfitPrice = OrderInfo.Ma
							}
							//A2止损在原来的基础上下调万五
							a2stopprice := stopPrice
							takeProfitQuantity := quantity_safe.Add(quantity2_safe)
							logger.Warnf("做空数量,takeProfitQuantity=%v,quantity_safe=%s,quantity2_safe=%v", takeProfitQuantity, quantity_safe, quantity2_safe)
							orderInfo := &datastruct.OrderLimitInfo{
								Symbol:               OrderInfo.Symbol,
								ClientOrderID:        newClientOrderID,
								OrderStopSideType:    string(futures.SideTypeSell),
								OrderStopPrice:       stopPrice,
								OrderStopQuantity:    quantity_order,
								OrderStopAllQuantity: takeProfitQuantity,
								OrderClientOrderID:   newClientOrderID + "B1",
								A2_stop_price:        a2stopprice,

								OrderLimitSideType:      string(futures.SideTypeBuy),
								TakeProfitSideType:      string(futures.SideTypeSell),
								TakeProfitPrice:         takeProfitPrice,
								TakeProfitQuantity:      quantity2_order,
								TakeProfitClientOrderID: newClientOrderID + "C1",
							}
							//把orderInfo保存到全局变量中
							inits.SafeOrderInfo1.SetValue(OrderInfo.Symbol+newClientOrderID+"A1", orderInfo)
							//设置A1成交后对应的订单ID
							inits.SafeOrderInfo3.SetOrderClientOrderID(OrderInfo.Symbol, newClientOrderID+"A1")
							orderMsg = "开多挂单已提交，待成交:ClientOrderID:" + newClientOrderID + "A1," + newClientOrderID + "A2"
						}
					}
					orderMsg = orderMsg + "\n" + rateStr
				}
			} else {
				orderMsg = "最优价已小于止损价，不创建订单,stopPrice=" + stopPrice.String() + ",bestPrice=" + bestPrice.String()
			}
		} else {
			orderMsg = "未开启创建订单功能，只是预警"
		}

		ftime := utils.FormatTime()
		logger.Infof("msg=%s,time=%s,symbol=%s,up=%s,ma=%s,dn=%s", "合约，快乐预警:下轨反包做多有信号",
			ftime, OrderInfo.Symbol, OrderInfo.Up, OrderInfo.Ma, OrderInfo.Dn)
		logger.Infof("k1_close > k2_high(%s,%s) && k2_open > k2_close(%s,%s) && k1_close > k1_open(%s,%s)", OrderInfo.K1_close.String(), OrderInfo.K2_high.String(), OrderInfo.K2_open.String(), OrderInfo.K2_close.String(), OrderInfo.K1_close.String(), OrderInfo.K1_open.String())
		logger.Infof("k1_low < dn(%s,%s),k1_high > dn(%s,%s),k2_low < dn(%s,%s),k2_high > dn(%s,%s)", OrderInfo.K1_low.String(), OrderInfo.Dn.String(), OrderInfo.K1_high.String(), OrderInfo.Dn.String(), OrderInfo.K2_low.String(), OrderInfo.Dn.String(), OrderInfo.K2_high.String(), OrderInfo.Dn.String())
		logger.Infof("symbol=%s,(up-k1_close)/k1_close(%s>%d)", OrderInfo.Symbol, doUpTj3Value.String(), inits.Config.Custom.DoUpTj3)

		msg := "策略名称：反包策略预警" + macdMsg + "\n 交易对：" + OrderInfo.Symbol + "\n 方向：多 \n 预警时间：" + ftime + "\n预警价格：" + OrderInfo.K1_close.String() + " \n上一根K线的最低价：" + OrderInfo.K2_low.String() + "\n这一根k线的收盘价格：" + OrderInfo.K1_close.String() + "\n 布林线差值比：" + up_dn.String() + "\n" + orderMsg
		notice.SendDingTalk(msg)
		//写入数据库
		sqlite.WarngingInsert(newClientOrderID, OrderInfo.Symbol, 1, OrderInfo.K1_close, OrderInfo.K2_low, OrderInfo.K1_close, up_dn, ftime)
	}
	return nil
}

// 开空
func SellOrder(OrderInfo *OrderInfo) error {
	uptj1, uptj2, lessma, doDownTj3_bool, kmacdBool, bigma := false, false, false, false, false, false
	// k1是阴，k2是阳 阴包阳
	if (OrderInfo.K1_close.LessThan(OrderInfo.K2_low)) && (OrderInfo.K1_open.GreaterThan(OrderInfo.K1_close)) && (OrderInfo.K2_open.LessThan(OrderInfo.K2_close)) {
		logger.Infof("msg=条件1，阴包阳条件符合,symbol=%s", OrderInfo.Symbol)
		uptj1 = true
	} else {
		return nil
	}
	// 必须有一根k线在上轨上
	if (OrderInfo.K1_high.GreaterThan(OrderInfo.Up) && OrderInfo.K1_low.LessThan(OrderInfo.Up)) || (OrderInfo.K2_high.GreaterThan(Oldboll.Up) && OrderInfo.K2_low.LessThan(Oldboll.Up)) {
		uptj2 = true
	} else {
		return nil
	}
	//做空：(预警价-下轨价)/预警价＞千3
	doDownTj3value1 := OrderInfo.K1_close.Sub(OrderInfo.Dn).Div(OrderInfo.K1_close).Mul(decimal.NewFromInt(1000))
	doDownTj3_bool = doDownTj3value1.GreaterThanOrEqual(decimal.NewFromInt(int64(inits.Config.Custom.DoDownTj3)))
	if doDownTj3_bool {
		logger.Infof("msg=条件2，(预警价-下轨价)/预警价＞千3,symbol=%s,doDownTj3value1=%s,OrderInfo.K1_close=%s,dn=%s,doDownTj3_bool=%t",
			OrderInfo.Symbol, doDownTj3value1.String(), OrderInfo.K1_close.String(), OrderInfo.Dn.String(), doDownTj3_bool)

	} else {
		logger.Infof("msg=DoDownTj3,symbol=%s,doDownTj3value1=%s,OrderInfo.K1_close=%s,dn=%s,doDownTj3_bool=%t",
			OrderInfo.Symbol, doDownTj3value1.String(), OrderInfo.K1_close.String(), OrderInfo.Dn.String(), doDownTj3_bool)
		return nil
	}
	// 如果K1的收盘价小于ma，放弃
	if OrderInfo.K1_close.GreaterThan(OrderInfo.Ma) {
		logger.Infof("msg=条件3，K1的收盘价大于ma,symbol=%s", OrderInfo.Symbol)
		lessma = true
	} else {
		return nil
	}
	var macdMsg, orderMsg string
	mulvalue := decimal.NewFromInt(100)
	btcmacd, ok := inits.SafeMacdInfo.GetBySymbol(inits.Config.Symbol.RootSymbol)
	if ok && inits.Config.Kline.Macd_open {
		macdMsg = "macd 已开启"
		//做空时：要求btc永续合约macd当前红柱相对于上一根变长 或 绿柱变红柱
		if btcmacd.Macd[3].LessThan(decimal.Zero) {
			if btcmacd.Macd[3].LessThan(btcmacd.Macd[2]) {
				kmacdBool = true
			}
		}
		if btcmacd.Macd[3].Equal(decimal.Zero) {
			if btcmacd.Macd[3].LessThanOrEqual(btcmacd.Macd[2]) {
				kmacdBool = true
			}
		}
		if kmacdBool {
			logger.Infof("msg=条件4，符合macd策略,symbol=%s", OrderInfo.Symbol)
		} else {
			return nil
		}
	} else {
		kmacdBool = true
	}
	//k1如果在中轨线上，放弃
	if OrderInfo.K1_high.GreaterThan(OrderInfo.Ma) && OrderInfo.K1_low.LessThan(OrderInfo.Ma) {
		logger.Infof("msg=当前K线在中轨线上,symbol=%s,k1_high=%s,k1_low=%s,ma=%s", OrderInfo.Symbol, OrderInfo.K1_high.String(), OrderInfo.K1_low.String(), OrderInfo.Ma.String())
		return nil
	} else {
		logger.Infof("msg=条件5，k1不在中轨线上,symbol=%s", OrderInfo.Symbol)
		bigma = true
	}

	logger.Warnf("msg=tj check result,symbol=%s,uptj1=%t,uptj2=%t,lessma=%t,doDownTj3_bool=%t,kmacdBool=%t",
		OrderInfo.Symbol, uptj1, uptj2, lessma, doDownTj3_bool, kmacdBool)
	if uptj1 && uptj2 && lessma && bigma && doDownTj3_bool && kmacdBool && (inits.Config.Order.Side == 2 || inits.Config.Order.Side == 3) {
		logger.Infof("msg=%s,symbol=%s", "合约，快乐预警:上轨反包做空有信号,开始挂单", OrderInfo.Symbol)
		//获取交易对的价格
		bestPriceMap, ok := inits.BestPriceInfo.GetBySymbol(OrderInfo.Symbol)
		bestPrice := decimal.NewFromFloat(bestPriceMap.BestBidPrice)
		//计算止损价
		// stopPrice := GetStopPrice(OrderInfo.Symbol, futures.SideTypeSell, OrderInfo.K1_close, OrderInfo.K1_high)
		stopPrice := GetStopPrice(futures.SideTypeSell, OrderInfo.K1_high, OrderInfo.K2_high, OrderInfo.K1_low, OrderInfo.K2_low)
		//如果stopprice大于千五，就直接设置为千五
		//盈利百分比 = ((卖价 - 买价) / 买价) * 100=0.5
		//如果stopprice大于千五，就直接设置为千五
		//盈利百分比 = ((卖价 - 买价) / 买价) * 100=0.5
		// if stopPrice.Sub(bestPrice).Div(bestPrice).Mul(decimal.NewFromFloat(100)).GreaterThan(decimal.NewFromFloat(100).Mul(decimal.NewFromFloat(inits.Config.Order.Max_loss))) {
		// 	stopPrice = bestPrice.Mul(decimal.NewFromFloat(1).Add(decimal.NewFromFloat(inits.Config.Order.Max_loss)))
		// 	logger.Infof("msg=止损大于千五，设置为最大千五,symbol=%s,stopprice=%s", OrderInfo.Symbol, stopPrice.String())
		// }

		//防止最优量没有订阅到，或者最近订部失败，则使用k1_close
		if bestPrice.LessThan(decimal.Zero) {
			bestPrice = decimal.NewFromFloat(bestPriceMap.BestAskPrice)
		}
		if bestPrice.GreaterThan(OrderInfo.K1_close) {
			bestPrice = OrderInfo.K1_close
		}
		logger.Infof(OrderInfo.Symbol+"bestPrice=%s,k1_close=%s", bestPrice.String(), OrderInfo.K1_close)

		stopPriceBool := true
		if !ok {
			inits.ErrorMsg(inits.ErrorBestPrice, errors.New("best price not found"))
			stopPriceBool = false
		} else if bestPrice.GreaterThan(stopPrice) {
			//开空单，如果最优价已大于止损价，则不创建订单
			stopPriceBool = false
		}
		var order_id string
		if inits.Config.Order.Enable {
			if !stopPriceBool {
				orderMsg = "最优价已大于止损价，不创建订单,stopPrice=" + stopPrice.String() + ",bestPrice=" + bestPrice.String()
			} else {
				newClientOrderID := define.NewClientOrderID()
				order_id = newClientOrderID
				//市场情绪
				reta := sqlite.MarketQuery()
				rateStr := "市场情绪:[涨" + fmt.Sprintf("%.2f", reta.UpRate) + "],跌[" + fmt.Sprintf("%.2f", reta.DownRate) + "]"
				if reta.DownRate > inits.Config.Order.Market_rate || reta.UpRate > inits.Config.Order.Market_rate {
					orderMsg = rateStr + ",放弃做空"
				} else {
					//创建订单,判断持仓模式 1:独仓模式 2:多仓模式
					if inits.Config.Order.Warehouse_mode == 1 {
						price := spot.PriceCorrection(OrderInfo.Symbol, bestPrice.String())
						//下单数量校验
						quantity := decimal.NewFromInt(int64(inits.Config.Order.Default_quantity))
						quantity = spot.QuantityCorrection(OrderInfo.Symbol, quantity.Div(price).String())
						_, _, err := CreateLimitOrder(OrderInfo.Symbol, futures.SideTypeSell, price, quantity, newClientOrderID+"A1")
						if err != nil {
							orderMsg = "开多单创建失败:" + OrderInfo.Symbol + "失败原因" + err.Error()
						} else {
							TakeProfitPrice := GetTakeProfitPrice(futures.SideTypeSell, OrderInfo.K1_close)
							orderInfo := &datastruct.OrderLimitInfo{
								Symbol:             OrderInfo.Symbol,
								ClientOrderID:      newClientOrderID,
								OrderStopSideType:  string(futures.SideTypeBuy),
								OrderStopPrice:     stopPrice,
								OrderStopQuantity:  quantity,
								OrderClientOrderID: newClientOrderID + "B1",

								OrderLimitSideType:      string(futures.SideTypeSell),
								TakeProfitSideType:      string(futures.SideTypeBuy),
								TakeProfitPrice:         TakeProfitPrice,
								TakeProfitQuantity:      quantity,
								TakeProfitClientOrderID: newClientOrderID + "C1",
							}
							//把orderInfo保存到全局变量中
							inits.SafeOrderInfo1.SetValue(OrderInfo.Symbol+newClientOrderID+"A1", orderInfo)
							//设置A1成交后对应的订单ID
							inits.SafeOrderInfo3.SetOrderClientOrderID(OrderInfo.Symbol, OrderInfo.Symbol+newClientOrderID+"A1")
							orderMsg = "开空单挂单已提交，待成交,ClientOrderID:" + newClientOrderID
						}
					} else {
						//一仓价格 预警价（收盘价）
						price1 := spot.PriceCorrection(OrderInfo.Symbol, bestPrice.String())
						//二仓价格  最高价和预警价和的二分之一
						price2 := spot.PriceCorrection(OrderInfo.Symbol, OrderInfo.K1_high.Add(OrderInfo.K1_close).Div(decimal.NewFromInt(2)).String())
						//下单数量校验
						quantity1, _ := decimal.NewFromString(inits.Config.Order.Warehouse_cnt[0])
						quantity1_safe := spot.QuantityCorrection(OrderInfo.Symbol, quantity1.Div(price1).String())
						_, quantity1_order, err1 := CreateLimitOrder(OrderInfo.Symbol, futures.SideTypeSell, price1, quantity1_safe, newClientOrderID+"A1")
						if err1 != nil {
							orderMsg = "开空单创建失败:" + OrderInfo.Symbol + "失败原因:一单" + err1.Error()
							notice.SendDingTalk("[开空单创建失败预警]" + orderMsg)
							return nil
						}
						quantity2, _ := decimal.NewFromString(inits.Config.Order.Warehouse_cnt[1])
						quantity2_safe := spot.QuantityCorrection(OrderInfo.Symbol, quantity2.Div(price2).String())
						_, quantity2_order, err2 := CreateLimitOrder(OrderInfo.Symbol, futures.SideTypeSell, price2, quantity2_safe, newClientOrderID+"A2")
						if err2 != nil {
							orderMsg = "开空单创建失败:" + OrderInfo.Symbol + "失败原因:二单" + err2.Error()
							notice.SendDingTalk("[开空单失败预警]" + orderMsg)
						}
						if quantity1_order.GreaterThan(decimal.Zero) || quantity2_order.GreaterThan(decimal.Zero) {
							TakeProfitPrice := GetTakeProfitPrice(futures.SideTypeSell, OrderInfo.K1_close)
							//如果止盈价不到中轨，就用中轨价
							if TakeProfitPrice.GreaterThan(OrderInfo.Ma) {
								TakeProfitPrice = OrderInfo.Ma
							}
							orderStopAllQuantity := quantity1_safe.Add(quantity2_safe)
							logger.Warnf("做空数量:,orderStopAllQuantity=%v,quantity1_safe=%s,quantity2_safe=%v", orderStopAllQuantity, quantity1_safe, quantity2_safe)
							a2stopprice := stopPrice.Mul(decimal.NewFromFloat(0.9995))
							orderInfo := &datastruct.OrderLimitInfo{
								Symbol:               OrderInfo.Symbol,
								ClientOrderID:        newClientOrderID,
								OrderStopSideType:    string(futures.SideTypeBuy),
								OrderStopPrice:       stopPrice,
								OrderStopQuantity:    quantity1,
								OrderStopAllQuantity: orderStopAllQuantity,
								OrderClientOrderID:   newClientOrderID + "B1",
								A2_stop_price:        a2stopprice,

								OrderLimitSideType:      string(futures.SideTypeSell),
								TakeProfitSideType:      string(futures.SideTypeBuy),
								TakeProfitPrice:         TakeProfitPrice,
								TakeProfitQuantity:      quantity2,
								TakeProfitClientOrderID: newClientOrderID + "C1",
							}
							//把orderInfo保存到全局变量中
							inits.SafeOrderInfo1.SetValue(OrderInfo.Symbol+newClientOrderID+"A1", orderInfo)
							//设置A1成交后对应的订单ID
							inits.SafeOrderInfo3.SetOrderClientOrderID(OrderInfo.Symbol, newClientOrderID+"A1")
							orderMsg = "开空挂单已提交，待成交:ClientOrderID:" + newClientOrderID + "A1," + newClientOrderID + "A2"
						}
					}
					orderMsg = orderMsg + "\n" + rateStr
				}
			}
		} else {
			orderMsg = "未开启创建订单功能，只是预警"
		}
		//布林线差值比=（上轨-下轨）/ 中规
		ftime := utils.FormatTime()
		up_dn := OrderInfo.Up.Sub(OrderInfo.Dn).Div(OrderInfo.Ma).Mul(mulvalue)
		//市场情绪
		logger.Infof("msg=%s,time=%s,symbol=%s,up=%s,ma=%s,dn=%s", "合约，快乐预警:上轨反包做空有信号", ftime, OrderInfo.Symbol, OrderInfo.Up.String(), OrderInfo.Ma.String(), OrderInfo.Dn.String())
		logger.Infof("symbol=%s,k1_close <= k2_low(%s,%s),k1_open >= k1_close(%s,%s),k2_open <= k2_close(%s,%s) ",
			OrderInfo.Symbol, OrderInfo.K1_close.String(), OrderInfo.K2_low.String(), OrderInfo.K1_open.String(), OrderInfo.K1_close.String(), OrderInfo.K2_open.String(), OrderInfo.K2_close.String())
		msg := "策略名称：反包策略预警" + macdMsg + " \n 交易对：" + OrderInfo.Symbol + "\n方向：空 \n 预警时间：" + ftime + "\n预警价格：" + OrderInfo.K1_close.String() + " \n上一根K线的最高价：" + OrderInfo.K2_high.String() + " \n 这一根k线的收盘价格：" + OrderInfo.K1_close.String() + "\n 布林线差值比：" + up_dn.String() + "\n" + orderMsg
		notice.SendDingTalk(msg)
		//写入数据库
		sqlite.WarngingInsert(order_id, OrderInfo.Symbol, 2, OrderInfo.K1_close, OrderInfo.K2_low, OrderInfo.K1_close, up_dn, ftime)
	}
	return nil
}

// 创建合约限价单
func CreateLimitOrder(symbol string, SideType futures.SideType, bestPrice, quantity decimal.Decimal, clientOrderID string) (*futures.CreateOrderResponse, decimal.Decimal, error) {
	acount := inits.Config.Account
	apiKey := acount.ApiKey
	secretKey := acount.SecretKey
	timeInForce := futures.TimeInForceTypeGTC
	newClientOrderID := clientOrderID
	workingType := futures.WorkingTypeContractPrice
	newOrderRespType := futures.NewOrderRespTypeRESULT
	orderType := futures.OrderTypeLimit
	positionSide := futures.PositionSideTypeBoth

	res, err := client.GetFutureClient(apiKey, secretKey).NewCreateOrderService().
		Symbol(symbol).
		Side(SideType).           //买（开多），卖（开空）
		Type(orderType).          //limit,限价单
		TimeInForce(timeInForce). //GTC
		Quantity(quantity.String()).
		Price(bestPrice.String()).
		NewClientOrderID(newClientOrderID).
		Do(newContext())
	if err != nil {
		logger.Infof("msg=CreateLimitOrder fail,symbol=%s,sideType=%s,orderType=%s,positionSide=%s,timeInForce=%s,quantity=%s,price=%s,newClientOrderID=%s,workingType=%s,newOrderRespType=%s,err=%v",
			symbol, SideType, orderType, positionSide, timeInForce, quantity.String(), bestPrice.String(), newClientOrderID, workingType, newOrderRespType, err)
		return res, quantity, err
	} else {
		logger.Infof("msg=CreateLimitOrder success,symbol=%s,orderType=%s,positionSide=%s,timeInForce=%s,quantity=%s,price=%s,newClientOrderID=%s,workingType=%s,newOrderRespType=%s",
			symbol, orderType, positionSide, timeInForce, quantity.String(), bestPrice.String(), newClientOrderID, workingType, newOrderRespType)
		//把res结果输出到日志文件中
		jsonBytes, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			logger.Errorf("Failed to marshal CreateOrderResponse to JSON: %v", err)
		} else {
			logger.Infof("msg=%s", string(jsonBytes))
		}
		return res, quantity, nil
	}
}

// 创建合约止损限价单
func CreateStopLimitOrder(symbol string, SideType futures.SideType, price, quantity decimal.Decimal, clientOrderID string) (*futures.CreateOrderResponse, error) {
	//计算价格，如果是做空，则价格为k1_close*0.9998，如果多，则价格为k1_close*1.00002

	logger.Warnf("止损单参数,clientOrderID=%v,symbol=%s,SideType=%v,OrderStopPrice=%v,quantity=%v", clientOrderID, symbol, SideType, price, quantity)
	price = spot.PriceCorrection(symbol, price.String())
	priceString := price.String()

	acount := inits.Config.Account
	apiKey := acount.ApiKey
	secretKey := acount.SecretKey
	timeInForce := futures.TimeInForceTypeGTC
	newClientOrderID := clientOrderID
	workingType := futures.WorkingTypeContractPrice
	newOrderRespType := futures.NewOrderRespTypeRESULT
	orderType := futures.OrderTypeStop
	positionSide := futures.PositionSideTypeBoth

	res, err := client.GetFutureClient(apiKey, secretKey).NewCreateOrderService().
		Symbol(symbol).
		Side(SideType).           //买（开多），卖（开空）
		Type(orderType).          //limit,限价单
		TimeInForce(timeInForce). //GTC
		Quantity(quantity.String()).
		Price(priceString).
		StopPrice(priceString).
		ReduceOnly(true).
		NewClientOrderID(newClientOrderID).
		Do(newContext())
	if err != nil {
		logger.Infof("msg=Create stop limit Order fail,symbol=%s,sideType=%s,orderType=%s,positionSide=%s,timeInForce=%s,quantity=%s,price=%s,stopPrice=%s,newClientOrderID=%s,workingType=%s,newOrderRespType=%s,err=%v",
			symbol, SideType, orderType, positionSide, timeInForce, quantity, priceString, priceString, newClientOrderID, workingType, newOrderRespType, err)
		//创建市价止损单
		market, marketError := CreateStopMarkerOrder(symbol, futures.SideType(SideType), quantity, newClientOrderID)
		if marketError != nil {
			logger.Warnf("logid=%s,msg=止损单市价创建失败,symbol=%s,err=%v", newClientOrderID, symbol, marketError.Error())
		} else {
			logger.Warnf("logid=%s,msg=止损单市价创建成功,symbol=%s,res=%v", newClientOrderID, symbol, market)
		}
		return res, err
	} else {
		logger.Infof("msg=Create stop limit Order success,symbol=%s,sideType=%s,orderType=%s,positionSide=%s,timeInForce=%s,quantity=%s,price=%s,stopPrice=%s,newClientOrderID=%s,workingType=%s,newOrderRespType=%s",
			symbol, SideType, orderType, positionSide, timeInForce, quantity, priceString, priceString, newClientOrderID, workingType, newOrderRespType)
		//把res结果输出到日志文件中
		jsonBytes, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			logger.Errorf("Failed to marshal CreateOrderResponse to JSON: %v", err)
		} else {
			logger.Infof("msg=%s", string(jsonBytes))
		}
		return res, nil
	}
}

func CreateStopMarkerOrder(symbol string, SideType futures.SideType, quantity decimal.Decimal, clientOrderID string) (*futures.CreateOrderResponse, error) {
	//市价单止损，不需要价格
	acount := inits.Config.Account
	apiKey := acount.ApiKey
	secretKey := acount.SecretKey
	newClientOrderID := clientOrderID
	workingType := futures.WorkingTypeContractPrice
	newOrderRespType := futures.NewOrderRespTypeRESULT
	orderType := futures.OrderTypeMarket
	positionSide := futures.PositionSideTypeBoth
	res, err := client.GetFutureClient(apiKey, secretKey).NewCreateOrderService().
		Symbol(symbol).
		Side(SideType).  //买（开多），卖（开空）
		Type(orderType). //limit,限价止损
		Quantity(quantity.String()).
		ReduceOnly(true).
		NewClientOrderID(newClientOrderID).
		Do(newContext())
	if err != nil {
		logger.Infof("msg=挂单失败，市价单失败,symbol=%s,sideType=%s,orderType=%s,positionSide=%s,quantity=%s,newClientOrderID=%s,workingType=%s,newOrderRespType=%s,err=%v",
			symbol, SideType, orderType, positionSide, quantity, newClientOrderID, workingType, newOrderRespType, err)
		return res, err
	} else {
		logger.Infof("msg=挂单失败，市价单成功,symbol=%s,sideType=%s,orderType=%s,positionSide=%s,quantity=%s,newClientOrderID=%s,workingType=%s,newOrderRespType=%s",
			symbol, SideType, orderType, positionSide, quantity, newClientOrderID, workingType, newOrderRespType)
		cance_err1 := client.GetFutureClient(apiKey, secretKey).NewCancelAllOpenOrdersService().Symbol(symbol).Do(newContext())
		if cance_err1 != nil {
			logger.Infof("市价止损单-取消委托单失败,symbol=%s,err=%v", symbol, cance_err1)
		} else {
			logger.Infof("市价止损单-取消所有委托单成功,symbol=%s", symbol)
		}
		return res, nil
	}
}

// 创建超级止损单
func CreateSuperMarkerOrder(symbol, newClientOrderID string) (*futures.CreateOrderResponse, error) {
	//市价单止损，不需要价格
	acount := inits.Config.Account
	apiKey := acount.ApiKey
	secretKey := acount.SecretKey
	orderType := futures.OrderTypeMarket
	positionInfo, o_err := client.GetFutureClient(apiKey, secretKey).NewGetPositionRiskService().Symbol(symbol).Do(context.Background())
	var SideType string
	var quantity decimal.Decimal
	if o_err != nil {
		logger.Errorf("msg=%s||symbol=%s||api_key=%s||secret_key=%s||err=%s",
			"super oredr get fail", symbol, apiKey, secretKey, o_err.Error())
	} else {
		for _, v := range positionInfo {
			positionAmt, _ := decimal.NewFromString(v.PositionAmt)
			if positionAmt.GreaterThan(define.Decimal0) {
				SideType = "SELL"
			} else {
				SideType = "BUY"
			}
			quantity = positionAmt.Abs()
		}
	}
	res, err := client.GetFutureClient(apiKey, secretKey).NewCreateOrderService().
		Symbol(symbol).
		Side(futures.SideType(SideType)). //买（开多），卖（开空）
		Type(orderType).                  //market 市价单
		Quantity(quantity.String()).
		ReduceOnly(true).
		NewClientOrderID(newClientOrderID).
		Do(newContext())
	if err != nil {
		logger.Infof("msg=超级止损单失败,symbol=%s,sideType=%s,orderType=%s,positionSide=%s,quantity=%s,newClientOrderID=%s,workingType=%s,newOrderRespType=%s,err=%v",
			symbol, SideType, orderType, SideType, quantity.String(), newClientOrderID, err)
		return res, err
	} else {
		logger.Infof("msg=超级止损单成功,symbol=%s,sideType=%s,orderType=%s,positionSide=%s,quantity=%s,newClientOrderID=%s",
			symbol, SideType, orderType, SideType, quantity.String(), newClientOrderID)
		cance_err1 := client.GetFutureClient(apiKey, secretKey).NewCancelAllOpenOrdersService().Symbol(symbol).Do(newContext())
		if cance_err1 != nil {
			logger.Infof("超级止损单-取消委托单失败,symbol=%s,err=%v", symbol, cance_err1)
		} else {
			logger.Infof("超级止损单-取消所有委托单成功,symbol=%s", symbol)
		}
		return res, nil
	}
}

// 创建合约止盈限价单
func CreateTakeProfitOrder(symbol string, SideType futures.SideType, price, quantity decimal.Decimal, clientOrderID string) (*futures.CreateOrderResponse, error) {

	price = spot.PriceCorrection(symbol, price.String())
	priceString := price.String()

	acount := inits.Config.Account
	apiKey := acount.ApiKey
	secretKey := acount.SecretKey
	timeInForce := futures.TimeInForceTypeGTC
	workingType := futures.WorkingTypeContractPrice
	newOrderRespType := futures.NewOrderRespTypeRESULT
	orderType := futures.OrderTypeTakeProfit
	positionSide := futures.PositionSideTypeBoth

	res, err := client.GetFutureClient(apiKey, secretKey).NewCreateOrderService().
		Symbol(symbol).
		Side(SideType).           //买（开多），卖（开空）
		Type(orderType).          //limit,限价单
		TimeInForce(timeInForce). //GTC
		Quantity(quantity.String()).
		Price(priceString).
		StopPrice(priceString).
		NewClientOrderID(clientOrderID).
		Do(newContext())
	if err != nil {
		logger.Infof("msg=Create takeprofit limit Order fail,symbol=%s,sideType=%s,orderType=%s,positionSide=%s,timeInForce=%s,quantity=%s,price=%s,stopPrice=%s,newClientOrderID=%s,workingType=%s,newOrderRespType=%s,err=%v",
			symbol, SideType, orderType, positionSide, timeInForce, quantity, priceString, priceString, clientOrderID, workingType, newOrderRespType, err)
		//创建市价止盈单
		market1, marketError1 := CreateStopMarkerOrder(symbol, futures.SideType(SideType), quantity, clientOrderID)
		if marketError1 != nil {
			logger.Warnf("logid=%s,msg=止盈单市价创建失败,symbol=%s,err=%v", clientOrderID, symbol, marketError1.Error())
		} else {
			logger.Warnf("logid=%s,msg=止盈单市价创建成功,symbol=%s,res=%v", clientOrderID, symbol, market1)
		}
		return res, err
	} else {
		//把止盈数据写入到共享变量中
		logger.Infof("msg=Create takeprofit limit Order success,symbol=%s,sideType=%s,orderType=%s,positionSide=%s,timeInForce=%s,quantity=%s,price=%s,stopPrice=%s,newClientOrderID=%s,workingType=%s,newOrderRespType=%s",
			symbol, SideType, orderType, positionSide, timeInForce, quantity, priceString, priceString, clientOrderID, workingType, newOrderRespType)
		//把res结果输出到日志文件中
		jsonBytes, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			logger.Errorf("Failed to marshal CreateOrderResponse to JSON: %v", err)
		} else {
			logger.Infof("msg=%s", string(jsonBytes))
		}
		return res, nil
	}

}

func CanceleOrder(symbol, origClientOrderId string) error {
	acount := inits.Config.Account
	apiKey := acount.ApiKey
	secretKey := acount.SecretKey
	res, err := client.GetFutureClient(apiKey, secretKey).NewCancelOrderService().Symbol(symbol).OrigClientOrderID(origClientOrderId).Do(newContext())
	if err != nil {
		logger.Infof("msg=Cancel Order fail,symbol=%s,origClientOrderId=%s,err=%v", symbol, origClientOrderId, err)
		return err
	} else {
		logger.Infof("msg=Cancel Order success,symbol=%s,origClientOrderId=%s", symbol, origClientOrderId)
		//把res结果输出到日志文件中
		jsonBytes, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			logger.Errorf("Failed to marshal CancelOrderResponse to JSON: %v", err)
		} else {
			logger.Infof("msg=%s", string(jsonBytes))
		}
		return nil
	}
}

type StockData struct {
	High  float64
	Low   float64
	Close float64
}

func GetKdjBySymbol(symbol string, klineEvents []datastruct.WsKlineEvent) ([]float64, []float64, []float64) {
	var k, d, j []float64
	StockDataSlice := []StockData{}
	// 计算有几位小数
	places, _ := GetPlaces(symbol)
	for _, e := range klineEvents {
		high, _ := utils.StrTofloat64(e.Kline.High)
		low, _ := utils.StrTofloat64(e.Kline.Low)
		close, _ := utils.StrTofloat64(e.Kline.Close)
		StockDataSlice = append(StockDataSlice, StockData{
			High:  high,
			Low:   low,
			Close: close,
		})
	}
	k, d, j = calcKDJ(StockDataSlice, 9, 3, 3, places)
	k_len := len(k)
	fmt.Println(symbol, k[k_len-1], d[k_len-1], j[k_len-1])
	return k, d, j
}

// 计算RSV值
func calcRSV(data []StockData, index int, n int) float64 {
	if index < n {
		// 数据不足以计算RSV
		return 0
	}

	lowPrices := data[index-n+1 : index+1]
	minLow := lowPrices[0].Low
	for _, d := range lowPrices {
		if d.Low < minLow {
			minLow = d.Low
		}
	}

	highPrices := data[index-n+1 : index+1]
	maxHigh := highPrices[0].High
	for _, d := range highPrices {
		if d.High > maxHigh {
			maxHigh = d.High
		}
	}

	if maxHigh == minLow {
		// 防止除零错误
		return 50
	}

	return (data[index].Close - minLow) / (maxHigh - minLow) * 100
}

// 计算KDJ指标
func calcKDJ(data []StockData, n, m1, m2, places int) ([]float64, []float64, []float64) {
	if len(data) < n {
		// 数据不足以计算KDJ
		return nil, nil, nil
	}

	k := make([]float64, len(data))
	d := make([]float64, len(data))
	j := make([]float64, len(data))

	// 初始化KDJ值（例如使用50作为初始值）
	for i := 0; i < n; i++ {
		k[i] = 50.0
		d[i] = 50.0
		j[i] = 50.0
	}

	// 计算KDJ值
	for i := n; i < len(data); i++ {
		rsv := calcRSV(data, i, n)
		if i == n {
			k[i] = rsv
		} else {
			k[i] = (float64(m1)*k[i-1] + float64(100-m1)*rsv) / 100
		}
		if i >= n+m1-1 {
			d[i] = (float64(m2)*d[i-1] + float64(100-m2)*k[i]) / 100
		}
		j[i] = 3*k[i] - 2*d[i]
	}

	return k, d, j
}

func GetMacdBySymbol(symbol string, klineEvents []datastruct.WsKlineEvent) ([]decimal.Decimal, []decimal.Decimal, []decimal.Decimal) {
	var closePrices, rdiff, rdea, rmacd []decimal.Decimal
	// 收盘价
	for _, v := range klineEvents {
		close, _ := decimal.NewFromString(v.Kline.Close)
		closePrices = append(closePrices, close)
	}
	//计算有几位小数
	var places int
	if symbol == inits.Config.Symbol.RootSymbol {
		places = 2
	} else {
		p, _ := GetPlaces(symbol)
		places = p
	}

	// 计算diff和dea值
	diff, dea := CalcMACD(closePrices, inits.Config.Kline.Macd_fast, inits.Config.Kline.Macd_slow, inits.Config.Kline.Macd_signal, places)

	closeLen := len(closePrices)
	for i := closeLen - 4; i < len(closePrices); i++ {
		rdiff = append(rdiff, utils.RoundDecimal45v2(diff[i], places))
		rdea = append(rdea, utils.RoundDecimal45v2(dea[i], places))
		rmacd = append(rmacd, utils.RoundDecimal45v2(diff[i].Sub(dea[i]), places))
	}
	return rdiff, rdea, rmacd

}

func CalcMACD(closePrices []decimal.Decimal, fastPeriod, slowPeriod, signalPeriod, places int) (diff []decimal.Decimal, dea []decimal.Decimal) {
	// 计算快速移动平均线(EMA)和慢速移动平均线(EMA)
	fastEMA := CalcEMA(closePrices, fastPeriod, places)
	slowEMA := CalcEMA(closePrices, slowPeriod, places)

	// diff值
	for i := range closePrices {
		diff = append(diff, fastEMA[i].Sub(slowEMA[i]))
	}
	// DEA
	dea = CalcEMA(diff, signalPeriod, places)

	return
}

// 计算EMA（指数移动平均线）
func CalcEMA(closePrices []decimal.Decimal, period, places int) []decimal.Decimal {
	result := make([]decimal.Decimal, len(closePrices))
	// 计算alpha
	two := decimal.NewFromInt(2)
	one := decimal.NewFromInt(1)
	periodDecimal := decimal.NewFromInt(int64(period))
	addOne := periodDecimal.Add(one)
	sumOne := periodDecimal.Sub(one)

	// 初始值
	result[0] = closePrices[0]
	for i := 1; i < len(closePrices); i++ {
		//EMA（12）=前一日EMA（12）×11/13+今日收盘价×2/13
		result[i] = result[i-1].Mul(sumOne).Div(addOne).Add(closePrices[i].Mul(two).Div(addOne))
	}
	return result
}

func GetPlaces(symbol string) (int, error) {
	//计算有几位小数
	tickSize, ok := inits.SpotPriceFilterInfo.GetTickSizeBySymbol(symbol)
	if !ok {
		inits.ErrorMsg(inits.ErrorBestPrice, errors.New("tickSize not found"))
		return 0, errors.New("tickSize not found")
	}
	//计算小数位数
	places := utils.CountDecimalPlaces(tickSize)
	return places, nil
}

// 止损价格
func GetStopPrice(sideType futures.SideType, k1_high, k2_high, k1_low, k2_low decimal.Decimal) decimal.Decimal {
	var price decimal.Decimal
	//做空
	if sideType == futures.SideTypeSell {
		if k1_high.GreaterThanOrEqual(k2_high) {
			price = k1_high
		} else {
			price = k2_high
		}
		//止损点上浮万五
		price = price.Mul(define.Decimal1.Add(decimal.NewFromFloat(inits.Config.Order.Stop_loss)))
	} else {
		if k1_low.GreaterThanOrEqual(k2_low) {
			price = k2_low
		} else {
			price = k1_low
		}
		//止损点下浮万五
		price = price.Mul(define.Decimal1.Sub(decimal.NewFromFloat(inits.Config.Order.Stop_loss)))
	}

	return price
}

// func GetStopPrice(symbol string, sideType futures.SideType, k1_close, k1_high decimal.Decimal) decimal.Decimal {
// 	var price decimal.Decimal
// 	//二仓价格
// 	twoPrice := spot.PriceCorrection(symbol, k1_high.Add(k1_close).Div(decimal.NewFromInt(2)).String())
// 	//收盘价上浮千1
// 	closePrice := k1_close.Mul(define.Decimal1.Sub(decimal.NewFromFloat(inits.Config.Order.Stop_loss)))
// 	//做空
// 	if sideType == futures.SideTypeSell {
// 		if twoPrice.GreaterThan(closePrice) {
// 			price = closePrice
// 		} else {
// 			price = twoPrice
// 		}
// 		//止损点上浮千1
// 	} else {
// 		if twoPrice.GreaterThan(closePrice) {
// 			price = twoPrice
// 		} else {
// 			price = closePrice
// 		}

// 	}
// 	return price
// }

// 止盈价格
func GetTakeProfitPrice(s1SideType futures.SideType, k1_close decimal.Decimal) (price decimal.Decimal) {
	if s1SideType == futures.SideTypeSell {
		//开空止盈价格 收盘价 * (1 - 止盈比例)
		price = k1_close.Mul(decimal.NewFromInt(1).Sub(decimal.NewFromFloat(inits.Config.Order.Stop_profit)))
	} else {
		//开多止盈价格 收盘价 * (1 + 止盈比例)
		price = k1_close.Mul(decimal.NewFromInt(1).Add(decimal.NewFromFloat(inits.Config.Order.Stop_profit)))
	}
	return price
}

func Bollma(Symbol, sideType string) (decimal.Decimal, bool) {
	//获取boll数据
	if bollinfo, ok := inits.SafeBollInfo.GetBoll(Symbol); !ok {
		return decimal.NewFromInt(0), false
	} else {
		if len(bollinfo) >= 2 {
			// buy 买多 向上中轨：k1_mb+(k1_mb-k2_mb)
			if sideType == define.SideTypeBuy {
				return bollinfo[1].Ma.Add(bollinfo[1].Ma.Sub(bollinfo[0].Ma)), true
			} else {
				//sell 买空 向下中轨：k1_mb-(k2_mb-k1_mb)
				return bollinfo[1].Ma.Sub(bollinfo[0].Ma.Sub(bollinfo[1].Ma)), true
			}
		} else {
			return decimal.NewFromInt(0), false
		}
	}
}
func CreateSuperStopOrder(symbol string) {
	time.Sleep(300 * time.Millisecond)
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
		start, _ := getCurrentMinuteStartAndEndAsInt64(100)
		// logger.Infof("msg=start and end,start=%d,end=%d", start, end)
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
	var orderBefore string
	if orderinfo, isExists := inits.SafeOrderInfo2.GetValue(symbol); isExists && inits.Config.Order.Open_mb {
		logger.Infof("msg=check orderinfo=%v", orderinfo)
		orderBefore = orderinfo[0].ClientOrderID[0 : len(orderinfo[0].ClientOrderID)-2]
		var superIsStopOrder bool
		// var qty decimal.Decimal
		superCid := orderBefore + "B0"
		for _, v := range orderinfo {
			logger.Infof("msg=show orderinfo=%v,mackState=%d", v, mackState)
			if v.ClientOrderID == superCid {
				superIsStopOrder = true
			}
			// qty = v.Quantity
		}
		//如果已存在就不创建超级止损单了
		if superIsStopOrder {
			return
		}
		// var stopPrice decimal.Decimal
		// bestprice, _ := inits.BestPriceInfo.GetBySymbol(symbol)
		// if futures.SideType(rootOrderInfo.TakeProfitSideType) == "BUY" {
		// 	stopPrice = decimal.NewFromFloat(bestprice.BestAskPrice)
		// } else {
		// 	stopPrice = decimal.NewFromFloat(bestprice.BestBidPrice)
		// }

		res3, err := CreateSuperMarkerOrder(symbol, superCid)
		if err == nil {
			// orderMsg = symbol + "超级止损单创建失败:" + fmt.Sprintf("%v", err.Error())
			// notice.SendDingTalk(orderMsg)
			// } else {
			//撤消所有非超级单且没有成交的
			// for _, v := range orderinfo {
			// 	logger.Infof("msg=show orderinfo=%v,mackState=%d", v, mackState)
			// 	if v.OrderStatus == "NEW" {

			// 		CanceleOrder(symbol, v.ClientOrderID)
			// 	}
			// }
			// orderMsg = symbol + "超级止损单创建创建成功:" + fmt.Sprintf("%v", res3)
			// notice.SendDingTalk(orderMsg)
			logger.Infof("msg=超级止损单创建成功,symbol=%s,res1:%v", symbol, res3)
			inits.SafeOrderInfo1.Delete(symbol + orderBefore + "A1")
			inits.SafeOrderInfo2.DeleteValue(symbol)
		}
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
