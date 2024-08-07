package cron

import (
	"context"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"go_code/myselfgo/client"
	"go_code/myselfgo/define"
	"go_code/myselfgo/inits"
	"go_code/myselfgo/modules/datastruct"
	"go_code/myselfgo/notice"
	"go_code/myselfgo/tradgo"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/open-binance/logger"
	"github.com/shopspring/decimal"
)

func StartHttpKlinesRootSymbol() {
	symbol := inits.Config.Symbol.RootSymbol
	if inits.Config.Symbol.RootOpen {
		go SyncKlineOnceRootSymbol(symbol)
	}

}

func SyncKlineOnceRootSymbol(symbol string) error {
	acount := inits.Config.Account
	apiKey := acount.ApiKey
	secretKey := acount.SecretKey
	ctx, cancel := context.WithTimeout(context.Background(), define.TimeoutBinanceAPI)
	defer cancel()
	// 获取当前时间
	now := time.Now()
	// 截断到分钟（去掉秒和纳秒部分）
	truncated := now.Truncate(time.Minute)
	// 开始时间
	startK := inits.Config.Kline.Kine_count - 1
	startTime := truncated.Add(-time.Duration(startK) * time.Minute)
	// 获取前一分钟结束时间的Unix时间戳（秒为单位）
	endTimeUnix := truncated.Unix()
	if inits.Config.Symbol.RootType == 2 {
		kList, err := client.GetFutureClient(apiKey, secretKey).NewKlinesService().Symbol(symbol).Interval(inits.Config.Kline.Kine_type).StartTime(startTime.Unix() * 1000).EndTime(endTimeUnix * 1000).Do(ctx)
		if err != nil {
			logger.Infof("msg=Hy NewKlinesService.Do err=%s", err)
			return err
		}
		WKlineInfo_Hy(symbol, kList)
	} else {
		kList, err := client.GetClient(apiKey, secretKey).NewKlinesService().Symbol(symbol).
			Interval(inits.Config.Kline.Kine_type).StartTime(startTime.Unix() * 1000).EndTime(endTimeUnix * 1000).Do(ctx)
		if err != nil {
			logger.Infof("msg=Xh NewKlinesService.Do err=%s", err)
			return err
		}
		WKlineInfo_Xh(symbol, kList)
	}
	return nil
}

func WKlineInfo_Hy(symbol string, kList []*futures.Kline) {
	var klineEvents []datastruct.WsKlineEvent // 使用切片存储Kline数据
	for _, kline := range kList {             // 不需要索引index
		klineEvents = append(klineEvents, datastruct.WsKlineEvent{
			Kline: datastruct.WsKline{
				StartTime:   kline.OpenTime,
				EndTime:     kline.CloseTime,
				Symbol:      symbol, // 使用函数参数中的 symbol
				Open:        kline.Open,
				High:        kline.High,
				Low:         kline.Low,
				Close:       kline.Close,
				Volume:      kline.Volume,
				QuoteVolume: kline.QuoteAssetVolume,
				TradeNum:    kline.TradeNum,
			},
		})
	}
	inits.KlineInfo.ReInit(symbol, klineEvents)
}

func WKlineInfo_Xh(symbol string, kList []*binance.Kline) {
	var klineEvents []datastruct.WsKlineEvent // 使用切片存储Kline数据
	for _, kline := range kList {             // 不需要索引index
		klineEvents = append(klineEvents, datastruct.WsKlineEvent{
			Kline: datastruct.WsKline{
				StartTime:   kline.OpenTime,
				EndTime:     kline.CloseTime,
				Symbol:      symbol, // 使用函数参数中的 symbol
				Open:        kline.Open,
				High:        kline.High,
				Low:         kline.Low,
				Close:       kline.Close,
				Volume:      kline.Volume,
				QuoteVolume: kline.QuoteAssetVolume,
				TradeNum:    kline.TradeNum,
			},
		})
	}
	inits.KlineInfo.ReInit(symbol, klineEvents)
}

func StartHttpKlinesServe2() {

	logger.Infof("合约同步K线开始,间隔:%dm", 1)
	// 先同步一次
	symbols := inits.Config.Symbol.SymbolWhiteList
	// fmt.Println("symbols-----:", symbols)
	for _, symbol := range symbols {
		go SyncKlineOnce2(symbol)
	}

}

func SyncKlineOnce2(symbol string) error {
	acount := inits.Config.Account
	apiKey := acount.ApiKey
	secretKey := acount.SecretKey
	ctx, cancel := context.WithTimeout(context.Background(), define.TimeoutBinanceAPI)
	defer cancel()
	// 获取当前时间
	now := time.Now()
	// 截断到分钟（去掉秒和纳秒部分）
	truncated := now.Truncate(time.Minute)
	// 开始时间
	startTime := truncated.Add(-99 * time.Minute)
	//year2000 := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)

	// 从2000年1月1日00:00:00 UTC开始减去21分钟
	//startTime := year2000.Add(-21 * time.Minute)
	// 前一分钟的结束时间（即当前分钟的第0秒，但通常用于表示一分钟的结束）
	// 注意：由于Unix时间戳是基于整点的，我们通常使用当前分钟的开始来代表“前一分钟的结束”
	// endTime := truncated.Add(-1 * time.Minute)
	// 获取前一分钟结束时间的Unix时间戳（秒为单位）
	endTimeUnix := truncated.Unix()
	// ftime := formatTime()

	// fmt.Println(startTime.Format("2006ç-01-02 15:04:05"),
	// 	endTime.Format("2006-01-02 15:04:05"), ftime)

	kList, err := client.GetFutureClient(apiKey, secretKey).NewKlinesService().Symbol(symbol).
		Interval(inits.Config.Kline.Kine_type).StartTime(startTime.Unix() * 1000).EndTime(endTimeUnix * 1000).Do(ctx)
	if err != nil {
		logger.Infof("msg=symbol=%s,NewKlinesService failed: %v", symbol, err)
		return err
	}
	var klineEvents []datastruct.WsKlineEvent // 使用切片存储Kline数据
	for _, kline := range kList {             // 不需要索引index
		klineEvents = append(klineEvents, datastruct.WsKlineEvent{
			Kline: datastruct.WsKline{
				StartTime:   kline.OpenTime,
				EndTime:     kline.CloseTime,
				Symbol:      symbol, // 使用函数参数中的 symbol
				Open:        kline.Open,
				High:        kline.High,
				Low:         kline.Low,
				Close:       kline.Close,
				Volume:      kline.Volume,
				QuoteVolume: kline.QuoteAssetVolume,
				TradeNum:    kline.TradeNum,
			},
		})
	}
	inits.KlineInfo.ReInit(symbol, klineEvents)
	return nil
}

func StrToFloat64(numStr string) float64 {
	num, err := strconv.ParseFloat(numStr, 64) // 64表示使用64位浮点数精度
	if err != nil {
		return 0 // 如果转换失败，返回错误
	}
	return num
}

func FormatTime() string {
	now := time.Now()
	formatted := now.Format("2006-01-02 15:04:05")
	return formatted
}

// CountDecimalPlaces 计算浮点数的小数位数
func CountDecimalPlaces(f string) int {
	// 将浮点数转换为字符串
	df, _ := decimal.NewFromString(f)
	// 查找小数点的位置
	dfstring := df.String()
	decimalIdx := strings.IndexByte(dfstring, '.')
	if decimalIdx == -1 {
		// 没有小数点，返回0
		return 0
	}
	parts := strings.Split(dfstring, ".")
	return len(parts[1])
}

// 四舍五入浮点数
func RoundToNthDecimal45(num float64, precision int) float64 {
	var round float64
	pow := math.Pow(10, float64(precision))
	intermed := num * pow
	if math.Floor(intermed) < intermed {
		round = math.Ceil(intermed)
	} else {
		round = math.Floor(intermed)
	}
	return round / pow
}
func RoundToNthDecimal(num float64, precision int) float64 {
	// 计算10的precision次方
	pow := math.Pow(10, float64(precision))

	// 将num乘以pow，得到中间值
	intermed := num * pow

	// 对中间值向下取整
	truncated := math.Round(intermed)

	// 将向下取整后的值除以pow，得到最终结果
	return truncated / pow
}

func RoundDecimal(d decimal.Decimal, n int) decimal.Decimal {
	// 确定小数点后总的位数
	// 首先确定d的整数部分有多少位
	//intPart := d.IntPart()
	//intPartStr := strconv.FormatInt(intPart, 10)
	//intPartLen := len(intPartStr)

	// 然后加上要保留的小数位数
	totalDigitsAfterDecimal := n

	// 使用Round方法进行四舍五入
	rounded := d.Round(int32(totalDigitsAfterDecimal))
	return rounded
}

func GetStopPrice(sideType futures.SideType, k1_high, k2_high, k1_low, k2_low decimal.Decimal) decimal.Decimal {
	var price decimal.Decimal
	//做空
	if sideType == futures.SideTypeSell {
		if k1_high.GreaterThanOrEqual(k2_high) {
			price = k1_high
		} else {
			price = k2_high
		}

	} else {
		if k1_low.GreaterThanOrEqual(k2_low) {
			price = k2_low
		} else {
			price = k1_low
		}

	}
	//price := k1_high.Add(k1_low).Div(decimal.NewFromInt(2))
	return price
}

func SyncCheckOrder(interval int, wg *sync.WaitGroup, ch chan error) {
	defer wg.Done()
	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		for range ticker.C {
			go SyncCheckOrderOnce()
		}
	}()
}

func SyncCheckOrderOnce() {
	limitOrderInfo := inits.SafeOrderInfo.Get()
	// 获取当前时间
	currentTime := time.Now().Unix()
	for _, v := range limitOrderInfo {
		//如果大于2秒，就撤单
		if currentTime-v.CreateOrderTime > 6 {
			err := tradgo.CanceleOrder(v.Symbol, v.ClientOrderID)
			if err != nil {
				logger.Errorf("撤单失败，symbol=%s,err=%s", v.Symbol, err)
				notice.SendDingTalk("[撤消开仓失败预警]\n\n交易对:" + v.Symbol + "\n\n撤单失败，err=" + err.Error())
			} else {
				//撤消成功后删除这个Key
				logger.Infof("撤单成功，symbol=%s,ClientOrderID=%s", v.Symbol, v.ClientOrderID)
				notice.SendDingTalk("[撤消开仓订单成功预警]\n\n交易对:" + v.Symbol + "\n\nClientOrderID=" + v.ClientOrderID)
			}
			inits.SafeOrderInfo.Delete(v.Symbol + v.ClientOrderID)
		}
	}

}
