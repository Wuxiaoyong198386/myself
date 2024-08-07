package handler

import (
	"fmt"
	"net/http"

	json "github.com/json-iterator/go"

	"go_code/myselfgo/inits"
	"go_code/myselfgo/modules/datastruct"

	"github.com/open-binance/logger"
)

// ShowVersion shows app version
func ShowVersion(w http.ResponseWriter, r *http.Request) {
	versionInfo := inits.VersionMsg{
		AppName: inits.AppName,
		Version: inits.Version,
	}

	versionInfoJSON, err := json.Marshal(versionInfo)
	if err != nil {
		logger.Errorf("msg=failed to marshal version info||err=%s", err.Error())
	} else {
		logger.Infof("msg=succeed to get version||version=%s", string(versionInfoJSON))
	}

	SendResp(w, versionInfo)
}

// GetBestPrice gets best price cached in memory with one symbol
func GetBestPrice(w http.ResponseWriter, r *http.Request) {
	symbol := r.FormValue("symbol")
	if symbol != "" {
		bestPrice, existed := inits.BestPriceInfo.GetBySymbol(symbol)
		if !existed {
			logger.Infof("msg=get no best price||symbol=%s", symbol)
			SendResp(w, bestPrice)
			return
		}
		logger.Infof("msg=succeed to get best price||symbol=%s", symbol)
		SendResp(w, bestPrice)
	} else { // get all best price
		m := inits.BestPriceInfo.Get()
		logger.Infof("msg=succeed to get all best price||cnt=%d", len(m))
		SendResp(w, m)
	}
}

// GetSpotPriceFilterInfo gets price filter info for spot
func GetSpotPriceFilterInfo(w http.ResponseWriter, r *http.Request) {
	symbol := r.FormValue("symbol")
	if symbol != "" {
		priceFilterInfo, existed := inits.SpotPriceFilterInfo.GetBySymbol(symbol)
		if !existed {
			logger.Infof("msg=get no price filter info||symbol=%s", symbol)
		} else {
			logger.Infof("msg=succeed to get price filter info||symbol=%s", symbol)
		}
		SendResp(w, priceFilterInfo)
		return
	}

	// get all price filter info for spot
	m := inits.SpotPriceFilterInfo.Get()
	logger.Infof("msg=succeed to get all price filter info||cnt=%d", len(m))
	SendResp(w, m)
}

// GetListenKey gets listen key
func GetListenKey(w http.ResponseWriter, r *http.Request) {
	listenKey := inits.ListenKey.Get()
	logger.Infof("msg=succeed to get listen key||listen_key=%s", listenKey)
	SendResp(w, listenKey)
}

// CheckTradingSymbol checks the trading symbol
func CheckTradingSymbol(w http.ResponseWriter, r *http.Request) {
	symbol := r.FormValue("symbol")
	info, ok := datastruct.AllTradingSymbols.GetBySymbol(symbol)
	logger.Infof("msg=succeed to check trading symbol||symbol=%s", symbol)
	if !ok {
		logger.Infof("msg=miss the trading symbol||symbol=%s", symbol)
		SendResp(w, "miss")
		return
	}
	logger.Infof("msg=hit the trading symbol||symbol=%s", symbol)
	SendResp(w, info)
}

// GetAllTradingSymbols gets all trading symbols
func GetAllTradingSymbols(w http.ResponseWriter, r *http.Request) {
	m := datastruct.AllTradingSymbols.Get()
	logger.Infof("msg=succeed to get all trading symbols||cnt=%d", len(m))
	SendResp(w, m)
}

// CheckWantedSymbols checks whether the symbol is in the map or not
func CheckWantedSymbols(w http.ResponseWriter, r *http.Request) {
	symbol := r.FormValue("symbol")
	existed := inits.WantedSymbols.IsExisted(symbol)
	msg := MsgUnknown
	if !existed {
		msg = MsgMiss
		logger.Infof("msg=miss the wanted symbols||symbol=%s", symbol)
	} else {
		msg = MsgHit
		logger.Infof("msg=hit the wanted symbols||symbol=%s", symbol)
	}
	SendResp(w, msg)
}

// GetAllWantedSymbols gets all wanted symbols
func GetAllWantedSymbols(w http.ResponseWriter, r *http.Request) {
	m := inits.WantedSymbols.Get()
	logger.Infof("msg=succeed to get all wanted symbols||cnt=%d", len(m))
	SendResp(w, m)
}

// CheckBaseAssetWhiteList checks whether the base asset is in the map or not
func CheckBaseAssetWhiteList(w http.ResponseWriter, r *http.Request) {
	base := r.FormValue("base")
	existed := inits.BaseAssetWhiteList.IsExisted(base)
	msg := MsgUnknown
	if !existed {
		msg = MsgMiss
		logger.Infof("msg=miss base asset white list||base=%s", base)
	} else {
		msg = MsgHit
		logger.Infof("msg=hit base asset white list||base=%s", base)
	}
	SendResp(w, msg)
}

// GetAllBaseAssetWhiteList gets all base assets in the white list
func GetAllBaseAssetWhiteList(w http.ResponseWriter, r *http.Request) {
	m := inits.BaseAssetWhiteList.Get()
	logger.Infof("msg=succeed to get all base assets in the white list||cnt=%d", len(m))
	SendResp(w, m)
}

// CheckSymbolBlackList checks whether the symbol is in black list or not
func CheckSymbolBlackList(w http.ResponseWriter, r *http.Request) {
	base := r.FormValue("symbol")
	existed := inits.SymbolBlackList.IsExisted(base)
	msg := MsgUnknown
	if !existed {
		msg = MsgMiss
		logger.Infof("msg=miss symbol black list||symbol=%s", base)
	} else {
		msg = MsgHit
		logger.Infof("msg=hit symbol black list||symbol=%s", base)
	}
	SendResp(w, msg)
}

// GetAllSymbolBlackList gets all symbols in the black list
func GetAllSymbolBlackList(w http.ResponseWriter, r *http.Request) {
	m := inits.SymbolBlackList.Get()
	logger.Infof("msg=succeed to get all symbols in the black list||cnt=%d", len(m))
	SendResp(w, m)
}

// CheckSymbolWhiteList checks whether the symbol is in white list or not
func CheckSymbolWhiteList(w http.ResponseWriter, r *http.Request) {
	base := r.FormValue("symbol")
	existed := inits.SymbolWhiteList.IsExisted(base)
	msg := MsgUnknown
	if !existed {
		msg = MsgMiss
		logger.Infof("msg=miss symbol white list||symbol=%s", base)
	} else {
		msg = MsgHit
		logger.Infof("msg=hit symbol white list||symbol=%s", base)
	}
	SendResp(w, msg)
}

// GetAllSymbolWhiteList gets all symbols in the white list
func GetAllSymbolWhiteList(w http.ResponseWriter, r *http.Request) {
	m := inits.SymbolWhiteList.Get()
	logger.Infof("msg=succeed to get all symbols in the white list||cnt=%d", len(m))
	SendResp(w, m)
}

// GetSymbolPairsByBase gets symbol pairs cached in memory by base
func GetSymbolPairsByBase(w http.ResponseWriter, r *http.Request) {
	base := r.FormValue("base")
	if base != "" {
		symbolPairs, existed := inits.Base2SymbolPairs.GetByKey(base)
		if !existed {
			logger.Infof("msg=get no symbol pairs||base_asset=%s", base)
		} else {
			logger.Infof("msg=succeed to get symbol pairs by base||base_asset=%s||cnt=%d", base, len(symbolPairs))
		}
		SendResp(w, symbolPairs)
		return
	}

	// get all symbol pairs for all base asset
	m := inits.Base2SymbolPairs.Get()
	cnt := 0
	for _, symbolPairs := range m {
		cnt += len(symbolPairs)
	}
	logger.Infof("msg=succeed to get all symbol pairs||cnt=%d", cnt)
	SendResp(w, m)
}

// GetSymbolPairsBySymbol gets symbol pairs cached in memory by symbol
func GetSymbolPairsBySymbol(w http.ResponseWriter, r *http.Request) {
	symbol := r.FormValue("symbol")
	if symbol != "" {
		symbolPairs, existed := inits.Symbol2SymbolPairs.GetByKey(symbol)
		if !existed {
			logger.Infof("msg=get no symbol pairs||symbol=%s", symbol)
		} else {
			logger.Infof("msg=succeed to get symbol pairs by symbol||symbol=%s||cnt=%d", symbol, len(symbolPairs))
		}
		SendResp(w, symbolPairs)
		return
	}

	// get all symbol pairs for all base asset
	m := inits.Symbol2SymbolPairs.Get()
	cnt := 0
	for _, symbolPairs := range m {
		cnt += len(symbolPairs)
	}
	logger.Infof("msg=succeed to get all symbol pairs||cnt=%d", cnt)
	SendResp(w, m)
}

// GetAllFreeChargeSymbols gets symbols with free charge
func GetAllFreeChargeSymbols(w http.ResponseWriter, r *http.Request) {
	m := inits.FreeChargeSymbols.Get()
	logger.Infof("msg=succeed to get symbols with free charge||cnt=%d", len(m))
	SendResp(w, m)
}

// TradeFlagDone marks algo.TradeFlag as done
func TradeFlagDone(w http.ResponseWriter, r *http.Request) {
	inits.FlagTrade.Done()
	logger.Infof("msg=succeed to mark auto trde as done")
	SendResp(w, "succeed")
}

// GetUsedWeight1m gets the 1m used weight
func GetUsedWeight1m(w http.ResponseWriter, r *http.Request) {
	usedWeight1m := inits.UsedWeight1m.Get()
	logger.Infof("msg=succeed to get 1m used weight")
	SendResp(w, usedWeight1m)
}

// GetOrderCount10s gets the 10s order count
func GetOrderCount10s(w http.ResponseWriter, r *http.Request) {
	usedWeight1m := inits.OrderCount10s.Get()
	logger.Infof("msg=succeed to get 10s order count")
	SendResp(w, usedWeight1m)
}

// GetOrderCount1d gets the 1d order count
func GetOrderCount1d(w http.ResponseWriter, r *http.Request) {
	usedWeight1m := inits.OrderCount1d.Get()
	logger.Infof("msg=succeed to get 1d order count")
	SendResp(w, usedWeight1m)
}

// GetHostAndDelay gets the host and network delay
func GetHostAndDelay(w http.ResponseWriter, r *http.Request) {
	host, delay := inits.NetworkDelayMap.GetHostAndDelay()
	logger.Infof("msg=succeed to get host and network delay||host=%s||delay=%.1fms", host, delay)
	SendResp(w, fmt.Sprintf("host: %s, delay: %.1f ms", host, delay))
}

// GetAllHostAndDelay gets all the host and network delay
func GetAllHostAndDelay(w http.ResponseWriter, r *http.Request) {
	m := inits.NetworkDelayMap.GetAllHostAndDelay()
	logger.Infof("msg=succeed to get all host and network delay")
	SendResp(w, m)
}

// GetRateLimit gets rate limit info
func GetRateLimit(w http.ResponseWriter, r *http.Request) {
	rateLimit := inits.RateLimit.Get()
	rateLimitJSON, _ := json.Marshal(rateLimit)
	logger.Infof("msg=succeed to get rate limit info||rate_limit=%s", string(rateLimitJSON))
	SendResp(w, rateLimit)
}

// GetSpotAccountInfo gets spot account info price cached in memory
func GetSpotAccountInfo(w http.ResponseWriter, r *http.Request) {
	asset := r.FormValue("asset")
	if asset != "" {
		assetInfo, existed := inits.SpotAsset2AssetInfo.GetByAsset(asset)
		if !existed {
			logger.Infof("msg=get no spot account info||asset=%s", asset)
			SendResp(w, assetInfo)
			return
		}
		logger.Infof("msg=succeed to get spot account info||asset=%s", asset)
		SendResp(w, assetInfo)
	} else { // get all asset info
		m := inits.SpotAsset2AssetInfo.Get()
		mJSON, _ := json.Marshal(m)
		logger.Infof("msg=succeed to get all spot account info||cnt=%d||m=%s", len(m), string(mJSON))
		SendResp(w, m)
	}
}

// GetFundingAccountInfo gets funding account info price cached in memory
func GetFundingAccountInfo(w http.ResponseWriter, r *http.Request) {
	asset := r.FormValue("asset")
	if asset != "" {
		assetInfo, existed := inits.FundingAsset2AssetInfo.GetByAsset(asset)
		if !existed {
			logger.Infof("msg=get no funding account info||asset=%s", asset)
			SendResp(w, assetInfo)
			return
		}
		logger.Infof("msg=succeed to get funding account info||asset=%s", asset)
		SendResp(w, assetInfo)
	} else { // get all asset info
		m := inits.FundingAsset2AssetInfo.Get()
		logger.Infof("msg=succeed to get all funding account info||cnt=%d", len(m))
		SendResp(w, m)
	}
}

// GetQuoteQtyInfo gets quote quantity info
func GetQuoteQtyInfo(w http.ResponseWriter, r *http.Request) {
	info := inits.QuoteValueInfo.Get()
	commission := info.Commission
	commissionJSON, _ := json.Marshal(commission)
	logger.Infof("msg=succeed to get quote quantity info||plan=%v||cumulative=%v||delta=%v||commission=%s",
		info.PlanQuoteValue, info.CumulativeQuoteValue, info.DeltaQuoteValue, string(commissionJSON))
	SendResp(w, info)
}

// GetHoldCoins gets hold coins
func GetHoldCoins(w http.ResponseWriter, r *http.Request) {
	coins := inits.HoldCoins.Get()
	coinsJSON, _ := json.Marshal(coins)
	logger.Infof("msg=succeed to get hold coins||cnt=%d||coins=%s", len(coins), string(coinsJSON))
	SendResp(w, coins)
}

// GetReturnRate gets return rate
func GetReturnRate(w http.ResponseWriter, r *http.Request) {
	returnRateInfo := inits.CumulativeReturnRate.GetAll()
	returnRateInfoJSON, _ := json.Marshal(returnRateInfo)
	logger.Infof("msg=succeed to get return rate info||info=%s", string(returnRateInfoJSON))
	SendResp(w, returnRateInfo)
}

// GetReturnRateThreshold gets threshold of return rate
func GetReturnRateThreshold(w http.ResponseWriter, r *http.Request) {
	returnRateInfo := inits.ReturnRate.Get()
	returnRateInfoJSON, _ := json.Marshal(returnRateInfo)
	logger.Infof("msg=succeed to get threshold of return rate||info=%s", string(returnRateInfoJSON))
	SendResp(w, returnRateInfo)
}
