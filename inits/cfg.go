package inits

import (
	"errors"
	"fmt"
	"go_code/myselfgo/modules/datastruct"
	"go_code/myselfgo/utils"
	"os"
	"sync"

	"github.com/open-binance/logger"
	"github.com/shopspring/decimal"
	"gopkg.in/yaml.v2"
)

const (
	AppName = "myselfgo"
	Version = "V1.0"
)

// Config is global config
var (
	Config *ServerConfig
)

// ServerConfig定义服务器的配置
type ServerConfig struct {
	Proxy            ProxyConfig            `yaml:"proxy" json:"proxy"`                           // 代理配置
	Account          AccountConfig          `yaml:"account" json:"account"`                       // 账号配置
	Notice           NoticeInfo             `yaml:"notice" json:"notice"`                         // config of all notice
	Diagnosis        DiagnosisConfig        `yaml:"diagnosis" json:"diagnosis"`                   // 诊断配置
	HTTPClient       HTTPClientInfo         `yaml:"http_client" json:"http_client"`               // http client配置
	Binance          BinanceConfig          `yaml:"binance" json:"binance"`                       // binance server配置
	ReturnRate       ReturnRateConfig       `yaml:"return_rate" json:"return_rate"`               // 退货率配置
	ReturnRateUpdate ReturnRateUpdateConfig `yaml:"return_rate_update" json:"return_rate_update"` // 更新回报率配置
	Interval         IntervalConfig         `yaml:"interval" json:"interval"`                     // 间隔的配置
	Custom           CustomConfig           `yaml:"custom" json:"custom"`                         // 自定义配置
	Exit             ExitConfig             `yaml:"exit" json:"exit"`                             // 退出配置
	Log              LogConfig              `yaml:"log" json:"log"`                               // 日志配置
	File             FileConfigInfo         `yaml:"file" json:"file"`                             // 写入消息的文件配置
	Rebalance        RebalanceInfo          `yaml:"rebalance" json:"rebalance"`                   //调仓配置
	Symbol           SymbolConfig           `yaml:"symbol" json:"symbol"`                         // 交易对配置
	Symbols          []string               `yaml:"symbols" json:"symbols"`                       // 交易对列表配置
	Order            OrderConfig            `yaml:"order" json:"order"`
	Kline            KlineConfig            `yaml:"kline" json:"kline"`
	MysqlClient      MysqlConfig            `yaml:"mysql_client" json:"mysql_client"` // 数据库配置
}
type OrderConfig struct {
	Enable           bool     `yaml:"enable" json:"enable"` // 是否开启下单
	Side             int      `yaml:"side" json:"side"`     // 交易方向
	Default_quantity float64  `yaml:"default_quantity" json:"default_quantity"`
	Warehouse_mode   int      `yaml:"warehouse_mode" json:"warehouse_mode"`
	Warehouse_cnt    []string `yaml:"warehouse_cnt" json:"warehouse_cnt"`
	Stop_loss        float64  `yaml:"stop_loss" json:"stop_loss"`
	Stop_profit      float64  `yaml:"stop_profit" json:"stop_profit"`
	Max_loss         float64  `yaml:"max_loss" json:"max_loss"`
	Max_profit       float64  `yaml:"max_profit" json:"max_profit"`
	Open_mb          bool     `yaml:"open_mb" json:"open_mb"`
	Bollb            float64  `yaml:"bollb" json:"bollb"`
	Max_price        float64  `yaml:"max_price" json:"max_price"`
	Market_rate      float64  `yaml:"market_rate" json:"market_rate"`
}

type KlineConfig struct {
	Kine_type     string `yaml:"kine_type" json:"kine_type"`
	Kine_count    int    `yaml:"kine_count" json:"kine_count"`
	Macd_open     bool   `yaml:"macd_open" json:"macd_open"`
	Macd_fast     int    `yaml:"macd_fast" json:"macd_fast"`
	Macd_slow     int    `yaml:"macd_slow" json:"macd_slow"`
	Macd_signal   int    `yaml:"macd_signal" json:"macd_signal"`
	Kine_interval string `yaml:"kine_interval" json:"kine_interval"`
}

// ProxyConfig 指定代理的配置
type ProxyConfig struct {
	Enable  bool   `yaml:"enable" json:"enable"`
	Address string `yaml:"address" json:"address"`
}

// AccountConfig 指定帐户的配置
type AccountConfig struct {
	ApiKey    string `yaml:"api_key" json:"api_key"`
	SecretKey string `yaml:"secret_key" json:"secret_key"`
}

type ReturnRateUpdateConfig struct {
	Enable bool                 `yaml:"enable" json:"enable"` // enable updating return rate or not, true: enable, false: disabled
	Info0  ReturnRateDetailInfo `yaml:"info0" json:"info0"`
	Info1  ReturnRateDetailInfo `yaml:"info1" json:"info1"`
	Info2  ReturnRateDetailInfo `yaml:"info2" json:"info2"`
	Info3  ReturnRateDetailInfo `yaml:"info3" json:"info3"`
	Info4  ReturnRateDetailInfo `yaml:"info4" json:"info4"`
}

type ReturnRateDetailInfo struct {
	DeltaTime       int64            `yaml:"delta_time" json:"delta_time"` // unit: second
	DeltaQuoteValue float64          `yaml:"delta_quote_value" json:"delta_quote_value"`
	TradeValue      float64          `yaml:"trade_value" json:"trade_value"`
	ReturnRate      ReturnRateConfig `yaml:"return_rate" json:"return_rate"`
}

type ReturnRateNotice struct {
	Enable            bool    `yaml:"enable" json:"enable"`                         // 是否能够通知退货率
	NegativeThreshold float64 `yaml:"negative_threshold" json:"negative_threshold"` // 负阈值通知
	PositiveThreshold float64 `yaml:"positive_threshold" json:"positive_threshold"` // 注意的正阈值
}

type HTTPClientInfo struct {
	Common HTTPClientPoolConfig `yaml:"common" json:"common"` // Common http client with no specified api
}

// HTTPClientPoolConfig specifies the config for http client
type HTTPClientPoolConfig struct {
	Timeout             int    `yaml:"timeout" json:"timeout"`                                 // specifie a time limit for requests made by this http client
	MaxConnsPerHost     int    `yaml:"max_conns_per_host" json:"max_conns_per_host"`           // limit the total number of connections per host
	MaxIdleConnsPerHost int    `yaml:"max_idle_conns_per_host" json:"max_idle_conns_per_host"` // control the maximum idle(keep-alive) connections to keep per host
	API                 string `yaml:"api" json:"api"`                                         // api address of http server
}

// BinanceConfig 指定binance的接口地址
type BinanceConfig struct {
	Adaptive bool   `yaml:"adaptive" json:"adaptive"` // 如果值为false，它将使用文件中设置的主机的值；相反，它将根据网络延迟使用自适应主机
	Host     string `yaml:"host" json:"host"`         // host of binance
}

// ReturnRateConfig specifies the config of return rate
type ReturnRateConfig struct {
	Free0 ReturnRateInfo `yaml:"free0" json:"free0"` // return rate for no symbol with free charge
	Free1 ReturnRateInfo `yaml:"free1" json:"free1"` // return rate for 1 symbol with free charge
	Free2 ReturnRateInfo `yaml:"free2" json:"free2"` // return rate for 2 symbols with free charge
	Free3 ReturnRateInfo `yaml:"free3" json:"free3"` // return rate for 3 symbols with free charge
}

type ReturnRateInfo struct {
	Min float64 `yaml:"min" json:"max"`  // minimum return rate
	Max float64 `yaml:"max" json:"maxx"` // maximum return rate
}

// IntervalConfig specifies the config of interval
type IntervalConfig struct {
	BestPriceWrite     int     `yaml:"best_price_write" json:"best_price_write"`         // unit: ms, 将最优惠价格写入日志文件的间隔
	PriceFilterInfo    int     `yaml:"price_filter_info" json:"price_filter_info"`       // unit: s, 从binance同步筛选器信息的间隔
	ListenKey          int     `yaml:"listen_key" json:"listen_key"`                     // unit: s, 从binance同步侦听密钥的间隔
	SymbolRefresh      int     `yaml:"symbol_refresh" json:"symbol_refresh"`             // unit: s, 从binance刷新交易对的间隔
	NetworkDelay       int     `yaml:"network_delay" json:"network_delay"`               // unit: ms, 二进制网络延迟的更新间隔
	HostAndDelay       int     `yaml:"host_and_delay" json:"host_and_delay"`             // unit: ms, 二进制的更新主机间隔和网络延迟
	UpdateOrderCnt     int     `yaml:"update_order_cnt" json:"update_order_cnt"`         // unit: s, 更新订单计数的间隔
	Check1dOrderCnt    int     `yaml:"check_1d_order_cnt" json:"check_1d_order_cnt"`     // unit: s, 检查1d订单计数的间隔
	SyncFundingAccount int     `yaml:"sync_funding_account" json:"sync_funding_account"` // unit: s, 资金账户信息同步间隔
	SyncSpotAccount    int     `yaml:"sync_spot_account" json:"sync_spot_account"`       // unit: s, 同步现货帐户信息的间隔
	SyncBestPrice      int     `yaml:"sync_best_price" json:"sync_best_price"`           // unit: ms, 单位：毫秒，同步最优惠价格的间隔#使用http同步最优惠的价格
	CheckExit          int     `yaml:"check_exit" json:"check_exit"`                     // unit: s, 检查出口间隔
	GenDingTalkMsg     int     `yaml:"gen_ding_talk_msg" json:"gen_ding_talk_msg"`       // unit: s, 生成丁talk消息的间隔
	SendDingTalkMsg    int     `yaml:"send_ding_talk_msg" json:"send_ding_talk_msg"`     // unit: ms, 发送钉钉通话信息的间隔
	TakeUpTime         float64 `yaml:"take_up_time" json:"take_up_time"`                 // unit: ms, 从获取event交易对开始，到创建订单前的时间，如果超过此值，就放弃此次套利
	WsSocketCnt        int     `yaml:"ws_socket_cnt" json:"ws_socket_cnt"`               // 初始化wssocket的数量

}

// CustomConfig specifies the config of custom
type CustomConfig struct {
	GiveUpSome                   bool            `yaml:"give_up_some" json:"give_up_some"`                                       // if true, it will give up the group whose quantity of s1 is less than s2
	PriceGoUp                    bool            `yaml:"price_go_up" json:"price_go_up"`                                         // if true, price of s1 will multiple 1.00005
	SellS2AtOnce                 bool            `yaml:"sell_s2_at_once" json:"sell_s2_at_once"`                                 // if true, 立即出售s2
	TransferExit                 bool            `yaml:"transfer_exit" json:"transfer_exit"`                                     // if true, 如果传输失败，它将退出
	SellBNB                      bool            `yaml:"sell_bnb" json:"sell_bnb"`                                               // if true, 它将在优雅退出前出售BNB
	CheckPriceBeforeTrade        bool            `yaml:"check_price_before_trade" json:"check_price_before_trade"`               // if true, 它会在交易前检查价格
	SleepTimeBeforeCheck         int             `yaml:"sleep_time_before_check" json:"sleep_time_before_check"`                 // unit: ms, 在检查价格之前的睡眠时间，当该值大于0时，这将是有意义的
	SleepBeforeSell              int             `yaml:"sleep_before_sell" json:"sleep_before_sell"`                             // unit: s, 出售前的睡眠时间，当值大于0时，这将是有意义的
	SymbolBatch                  int             `yaml:"symbol_batch" json:"symbol_batch"`                                       // 订阅一个ws流中的交易对数
	UsedWeight1m                 int             `yaml:"used_weight_1m" json:"used_weight_1m"`                                   // 如果1秒的使用权重大于阈值，则停止交易
	OrderCount10s                int             `yaml:"order_count_10s" json:"order_count_10s"`                                 // 如果10s订单计数大于阈值，则停止交易
	OrderCount1d                 int             `yaml:"order_count_1d" json:"order_count_1d"`                                   // 如果1d订单计数大于阈值，则停止交易
	WaitTimeNextTrade            int             `yaml:"wait_time_next_trade" json:"wait_time_next_trade"`                       // unit: ms, 进行下一次交易前等待的时间
	SubscribeCnt                 int             `yaml:"subscribe_cnt" json:"subscribe_cnt"`                                     // websocket订阅计数
	SleepBeforeExit              int             `yaml:"sleep_before_exit" json:"sleep_before_exit"`                             // unit: s, 退出前的睡眠时间，当该值大于0时，这将是有意义的
	SleepTime                    int             `yaml:"sleep_time" json:"sleep_time"`                                           // unit: ms, 交易失败时的睡眠时间
	AwesomeSleepTime             int             `yaml:"awesome_sleep_time" json:"awesome_sleep_time"`                           // unit: s, 当达到惊人的损失率时的睡眠时间
	SleepTimeInsufficientBalance int             `yaml:"sleep_time_insufficient_balance" json:"sleep_time_insufficient_balance"` // unit: ms, 余额不足的睡眠时间
	CheckCostFeeCycle            int             `yaml:"check_cost_fee_cycle" json:"check_cost_fee_cycle"`                       // 检查成本费周期
	BuyBNB                       int64           `yaml:"buy_bnb" json:"buy_bnb"`                                                 // 购买bnb的间隔
	NetworkDelay                 float64         `yaml:"network_delay" json:"network_delay"`                                     // 如果网络延迟大于阈值，则停止交易, unit: ms
	CpuThreshold                 float64         `yaml:"cpu_threshold" json:"cpu_threshold"`                                     // 当cpu使用百分比达到阈值时发送ding-talk消息
	GiveUpTrade                  float64         `yaml:"give_up_trade" json:"give_up_trade"`                                     // unit: ms, 因为成本问题而放弃交易
	PickUpThreshold              float64         `yaml:"pick_up_threshold" json:"pick_up_threshold"`                             // 拾取阈值
	AwesomeLossRate              decimal.Decimal `yaml:"awesome_loss_rate" json:"awesome_loss_rate"`                             // 惊人的损失率
	BNBCntBuy                    float64         `yaml:"bnb_cnt_buy" json:"bnb_cnt_buy"`                                         // 如果bnb数量小于阈值，则购买bnb
	BNBCntExit                   float64         `yaml:"bnb_cnt_exit" json:"bnb_cnt_exit"`                                       // 如果bnb的计数小于阈值，则退出进度
	ValueThrd                    decimal.Decimal `yaml:"value_thrd" json:"value_thrd"`                                           // 跳过要价或出价低于阈值的交易对
	SubValue                     float64         `yaml:"sub_value" json:"sub_value"`                                             // 市场价值将低于sub_value
	QuoteValue                   decimal.Decimal `yaml:"quote_value" json:"quote_value"`                                         // 当累计报价值达到该值时，进度将退出
	PriceGoUpRate                float64         `yaml:"price_go_up_rate" json:"price_go_up_rate"`                               // 价格上涨倍数
	MinAssetValue                decimal.Decimal `yaml:"min_asset_value" json:"min_asset_value"`                                 // 默认情况下的最小资产价值
	ExpectedCostFee              decimal.Decimal `yaml:"expected_cost_fee" json:"expected_cost_fee"`                             // 当前api密钥的预期成本费用
	TradeBidValue                decimal.Decimal `yaml:"trade_bid_value" json:"trade_bid_value"`                                 // 交易 买单价
	TradeAskValue                decimal.Decimal `yaml:"trade_ask_value" json:"trade_ask_value"`                                 // 交易 卖单价
	LossRate                     float64         `yaml:"loss_rate" json:"loss_rate"`                                             // 损失率
	S1S2Multiple                 float64         `yaml:"s1_s2_multiple" json:"s1_s2_multiple"`                                   // 倍数
	ReturnRateMultiple           decimal.Decimal `yaml:"return_rate_multiple" json:"return_rate_multiple"`                       // 收益率倍数
	CustomWS                     []string        `yaml:"custom_ws" json:"custom_ws"`
	S1ExitReturn                 float64         `yaml:"s1_exit_return" json:"s1_exit_return"`
	S1OrderValue                 float64         `yaml:"s1_order_value" json:"s1_order_value"`
	ChangePrice                  float64         `yaml:"change_price" json:"change_price"`
	DoUpTj3                      int             `yaml:"do_up_tj3" json:"do_up_tj3"`
	DoDownTj3                    int             `yaml:"do_down_tj3" json:"do_down_tj3"`
}

type ExitConfig struct {
	Enable      bool  `yaml:"enable" json:"enable"`             // if true, 进程将在下一天的07:59退出
	LeftSeconds int64 `yaml:"left_seconds" json:"left_seconds"` // 如果该值大于0，进程将在LeftSeconds秒后退出
}

type LogConfig struct {
	Detail bool   `yaml:"detail" json:"detail"` // 如果为true，则显示详细信息日志
	Level  string `yaml:"level" json:"level"`   // 日志等级 debug, info, warn, error
}

// FileConfigInfo specifies all files to write raw message
type FileConfigInfo struct {
	BestPrice  FileConfig `yaml:"best_price" json:"best_price"`
	ReturnRate FileConfig `yaml:"return_rate" json:"return_rate"`
}
type NoticeInfo struct {
	DingTalk DingTalkInfo `yaml:"ding_talk" json:"ding_talk"` // config of all notice
}

type DingTalkInfo struct {
	Enable       bool           `yaml:"enable" json:"enable"`                 // start and stop
	StartAndStop DingTalkConfig `yaml:"start_and_stop" json:"start_and_stop"` // start and stop
	InfoLog1     DingTalkConfig `yaml:"info_log1" json:"info_log1"`           // error and warn
	InfoLog2     DingTalkConfig `yaml:"info_log2" json:"info_log2"`           // error and warn

}

type DingTalkConfig struct {
	Name    string `yaml:"name" json:"name"`       // name of the ding talk
	Webhook string `yaml:"webhook" json:"webhook"` // webhook from ding talk
	Keyword string `yaml:"keyword" json:"keyword"` // keyword of the custom robot
}

// DiagnosisConfig specifies the config of pprof for diagnosis
type DiagnosisConfig struct {
	Enable     bool   `yaml:"enable" json:"enable"`           // enable pprof and prometheus or not, true: enable, false: disabled
	Port       int    `yaml:"port" json:"port"`               // port of the pprof and prometheus
	Interval   int    `yaml:"interval" json:"interval"`       // interval to save cpu and heap profile, unit: second
	Seconds    int    `yaml:"seconds" json:"seconds"`         // look at a 30-second CPU profile when set it as 30
	Outdate    int64  `yaml:"outdate" json:"outdate"`         // time to delete the profile, unit: second
	ProfileDir string `yaml:"profile_dir" json:"profile_dir"` // directory to store cpu and heap profile
}

// FileConfig specifies detail config
type FileConfig struct {
	Compress   bool   `yaml:"compress" json:"compress"`       // determine if the rotated log files should be compressed using gzip
	MaxSize    int    `yaml:"max_size" json:"max_size"`       // the max size in megabytes of the log file before it gets rotated
	MaxBackups int    `yaml:"max_backups" json:"max_backups"` // the max number of old log file to retain
	MaxAge     int    `yaml:"max_age" json:"max_age"`         // the max number of days to retain old log files based on the timestamp encoded in their filename
	Level      string `yaml:"level" json:"level"`             // logger level
	Filename   string `yaml:"filename" json:"filename"`       // filename is the file to write logs to
}

type RebalanceInfo struct {
	Interval  int             `yaml:"interval" json:"interval"`   // unit: s, interval of rebalance
	Threshold decimal.Decimal `yaml:"threshold" json:"threshold"` // if result of the max substract of the min reaches the target, it will rebalance
}

// SymbolConfig specifies symbol config
type SymbolConfig struct {
	OnlyWhiteList       bool            `yaml:"only_white_list" json:"only_white_list"`               // true means only trade the symbol in the white list
	EnableBlackList     bool            `yaml:"enable_black_list" json:"enable_black_list"`           // true means that symbol in the symbol_black_list will not be subscribed
	PriceFilterRateThrd float64         `yaml:"price_filter_rate_thrd" json:"price_filter_rate_thrd"` //
	BaseWhiteList       []string        `yaml:"base_white_list" json:"base_white_list"`               // base asset white list
	SymbolBlackList     []string        `yaml:"symbol_black_list" json:"symbol_black_list"`           // symbol black list
	RootOpen            bool            `yaml:"root_open" json:"root_open"`                           // root open or not
	RootSymbol          string          `yaml:"root_symbol" json:"root_symbol"`                       // root symbol
	RootType            int             `yaml:"root_type" json:"root_type"`
	RootStopMacd        bool            `yaml:"root_stop_macd" json:"root_stop_macd"`
	Type                int             `yaml:"type" json:"type"`                             //1  现货  2 合约
	SymbolWhiteList     []string        `yaml:"symbol_white_list" json:"symbol_white_list"`   // symbol white list
	SymbolFreeCharge    []string        `yaml:"symbol_free_charge" json:"symbol_free_charge"` // symbol with free charge
	HoldCoins           []string        `yaml:"hold_coins" json:"hold_coins"`                 // coins with free
	SkipS2Quote         skipS2QuoteInfo `yaml:"skip_s2_quote" json:"skip_s2_quote"`           // skip the trading pair whose quote of s2 hists the list
	ProbabilityMax      float64         `yaml:"probability_max" json:"probability_max"`
	ProbabilityMin      float64         `yaml:"probability_min" json:"probability_min"`
	QuoteVolumeMin      float64         `yaml:"quote_volume_min" json:"quote_volume_min"`
	Oscillation         float64         `yaml:"oscillation" json:"oscillation"`
	SymbolCount         int             `yaml:"symbol_count" json:"symbol_count"`
}

type skipS2QuoteInfo struct {
	Enable bool     `yaml:"enable" json:"enable"`
	Coins  []string `yaml:"coins" json:"coins"`
}
type MysqlConfig struct {
	Host     string `yaml:"host" json:"host"` // true means only trade the symbol in the white list
	Port     string `yaml:"port" json:"port"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
	Db       string `yaml:"db" json:"db"`
	DbName   string `yaml:"db_name" json:"db_name"`
}

/*
 * @description: 从配置文件中加载配置项
 * @fileName: cfg.go
 * @author: vip120@126.com
 * @date: 2024-03-20 14:33:06
 */
func LoadConfig(filename string) error {
	if filename == "" {
		// 如果配置文件名为空，返回错误信息
		return errors.New(ErrorConfigFileNoFind)
	}

	// 读取配置文件内容
	contentBytes, err := os.ReadFile(filename)
	if err != nil {
		// 如果读取文件出错，返回错误信息
		return fmt.Errorf("%s,错误原因: %s", ErrorConfigRead, err.Error())
	}

	var sc ServerConfig
	// 将读取的内容反序列化为 ServerConfig 对象
	if err := yaml.Unmarshal(contentBytes, &sc); err != nil {
		// 如果反序列化出错，返回错误信息
		return fmt.Errorf("failed to unmarshal, err: %s", err.Error())
	}
	// 检查 ServerConfig 对象的网络配置参数是否合法
	if err := sc.Check(); err != nil {
		// 如果参数不合法，返回错误信息
		return fmt.Errorf("failed to check param, err: %s", err.Error())
	}

	// 将 ServerConfig 对象赋值给全局变量 Config
	Config = &sc

	// 初始化网络延时
	NetworkDelayMap = datastruct.NewNetworkDelay(sc.Binance.Adaptive, sc.Binance.Host)

	// 初始化 symbol white list
	buildData()

	return nil
}

func (sc *ServerConfig) Check() error {
	// 获取Binance配置
	binanceCfg := sc.Binance
	// 如果不启用自适应模式
	if !binanceCfg.Adaptive {
		// 获取主机地址
		host := binanceCfg.Host
		// 判断主机地址是否存在于支持的列表中
		if existed := BinanceHosts.IsExisted(host); !existed {
			// 如果不存在，返回错误
			return fmt.Errorf("invalid host when adaptive is false, only supports: %v, but got: %s",
				BinanceHosts.Keys(), host)
		}
	}

	// 检查通过，返回nil
	return nil
}

// DoSetProxy 函数用于设置代理。
// 如果代理未启用，则输出一条成功信息并返回nil。
// 如果代理启用，则根据代理地址进行代理设置。
// 如果代理设置成功，则输出一条成功信息并返回nil。
// 如果代理设置失败，则输出错误日志并返回错误信息。
func DoSetProxy() error {
	// 如果代理未启用，则输出成功信息并返回
	if !Config.Proxy.Enable {
		SuccessInfoMsg(InfoDisabledProxy)
		return nil
	}

	// 获取代理地址
	address := Config.Proxy.Address

	// 设置代理
	if err := utils.SetProxy(address); err != nil {
		// 如果设置代理失败，则记录错误日志并返回错误
		logger.Errorf("%s,代理地址：%s，错误原因：%s", DoSetProxyError, address, err.Error())
		return err
	}

	// 设置代理成功，记录日志并输出成功信息
	logger.Infof("%s,代理地址：%s", SuccessProxy, address)
	SuccessInfoMsg(SuccessProxy)

	return nil
}
func buildData() {
	// 创建一个用于存储白名单的map，其中key为符号，value始终为true
	symbolWhiteList := make(map[string]bool) // key: symbol, value: always be true

	// 遍历配置中的白名单符号列表
	for _, symbol := range Config.Symbol.SymbolWhiteList {
		// 将每个符号作为key添加到map中，value为true
		symbolWhiteList[symbol] = true
	}

	// 重新初始化SymbolWhiteList变量，将之前构建的白名单map作为参数传入
	SymbolWhiteList.ReInit(symbolWhiteList)
}

type SafeReturanRate struct {
	L  *sync.RWMutex
	RR ReturnRateConfig
}

func (sr *SafeReturanRate) Set(rr ReturnRateConfig) {
	sr.L.Lock()
	defer sr.L.Unlock()
	sr.RR = rr
}

func (sr *SafeReturanRate) Get() ReturnRateConfig {
	sr.L.RLock()
	defer sr.L.RUnlock()
	return sr.RR
}
