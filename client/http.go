package client

import (
	"net/http"
	_ "net/http/pprof" // pprof
	"os"
	"strconv"

	"go_code/myselfgo/handler"
	"go_code/myselfgo/inits"
	"go_code/myselfgo/utils"

	"github.com/open-binance/logger"
)

// StartHTTP starts http server
func StartHTTP() {
	defer func() {
		if r := recover(); r != nil {
			// 当发生 panic 时，记录错误日志并退出程序
			logger.Errorf("msg=recover from panic||err=%v", r)
			os.Exit(-1)
		}
	}()

	// 如果配置中启用了诊断功能
	// start pprof for diagnosis and prometheus for monitor
	if inits.Config.Diagnosis.Enable {

		// 添加诊断路由
		diagnosisRouters()

		// 组合 pprof 监听的地址和端口号
		address := utils.JoinStrWithSep(":", "0.0.0.0", strconv.Itoa(inits.Config.Diagnosis.Port))

		// 记录 pprof 启动的日志
		logger.Infof("msg=%s||address=%s", "pprof will start", address)

		// 在新的 goroutine 中启动 pprof 服务
		go func() {
			// 如果启动 pprof 服务失败，记录错误日志
			logger.Errorf("msg=failed to start pprof||err=%v", http.ListenAndServe(address, nil)) // will block when succeed
		}()
	}
}

// diagnosisRouters specifies routers for diagnosis
func diagnosisRouters() {
	// curl 'http://127.0.0.1:7000/diagnosis/version'
	http.HandleFunc("/diagnosis/version", handler.ShowVersion)
	// curl 'http://127.0.0.1:7000/diagnosis/best-price?symbol=BTCBUSD'
	// curl 'http://127.0.0.1:7000/diagnosis/best-price'
	http.HandleFunc("/diagnosis/best-price", handler.GetBestPrice)
	// curl 'http://127.0.0.1:7000/diagnosis/spot-price-filter?symbol=BTCBUSD'
	// curl 'http://127.0.0.1:7000/diagnosis/spot-price-filter'
	http.HandleFunc("/diagnosis/spot-price-filter", handler.GetSpotPriceFilterInfo)
	// curl 'http://127.0.0.1:7000/diagnosis/listen-key'
	http.HandleFunc("/diagnosis/listen-key", handler.GetListenKey)
	// curl 'http://127.0.0.1:7000/diagnosis/check-trading-symbol?symbol=BTCUSDT'
	http.HandleFunc("/diagnosis/check-trading-symbol", handler.CheckTradingSymbol)
	// curl 'http://127.0.0.1:7000/diagnosis/all-trading-symbols'
	http.HandleFunc("/diagnosis/all-trading-symbols", handler.GetAllTradingSymbols)
	// curl 'http://127.0.0.1:7000/diagnosis/check-wanted-symbol?symbol=BTCBUSD'
	http.HandleFunc("/diagnosis/check-wanted-symbol", handler.CheckWantedSymbols)
	// curl 'http://127.0.0.1:7000/diagnosis/wanted-symbols'
	http.HandleFunc("/diagnosis/wanted-symbols", handler.GetAllWantedSymbols)
	// curl 'http://127.0.0.1:7000/diagnosis/check-base-asset-white-list?base=BTC'
	http.HandleFunc("/diagnosis/check-base-asset-white-list", handler.CheckBaseAssetWhiteList)
	// curl 'http://127.0.0.1:7000/diagnosis/base-asset-white-list'
	http.HandleFunc("/diagnosis/base-asset-white-list", handler.GetAllBaseAssetWhiteList)
	// curl 'http://127.0.0.1:7000/diagnosis/check-symbol-black-list?symbol=BTCBUSD'
	http.HandleFunc("/diagnosis/check-symbol-black-list", handler.CheckSymbolBlackList)
	// curl 'http://127.0.0.1:7000/diagnosis/symbol-black-list'
	http.HandleFunc("/diagnosis/symbol-black-list", handler.GetAllSymbolBlackList)
	// curl 'http://127.0.0.1:7000/diagnosis/check-symbol-white-list?symbol=GBPBUSD'
	http.HandleFunc("/diagnosis/check-symbol-white-list", handler.CheckSymbolWhiteList)
	// curl 'http://127.0.0.1:7000/diagnosis/symbol-white-list'
	http.HandleFunc("/diagnosis/symbol-white-list", handler.GetAllSymbolWhiteList)
	// curl 'http://127.0.0.1:7000/diagnosis/symbol-pairs-by-base?base=BNB'
	// curl 'http://127.0.0.1:7000/diagnosis/symbol-pairs-by-base'
	http.HandleFunc("/diagnosis/symbol-pairs-by-base", handler.GetSymbolPairsByBase)
	// curl 'http://127.0.0.1:7000/diagnosis/symbol-pairs-by-symbol?symbol=BTCBUSD'
	// curl 'http://127.0.0.1:7000/diagnosis/symbol-pairs-by-symbol'
	http.HandleFunc("/diagnosis/symbol-pairs-by-symbol", handler.GetSymbolPairsBySymbol)
	// curl 'http://127.0.0.1:7000/diagnosis/free-charge-symbols'
	http.HandleFunc("/diagnosis/free-charge-symbols", handler.GetAllFreeChargeSymbols)
	// curl 'http://127.0.0.1:7000/diagnosis/auto-trade/done'
	http.HandleFunc("/diagnosis/auto-trade/done", handler.TradeFlagDone)
	// curl 'http://127.0.0.1:7000/diagnosis/used-weight-1m'
	http.HandleFunc("/diagnosis/used-weight-1m", handler.GetUsedWeight1m)
	// curl 'http://127.0.0.1:7000/diagnosis/order-count-10s'
	http.HandleFunc("/diagnosis/order-count-10s", handler.GetOrderCount10s)
	// curl 'http://127.0.0.1:7000/diagnosis/order-count-1d'
	http.HandleFunc("/diagnosis/order-count-1d", handler.GetOrderCount1d)
	// curl 'http://127.0.0.1:7000/diagnosis/host-and-delay'
	http.HandleFunc("/diagnosis/host-and-delay", handler.GetHostAndDelay)
	// curl 'http://127.0.0.1:7000/diagnosis/all-host-and-delay'
	http.HandleFunc("/diagnosis/all-host-and-delay", handler.GetAllHostAndDelay)
	// curl 'http://127.0.0.1:7000/diagnosis/rate-limit'
	http.HandleFunc("/diagnosis/rate-limit", handler.GetRateLimit)
	// curl 'http://127.0.0.1:7000/diagnosis/account/spot?asset=BNB'
	// curl 'http://127.0.0.1:7000/diagnosis/account/spot'
	http.HandleFunc("/diagnosis/account/spot", handler.GetSpotAccountInfo)
	// curl 'http://127.0.0.1:7000/diagnosis/account/funding?asset=BNB'
	// curl 'http://127.0.0.1:7000/diagnosis/account/funding'
	http.HandleFunc("/diagnosis/account/funding", handler.GetFundingAccountInfo)
	// curl 'http://127.0.0.1:7000/diagnosis/quote-qty-info'
	http.HandleFunc("/diagnosis/quote-qty-info", handler.GetQuoteQtyInfo)
	// curl 'http://127.0.0.1:7000/diagnosis/hold-coins'
	http.HandleFunc("/diagnosis/hold-coins", handler.GetHoldCoins)
	// curl 'http://127.0.0.1:7000/diagnosis/return-rate'
	http.HandleFunc("/diagnosis/return-rate", handler.GetReturnRate)
	// curl 'http://127.0.0.1:7000/diagnosis/return-rate-threshold'
	http.HandleFunc("/diagnosis/return-rate-threshold", handler.GetReturnRateThreshold)

}
