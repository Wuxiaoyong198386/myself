package inits

import (
	"go_code/myselfgo/define"
	"go_code/myselfgo/modules/datastruct"
	"go_code/myselfgo/utils"
	"sync"

	"github.com/adshao/go-binance/v2"
	binance_connector "github.com/binance/binance-connector-go"
	json "github.com/json-iterator/go"
	"github.com/shopspring/decimal"
)

var (
	emptyPriceFilterInfoMap = make(map[string]datastruct.SpotFilterInfo)
	emptyStrBoolMap         = make(map[string]bool)
	emptySymbolPairMap      = make(map[string][]datastruct.SymbolPair)
	MainPid                 int
	emptySafeOrderInfo1     = make(map[string]datastruct.OrderLimitInfo)
	WebsocketAPIClient      chan *binance_connector.WebsocketAPIClient
	OrderTjDefault          = decimal.NewFromInt(0)
)

// VersionMsg specifies the info of version
type VersionMsg struct {
	AppName     string `json:"app_name"`      // app name
	GitCommitID string `json:"git_commit_id"` // git commit id
	Version     string `json:"version"`       // version
}

// 定义全局变量
var (
	SafeOrderTjInfo = &datastruct.OrderTongJi{L: new(sync.RWMutex), MoneySum: OrderTjDefault, OrderFail: OrderTjDefault, OrderSuccess: OrderTjDefault, DorderFail: OrderTjDefault, DorderSuccess: OrderTjDefault, KorderFail: OrderTjDefault, KorderSuccess: OrderTjDefault}
	SafeMacdInfo    = &datastruct.KlineMacdMap{L: new(sync.RWMutex), M: make(map[string]datastruct.KlineMacd)}
	SafeBollInfo    = &datastruct.KlineBollMap{L: new(sync.RWMutex), Boll: make(map[string][]datastruct.BollInfo)}
	SafeOrderInfo   = &datastruct.SafeOrderInfo{L: new(sync.RWMutex), M: make(map[string]datastruct.OrderSymbols)}
	SafeOrderInfo2  = &datastruct.SafeOrderInfo2{L: new(sync.RWMutex), OrderInfo: make(map[string][]datastruct.OrderInfo2)}
	SafeOrderInfo3  = &datastruct.SafeOrderInfo3{L: new(sync.RWMutex), OrderInfoCid: make(map[string]string)}

	SafeWsPool            = &datastruct.AllState{L: new(sync.RWMutex), M: make(map[string]int)}
	SafeOrderS1status     = &datastruct.AllState{L: new(sync.RWMutex), M: make(map[string]int)}
	SafeWsOrderUpdateInfo = &datastruct.SafeWsOrderUpdate{L: new(sync.RWMutex), M: make(map[string]binance.WsOrderUpdate)}
	BestPriceInfo         = &datastruct.SafeBestPriceMap{L: new(sync.RWMutex), M: make(map[string]datastruct.BestPriceInfo)}
	KlineInfo             = &datastruct.SafeKlineMap{L: new(sync.RWMutex), M: make(map[string][]datastruct.WsKlineEvent)}
	WsService             = &datastruct.WebsocketClientManager{L: &sync.RWMutex{}, Client: WebsocketAPIClient, Pool: &datastruct.WebsocketClientManager{}}
	// SpotPriceFilterInfo stores price filter info
	SpotPriceFilterInfo = &datastruct.SafePriceFilterInfoMap{L: new(sync.RWMutex), Timestamp: 0, M: emptyPriceFilterInfoMap}
	SafeOrderInfo1      = &datastruct.SafeOrderInfo1{L: new(sync.RWMutex), OrderInfo: emptySafeOrderInfo1}
	// ListenKey stores the listen key received from binance
	ListenKey = &datastruct.SafeString{L: new(sync.RWMutex), Value: ""}
	// WantedSymbols stores the symbols should be subscribed ws stream
	WantedSymbols = &datastruct.SafeStrBoolMap{L: new(sync.RWMutex), M: emptyStrBoolMap}
	// Base2SymbolPairs stores all the symbol pairs for trading
	Base2SymbolPairs = &datastruct.SafeSymbolPairMap{L: new(sync.RWMutex), M: emptySymbolPairMap}
	// Symbol2SymbolPairs stores all the symbol pairs for trading
	Symbol2SymbolPairs = &datastruct.SafeSymbolPairMap{L: new(sync.RWMutex), M: emptySymbolPairMap}
	// FreeChargeSymbols stores the symbols with free maker and free taker
	FreeChargeSymbols = &datastruct.SafeStrBoolMap{L: new(sync.RWMutex), M: emptyStrBoolMap}
	// HoldCoins stores the coins with free
	HoldCoins = &datastruct.SafeStrBoolMap{L: new(sync.RWMutex), M: emptyStrBoolMap}
	// BaseAssetWhiteList stores the base asset white list
	BaseAssetWhiteList = &datastruct.SafeStrBoolMap{L: new(sync.RWMutex), M: emptyStrBoolMap}
	// SymbolBlackList stores the symbol black list
	SymbolBlackList = &datastruct.SafeStrBoolMap{L: new(sync.RWMutex), M: emptyStrBoolMap}
	// SymbolBlackList stores the symbol white list
	SymbolWhiteList = &datastruct.SafeStrBoolMap{L: new(sync.RWMutex), M: emptyStrBoolMap}
	// UsedWeight1m stores the 1m used weight from header of binance response
	UsedWeight1m = &datastruct.SafeInt{L: new(sync.RWMutex), Value: 0}
	// OrderCount10s stores the 10s order count from header of binance response
	OrderCount10s = &datastruct.SafeInt{L: new(sync.RWMutex), Value: 0}
	// OrderCount1d stores the 1d order count from header of binance header
	OrderCount1d = &datastruct.SafeInt{L: new(sync.RWMutex), Value: 0}
	// BinanceHosts stores all the valid binance host
	BinanceHosts = &datastruct.SafeStrBoolMap{L: new(sync.RWMutex), M: emptyStrBoolMap}
	// NetworkDelayMap stores network delay between different binance server and the current server
	NetworkDelayMap *datastruct.SafeNetworkDelayMap
	// RateLimit stores rate limit info from binance
	RateLimit = &datastruct.SafeRateLimit{L: new(sync.RWMutex), RateLimit: make([]binance.RateLimit, 0)}
	// SpotAsset2AssetInfo stores account info, key: base asset
	SpotAsset2AssetInfo = &datastruct.SafeAssetInfoMap{L: new(sync.RWMutex), M: make(map[string]datastruct.AssetInfo)}
	// FundingAsset2AssetInfo stores account info, key: base asset
	FundingAsset2AssetInfo = &datastruct.SafeAssetInfoMap{L: new(sync.RWMutex), M: make(map[string]datastruct.AssetInfo)}
	// BuyBNBTimestamp stores timestamp in ms of buying bnb last time
	BuyBNBTimestamp = &datastruct.SafeInt64{L: new(sync.RWMutex), Value: 0}
	// FlagTrade stores the trading status
	FlagTrade      = &datastruct.AutoTradeFlag{L: new(sync.RWMutex), Trading: false, TraceID: ""}
	OrderTraceFlag = &datastruct.AutoOrderTraceFlag{L: new(sync.RWMutex), M: make(map[string]bool)}
	// TradeValue stores the trade value
	TradeValue = &datastruct.SafeDecimal{L: new(sync.RWMutex), Value: define.Float0}
	// FlagGracefulExit stores the exiting flag
	FlagGracefulExit = &datastruct.SafeBool{L: new(sync.RWMutex), Value: false}
	// QuoteValueInfo stores the quote value info
	QuoteValueInfo = &datastruct.SafeQuoteQtyInfo{L: new(sync.RWMutex)}

	TraceId = &datastruct.SafeTraceId{L: new(sync.RWMutex), T: make(map[string]string)}

	// MinAssetValue stores the min asset value
	MinAssetValue = &datastruct.SafeDecimal{L: new(sync.RWMutex), Value: define.Float0}
	// AllAssetValueSum stores value of all assets
	AllAssetValueSum = &datastruct.SafeDecimal{L: new(sync.RWMutex), Value: define.Float0}
	// CumulativeReturnRate stores cumulative return rate
	CumulativeReturnRate = &datastruct.SafeCumulativeReturnRate{L: new(sync.RWMutex)}
	// DingTalkInfoMsg stores info message to sent to ding talk of error and warn message
	DingTalkInfoMsg = datastruct.NewStrBuf(1000)
	// DingTalkStartMsg stores info message to sent to ding talk of start and stop message
	DingTalkStartMsg = datastruct.NewStrBuf(1000)
	// CheckCostFeeCycle stores the trading count for checking cost fee
	CheckCostFeeCycle = datastruct.SafeInt{L: new(sync.RWMutex), Value: 0}
	// Symbol2UpdateIDMap stores the symbol with update id
	Symbol2UpdateIDMap = &datastruct.SafeStrInt64Map{L: new(sync.RWMutex), M: make(map[string]int64)}
	// CPUUsedPercent stores cpu usage of the progress
	CPUUsedPercent = &datastruct.SafeFloat64{L: new(sync.RWMutex), Value: 0}
	// ReturnRate stores the return rate info
	ReturnRate = &SafeReturanRate{L: new(sync.RWMutex)}
)

var (
	DefaultReturnRate = ReturnRateInfo{
		Min: define.Float2,
		Max: define.Float2,
	}
)

// DisplayInfo write some important info to log
func BuildDingTalkMsg() string {
	var msgList []string

	spotAccountInfo := SpotAsset2AssetInfo.Get()
	sJSON, err := json.Marshal(spotAccountInfo)
	if err == nil {
		msgList = append(msgList, utils.JoinStrWithSep("", "spot account info: ", string(sJSON)))
	}

	fundingAccountInfo := FundingAsset2AssetInfo.Get()
	fJSON, err := json.Marshal(fundingAccountInfo)
	if err == nil {
		msgList = append(msgList, utils.JoinStrWithSep("", "funding account info: ", string(fJSON)))
	}

	quoteValueInfo := QuoteValueInfo.Get()
	qJSON, err := json.Marshal(quoteValueInfo)
	if err == nil {
		msgList = append(msgList, utils.JoinStrWithSep("", "quote value info: ", string(qJSON)))
	}

	totalValue := AllAssetValueSum.Get()
	if err == nil {
		msgList = append(msgList, utils.JoinStrWithSep("", "total value: ", utils.Float64ToString(totalValue)))
	}

	msg := utils.JoinStrWithSep(define.SepDoubleEnter, msgList...)
	return msg
}

func GetSymbolWhiteList() []string {
	// 这里我们可以直接返回 SymbolWhiteList 的副本，避免额外的循环。
	return append([]string(nil), Config.Symbol.SymbolWhiteList...)
}
