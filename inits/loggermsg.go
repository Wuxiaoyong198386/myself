package inits

import (
	"fmt"
)

const (
	MsgOrderEventInfo = "logID=%s||msg=%s||symbol=%s||client_order_id=%s||side=%s||type=%s||jsondata=%s"
	MsgOrderEventMsg  = "Order information reminder"
	MsgOrderNew       = "The new order has been sincerely accepted"
	MsgOrderFilled    = "s1 trade has been successfully fulfilled"
	MsgOrderCanceled  = "order cancelled by user"

	MsgGetServerTime = "%s||ApiKey=%s||timestamp=%d"

	StartInfo              = "msg=Start program"
	EndCronMsg             = "msg=Cron execution ended"
	ErrorExchangeInfo      = "msg=%s||api_key=%s||secret_key=%s||err=%s"
	ErrorExchangeInfoMsg   = "msg=Transaction specification information acquisition failed"
	ErrorConfigFileNoFind  = "msg=Configuration file not found, please use - c to specify the configuration file"
	ErrorConfigRead        = "msg=Error reading configuration file"
	LoadConfigError        = "msg=Load config file failed||err="
	InitLoggerError        = "Initialize Log Failed||err="
	InitLoggerFileError    = "Initialize log file failed||err="
	DoSetProxyError        = "Setting network proxy failed"
	ErrorGetServerTime     = "msg=Getting server time failed||err="
	ErrorInitHttpPool      = "msg=Initializing HTTP link pool failed||err="
	ErrorInitWsPool        = "msg=Initializing Websocket link pool failed||err="
	ErrorCronStart         = "msg=Cron Plan Task Start Failed||err="
	ErrorBestPrice         = "msg=Getting the Best Price - Failed"
	ErrorListenkey         = "msg=%s||api_key=%s||secret_key=%s||err=%s"
	ErrorListenkeyMsg      = "msg=Failed to obtain listenkey"
	ErrorGetListenTime     = "msg=Obtaining key after %d attempts - failed"
	ErrorListenkeyDelayed  = "msg=Delay key failure,listen_key=%s||err=%s"
	ErrorCronMsg           = "%d Cron encountered an exception, please check the log file"
	ErrorSymbolsPairsEmpty = "msg=Transaction pair global variable is empty and cannot continue"

	SuccessListenkeyDelayed       = "msg=Delay key successful||listen_key=%s"
	SuccessLoadConfig             = "msg=Successfully truncated configuration file"
	SuccessInitLoger              = "msg=Successfully initialized log"
	SuccessProxy                  = "msg=Successfully set up network proxy"
	SuccessGetServerTime          = "msg=Successfully obtained server time"
	SuccessInitHttpPool           = "msg=Successfully initialized HTTP link pool"
	SuccessInitWsPool             = "msg=Successfully initialized ws link pool"
	SuccessInitHttpPoolLogmsg     = "msg=%s||Configuration timeout=%dms||Limit the total number of connections per host=%d|Control the maximum idle (active) connections that each host needs to maintain=%d||API=%s"
	SuccessCronStart              = "msg=Cron plan task successfully launched"
	SuccessExchange               = "msg=%s||Transaction pair quantity=%d||Price filter=%d||cost=%.3fms"
	SuccessExchangeMsg            = "msg=Successfully obtained transaction rules and transaction pair information"
	SuccessListenkey              = "msg=%s||api_key=%s||secret_key=%s||listen_key=%s"
	SuccessListenkeyMsg           = "Successfully get listen_key"
	SuccessFindSymbolPairs_format = "%s||count=%d||cost=%.3fms"
	SuccessFindSymbolPairs_msg    = "Successfully found transaction pairs"
	InfoDisabledProxy             = "The configuration item has disabled the network proxy"
	SuccessFindFilter             = "msg=Successfully synchronized transaction pair filters||trading_symbols=%d||price_filter_cnt=%d||cost=%.3fms"

	EmptyExchangeMsg = "The transaction specification information is empty and no data has been obtained"

	MsgProlongListenKey = "msg=Extend the interval between listening keys||interval=%ds"
	ErrorTrade_format   = "%s||source=%s||symbol=%s||update_id=%d||trace_id=%s"
	MsgCpuReaches       = "CPU usage percentage achieved target"

	MsgMarkWebSocket_source = "Mark websocket source"

	MsgExit                = "Exit signal detected, stop trading"
	MsgNetworkDelay_format = "msg=%s||network_delay=%.1fms||threshold=%.1fms||source=%s||symbol=%s||update_id=%d,trace_id=%s"
	MsgNetworkDelay        = "Transaction stopped due to network delay"
	MsgWeight1mThrd_format = "msg=%s||used_weight_1m=%d,threshold=%d,source=%s||update_id=%d||trace_id=%s"
	MsgWeight1mThrd        = "Trading stopped due to reaching a weight of 1 minute"

	MsgOrderCount10sThrd_format = "msg=%s||order_count_10s,%d||threshold=%d,source=%s||update_id=%d||trace_id=%s"
	MsgOrderCount10sThrd        = "Trading stopped due to reaching a weight of 10 seconds"

	MsgOrderCount1dThrd_format = "msg=%s||order_count_1d,%d||threshold=%d,source=%s||update_id=%d||trace_id=%s"
	MsgOrderCount1dThrd        = "Trading stopped due to reaching a weight of 1 day"

	MsgCheckLastTrading = "Stopped trading due to the previous transaction still in progress"

	ErrorNoGetBestPrice_format = "msg=%s||symbol=%s||update_id=%d||trace_id=%s"
	ErrorNoGetBestPrice        = "no get best price,stop trading"
	ErrorNoSpotFilter_format   = "msg=%s||symbol=%s||update_id=%d||trace_id=%s"
	ErrorNoSpotFilter          = "no get spot filter,stop trading"

	ErrorOrderTrades_format       = "logID=%s||s1 create order fail||symbol=%s||side=%s||price=%f||qty=%f||ClientOrderId=%s||traceID=%s||updateID=%d||response=%s"
	SuccessOrderTrades_format     = "logID=%s||s1 create order sucess||s1_symbol=%s||s1_side=%s||s1_price=%s||s1_qty=%s||s1_ClientOrderId=%s||traceID=%s||updateID=%d||response=%s||cost=%.3fms"
	ErrorS2S3OrderTrades_format   = "logID=%s||%s下单失败||symbol=%s||side=%s||price=%f||qty=%s||ClientOrderId=%s||traceID=%s||updateID=%d||response=%s"
	SuccessS2S3OrderTrades_format = "logID=%s||%s下单成功||symbol=%s||side=%s||price=%f||qty=%s||ClientOrderId=%s||traceID=%s||updateID=%d||cost=%.3fms||收益利率=%s"
)

func SuccessInfoMsg(s string) {
	fmt.Printf("%s\n", s)
}

func InfoMsg(msg string, err error) {
	fmt.Printf("%s%s\n", msg, err.Error())
}
func ErrorMsg(msg string, err error) {
	fmt.Printf("%s%s\n", msg, err.Error())
	InfoMsg(msg, err)
}
