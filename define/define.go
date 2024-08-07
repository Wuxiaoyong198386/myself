package define

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/shopspring/decimal"
)

const (
	ExpireTime            = "2024-12-01 00:00:00"
	OrderSucessStatusCode = 200
)

const (
	// endpoints
	EndpointFutureWS string = "wss://fstream.binance.com/ws"
	//
	TimeoutBinanceAPI time.Duration = 5 * time.Second // 5s
	ReconnectInterval time.Duration = 2 * time.Second // 2s
	SleepTime         time.Duration = 333 * time.Millisecond
	//
	MaxNetworkDelayCnt int = 5

	// binance host
	BinanceHost  string = "api.binance.com"
	BinanceHost1 string = "api1.binance.com"
	BinanceHost2 string = "api2.binance.com"
	BinanceHost3 string = "api3.binance.com"
	BinanceHost4 string = "api4.binance.com"

	// global quote coins
	QuoteUSDT  string = "USDT"
	QuoteTUSD  string = "TUSD"
	QuoteBUSD  string = "BUSD"
	QuoteBTC   string = "BTC"
	QuoteETH   string = "ETH"
	QuoteFDUSD string = "FDUSD"
	QuoteUSDC  string = "USDC"
	QuoteEUR   string = "EUR"
	QuoteBNB   string = "BNB"
	QuoteTRY   string = "TRY"

	// global quote coins
	CoinFDUSD string = "FDUSD"
	CoinUSDT  string = "USDT"
	CoinTUSD  string = "TUSD"
	CoinBUSD  string = "BUSD"
	CoinUSDC  string = "USDC"
	CoinAEUR  string = "AEUR"
	CoinEUR   string = "EUR"
	CoinBNB   string = "BNB"
	CoinBTC   string = "BTC"
	CoinETH   string = "ETH"
	CoinTRY   string = "TRY"
	CoinRUB   string = "RUB"
	CoinGBP   string = "GBP"
	CoinDAI   string = "DAI"

	// global symbols
	// with free charge
	SymbolBTCTUSD   string = "BTCTUSD"
	SymbolTUSDUSDT  string = "TUSDUSDT"
	SymbolTUSDBUSD  string = "TUSDBUSD"
	SymbolBUSDUSDT  string = "BUSDUSDT"
	SymbolUSDCUSDT  string = "USDCUSDT"
	SymbolBTCFDUSD  string = "BTCFDUSD"
	SymbolFDUSDUSDT string = "FDUSDUSDT"
	SymbolFDUSDBUSD string = "FDUSDBUSD"
	SymbolUSDPUSDT  string = "USDPUSDT"
	// with charge
	SymbolBTCBUSD string = "BTCBUSD"
	SymbolBTCUSDT string = "BTCUSDT"
	SymbolAPTETH  string = "APTETH"

	// buy BNB with USDT
	SymbolBNBUSDT string = "BNBUSDT"

	StatusTrading string = "TRADING"

	// asset
	AssetBNB string  = "BNB"
	USDT10   float64 = 10
	USDT15   float64 = 15
)

var (
	Decimal0       = decimal.NewFromFloat(0)
	Decimal001     = decimal.NewFromFloat(0.01)
	Decimal095     = decimal.NewFromFloat(0.95)
	Decimal099     = decimal.NewFromFloat(0.99)
	Decimal1       = decimal.NewFromFloat(1)
	Decimal1Point1 = decimal.NewFromFloat(1.1)
	Decimal1Point8 = decimal.NewFromFloat(1.8)
	Decimal10013   = decimal.NewFromFloat(1.0013)
	Decimal100005  = decimal.NewFromFloat(1.00005)
	Decimal2       = decimal.NewFromFloat(2)
	Decimal3       = decimal.NewFromFloat(3)
	Decimal13      = decimal.NewFromFloat(13)
	Decimal28      = decimal.NewFromFloat(28)
	Decimal30      = decimal.NewFromFloat(30)
	Decimal10000   = decimal.NewFromFloat(10000)
)
var (
	Float001        float64 = 0.01
	Float098        float64 = 0.98
	Float099        float64 = 0.99
	Float0999       float64 = 0.999
	Float1          float64 = 1
	Float1Point001  float64 = 1.001
	Float1Point02   float64 = 1.02
	Float1Point1    float64 = 1.1
	Float1Point8    float64 = 1.8
	Float1Point0013 float64 = 1.0013
	Float2          float64 = 2
	Float3          float64 = 3
	Float13         float64 = 13
	Float30         float64 = 30
	Float10000      float64 = 10000
	Float28         float64 = 28
	Float0          float64 = 0
)

const (
	ZeroFloat float64 = 1e-8

	MaxRetryCnt   int    = 3
	QuanlityConst string = "333333"

	Mod1000 int32 = 1000

	ExitMsgTransferFailure        string = "failed to transfer from funding to main"
	ExitMsgReachPlanQuoteQuantity string = "cumulative quantity reaches the plan quote quantity"
	ExitMsgBNBFree                string = "the progress will exit because of BNB"
	ExitMsgTotalValueThrwhold     string = "the progress will exit because total value is less than threshold"
	ExitMsgAssetZero              string = "one of the asset is zero"
	ExitMsgReachingExitingTime    string = "the progress will exit because of reaching exiting time"
	ExitMsgUnexpectedCostFee      string = "the progress will exit because of unexpected cost fee"
)

const (
	EventAccountUpdate binance.UserDataEventType = "outboundAccountPosition"
	EventBalanceUpdate binance.UserDataEventType = "balanceUpdate"
	EventOrderUpdate   binance.UserDataEventType = "executionReport"
)

var (
	HoldCoinMap = map[string]bool{}
)

const (
	SepEmpty           string = ""
	SepDoubleUnderline string = "__"
	SepDoubleEnter     string = "\n\n"

	DecimalPlaces3 int32 = 3
	DecimalPlaces8 int32 = 8
)

const (
	WSSourceBinance string = "binance"
)

const (
	ErrMsgAskOrBidValue string = "ask or bid value is too small"
)

const (
	SideTypeBuy     string = "BUY"
	SideTypeSell    string = "SELL"
	OrderType       string = "LIMIT"
	OrderTypeLimit  string = "LIMIT"
	OrderTypeMarket string = "MARKET"
	TimeInForceGtc  string = "GTC"
	TimeInForceIoc  string = "IOC"
	TimeInForceFok  string = "FOK"
	TimeInForceGtd  string = "GTD" //Good Till Date 在特定时间之前有效，到期自动撤销

	//条件价格触发类型
	WorkingTypeMark   string = "MARK_PRICE"     //标记价格
	WorkingTypeCPRICE string = "CONTRACT_PRICE" //合约最新价

	NewOrderRespAck    string = "ACK"
	NewOrderRespResult string = "RESULT"
)

const SHORT_NAME string = "SHORT"
const LONG_NAME string = "LONG"

const (
	OrderStatusNew       string = "NEW"
	OrderStatusPartially string = "PARTIALLY_FILLED"
	OrderStatusFilled    string = "FILLED"
	OrderStatusCanceled  string = "CANCELED"
	OrderStatusRejected  string = "REJECTED"
	OrderStatusExpired   string = "EXPIRED"
	OrderStatusPending   string = "PENDING_CANCEL"
	OrderStatusCancelled string = "CANCELED"

	OrderStatusTrade string = "TRADE"
)

func NewClientOrderID() string {
	return fmt.Sprintf("%d", time.Now().UnixMicro())
}

func NewClientOrderIDv2() string {
	randomBytes := make([]byte, 5)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err) // 通常情况下，rand.Reader不会返回错误
	}

	// 将随机字节转换为十六进制字符串
	randomHex := hex.EncodeToString(randomBytes)

	// 截取前10个字符作为结果
	randomHash := randomHex[:10]
	return randomHash
}
