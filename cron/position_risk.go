package cron

import (
	"context"
	"go_code/myselfgo/client"
	"go_code/myselfgo/define"
	"go_code/myselfgo/inits"
	"go_code/myselfgo/notice"
	"sync"
	"time"

	"github.com/open-binance/logger"
	"github.com/shopspring/decimal"
)

func syncPositionRisk(interval int, wg *sync.WaitGroup) {
	defer wg.Done()
	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		for range ticker.C {
			for _, symbol := range inits.Config.Symbol.SymbolWhiteList {
				//如果有订单就查仓位
				_, ok := inits.SafeOrderInfo3.GetOrderClientOrderID(symbol)
				if ok {
					go SyncPositionRiskOnce(symbol)
				}
			}
		}
	}()
}

func SyncPositionRiskOnce(symbol string) {
	//获取接口需要用到的签名数据
	acount := inits.Config.Account
	apiKey := acount.ApiKey
	secretKey := acount.SecretKey

	//WithTimeout是Go标准库context中的函数，它接受一个Context和一个超时时间作为参数，返回一个子Context和一个取消函数CancelFunc
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(define.TimeoutBinanceAPI)*time.Second)
	//调用取消函数CancelFunc
	defer cancel()
	//获取数据
	prMsg := "[持仓预警]\n"
	var side string
	positionInfo, err := client.GetFutureClient(apiKey, secretKey).NewGetPositionRiskService().Symbol(symbol).Do(ctx)
	if err != nil {
		logger.Errorf("msg=%s||symbol=%s||api_key=%s||secret_key=%s||err=%s",
			"Failed to NewGetPositionRiskService", symbol, apiKey, secretKey, err.Error())
	} else {
		for _, v := range positionInfo {
			logger.Infof("symbol=%spositionInfo=%s", symbol, v)
			positionAmt, _ := decimal.NewFromString(v.PositionAmt)
			if positionAmt.GreaterThan(define.Decimal0) {
				side = "多"
			} else {
				side = "空"
			}
			prMsg = prMsg + "交易对: " + v.Symbol + " 方向：" + side + " 当前仓位: " + positionAmt.Abs().String() + " 持仓均价: " + v.EntryPrice + " 盈亏: " + v.UnRealizedProfit + "\n"
		}
		notice.SendDingTalk(prMsg)
	}

}

// {
// 	"entryPrice": "0.00000",  // 开仓均价
// 	"breakEvenPrice": "0.0",  // 盈亏平衡价
// 	"marginType": "isolated", // 逐仓模式或全仓模式
// 	"isAutoAddMargin": "false",
// 	"isolatedMargin": "0.00000000", // 逐仓保证金
// 	"leverage": "10", // 当前杠杆倍数
// 	"liquidationPrice": "0", // 参考强平价格
// 	"markPrice": "6679.50671178",   // 当前标记价格
// 	"maxNotionalValue": "20000000", // 当前杠杆倍数允许的名义价值上限
// 	"positionAmt": "0.000", // 头寸数量，符号代表多空方向, 正数为多，负数为空
// 	"notional": "0",
// 	"isolatedWallet": "0",
// 	"symbol": "BTCUSDT", // 交易对
// 	"unRealizedProfit": "0.00000000", // 持仓未实现盈亏
// 	"positionSide": "BOTH", // 持仓方向
// 	"updateTime": 1625474304765   // 更新时间
// }
