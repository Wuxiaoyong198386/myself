package cron

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"go_code/myselfgo/client"
	"go_code/myselfgo/define"
	"go_code/myselfgo/inits"
	"go_code/myselfgo/modules/datastruct"
	"go_code/myselfgo/notice"
	"go_code/myselfgo/utils"

	"github.com/open-binance/logger"
	"github.com/shopspring/decimal"
)

/*
 * @description: 使用http同步最优惠的价格
 * @fileName: best_price.go
 * @author: vip120@126.com
 * @date: 2024-03-20 14:50:17
 * @interval 单位：毫秒，同步最优惠价格的间隔
 */

var AllKlineAmplitudeMap = make(map[string][]datastruct.WsKlineAmplitude)

type SomeKlineStruct struct {
	Symbol           string  `json:"symbol"`
	Probability      float64 `json:"probability"`
	SumQuoteVolume   float64 `json:"sumQuoteVolume"`   //成交量
	Oscillation      float64 `json:"oscillation"`      //振幅
	ProbabilityRatio string  `json:"probabilityRatio"` //相似度
	UpDown           float64 `json:"upDown"`           //涨跌幅度
}

var mu sync.Mutex

func syncSymbolUp(interval int, wg *sync.WaitGroup, ch chan error) {
	defer wg.Done()
	logger.Infof("msg=The interval for syncSymbolUp filter information||interval=%ds", interval)
	if err := SyncSymbolUpServer(); err != nil {
		ch <- err
		return
	}
	go func() {
		//创建一个周期性的定时器
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		//从定时器中获取时间
		for range ticker.C {
			SyncSymbolUpServer()
		}
	}()
}
func SyncSymbolUpServer() error {
	acount := inits.Config.Account
	apiKey := acount.ApiKey
	secretKey := acount.SecretKey
	ctx, cancel := context.WithTimeout(context.Background(), define.TimeoutBinanceAPI)
	defer cancel()
	var wg sync.WaitGroup
	if inits.Config.Symbol.Type == 1 { //现货
		exchangeInfo, err := client.GetClient(apiKey, secretKey).NewExchangeInfoService().Do(ctx)
		if err != nil {
			fmt.Println("获取现货交易对失败", err)
			return err
		}
		//如果没有获取到数据，抛出异常
		if exchangeInfo == nil {
			logger.Errorf("msg=exchange info is empty")
			return errors.New("msg=exchange info is empty")
		}
		symbols := []string{}
		for i := 0; i < len(exchangeInfo.Symbols); i++ {
			symbol := exchangeInfo.Symbols[i].Symbol
			if exchangeInfo.Symbols[i].Status != define.StatusTrading { //状态是否为trading
				continue
			}
			lastFour := symbol[len(symbol)-4:]
			if lastFour != "USDT" { // symbol后四位不等于"USDT"跳过
				continue
			}
			symbols = append(symbols, symbol)
		}
		// fmt.Println("现货交易对", symbols, len(symbols))
		logger.Infof("现货交易对:", symbols)
		for _, symbol := range symbols {
			wg.Add(1)
			go SyncKlinesClient(symbol, &wg)
		}
	} else if inits.Config.Symbol.Type == 2 { //合约
		symbols := []string{}
		exchangeInfo, err := client.GetFutureClient(apiKey, secretKey).NewExchangeInfoService().Do(ctx)
		if err != nil {
			fmt.Println("获取合约交易对失败", err)
			return err
		}
		//如果没有获取到数据，抛出异常
		if exchangeInfo == nil {
			logger.Errorf("msg=exchange info is empty")
			return errors.New("msg=exchange info is empty")
		}
		for i := 0; i < len(exchangeInfo.Symbols); i++ {
			symbol := exchangeInfo.Symbols[i].Symbol
			if exchangeInfo.Symbols[i].Status != define.StatusTrading { //状态是否为trading
				continue
			}
			lastFour := symbol[len(symbol)-4:]
			if lastFour != "USDT" { // symbol后四位不等于"USDT"跳过
				continue
			}
			symbols = append(symbols, symbol)
		}
		for _, val := range symbols {
			wg.Add(1)
			go SyncKlinesFutureClient(val, &wg)
		}
	} else {
		fmt.Println("请在配置文件中填写正确的交易类型，1表示现货交易，2表示合约交易")
	}
	wg.Wait()
	var KlineResult []SomeKlineStruct
	//[]string{"BTCUSDT", "BNBUSDT"}
	defineSmbol := inits.Config.Symbol.SymbolWhiteList
	KlineResultDesc := []string{}
	for key, klineList := range AllKlineAmplitudeMap {
		btcKlineList, exists := AllKlineAmplitudeMap["BTCUSDT"]
		if !exists || len(btcKlineList) != len(klineList) {
			fmt.Printf("Key %s btclen %d klen %d either doesn't exist or has a different length klineList compared to BTCUSDT\n", key, len(btcKlineList), len(klineList))
			continue // 如果 BTCUSDT 不存在或长度不匹配，则跳过当前 key
		}
		var equallyNum int                                   // 计数变量
		var sumQuoteVolume, oscillation, probability float64 //成交量
		for k, val := range klineList {
			//判读同一时间是否相同
			if val.Amplitude == AllKlineAmplitudeMap["BTCUSDT"][k].Amplitude {
				equallyNum++
			}
			sumQuoteVolume = sumQuoteVolume + StrToFloat64(val.QuoteVolume) //交易量
			oscillation = oscillation + val.Oscillation                     //振幅
		}
		keyOpen := StrToFloat64(klineList[0].Open)
		keyClose := StrToFloat64(klineList[len(klineList)-1].Close)
		//涨幅 (最后一根收盘价-第一根开盘价)/第一根开盘价
		upDown := (keyClose - keyOpen) / keyOpen
		probability = float64(equallyNum) / float64(len(klineList)) //相似度
		//百分比
		probabilityRatio := fmt.Sprintf("%.2f%%", probability*100)
		if keyClose < inits.Config.Order.Max_price { //最高价
			KlineResult = append(KlineResult, SomeKlineStruct{Symbol: key, UpDown: upDown, SumQuoteVolume: sumQuoteVolume, Oscillation: oscillation, Probability: probability, ProbabilityRatio: probabilityRatio})
		}
	}
	//涨幅排序
	sort.Slice(KlineResult, func(i, j int) bool {
		return KlineResult[i].UpDown > KlineResult[j].UpDown
	})
	for _, v := range KlineResult {
		KlineResultDesc = append(KlineResultDesc, v.Symbol)
	}
	rSmbol := append(defineSmbol, KlineResultDesc[10:inits.Config.Symbol.SymbolCount]...)
	//相似度排序
	// for _, v := range KlineResult {
	// 	if v.Probability > inits.Config.Symbol.ProbabilityMax { //大于百分之80
	// 		KlineResultDesc = append(KlineResultDesc, v.Symbol)
	// 	}
	// }

	//去重
	uniqueSymbol := uniqueSymbols(rSmbol)
	inits.Config.Symbol.SymbolWhiteList = uniqueSymbol
	time := utils.FormatTime()
	logger.Infof("时间：%v 本次交易对:%v\n", time, uniqueSymbol)
	fmt.Printf("symbol为：%v,刷新时间：%v\n", uniqueSymbol, time)
	notice.SendDingTalk(fmt.Sprintf("交易对预警：%v", strings.Join(uniqueSymbol, " ")))
	return nil
}

func uniqueSymbols(symbols []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range symbols {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
func SyncKlinesClient(symbol string, wg *sync.WaitGroup) error {
	defer wg.Done()
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
	startTime := truncated.Add(-31*time.Minute).Unix() * 1000
	// 前一分钟的结束时间（即当前分钟的第0秒，但通常用于表示一分钟的结束）
	endTime := truncated.Add(-1*time.Minute).Unix() * 1000
	kList, err := client.GetClient(apiKey, secretKey).NewKlinesService().Symbol(symbol).
		Interval(inits.Config.Kline.Kine_type).StartTime(startTime).EndTime(endTime).Do(ctx)
	if err != nil {
		logger.Errorf("获取数据失败：%v,error:%v", symbol, err)
	}
	mu.Lock()
	defer mu.Unlock()
	var klineAmplitude []datastruct.WsKlineAmplitude
	for _, kline := range kList { // 不需要索引index
		klineOpen, _ := decimal.NewFromString(kline.Open)
		klineClose, _ := decimal.NewFromString(kline.Close)
		diff := klineClose.Sub(klineOpen)
		// amplitudeStr, amplitude := getAmplitude(diff)
		_, amplitude := getAmplitude(diff)
		high := StrToFloat64(kline.High)    //最高价
		low := StrToFloat64(kline.Low)      //最低价
		close := StrToFloat64(kline.Close)  //收盘价
		oscillation := (high - low) / close //振幅
		klineAmplitude = append(klineAmplitude, datastruct.WsKlineAmplitude{
			Symbol:      symbol,      // 使用函数参数中的 symbol
			Oscillation: oscillation, //振幅
			Open:        kline.Open,
			Close:       kline.Close,
			QuoteVolume: kline.QuoteAssetVolume, //成交量
			Amplitude:   amplitude,              //ture涨
		})
	}
	// logger.Infof("K线结果：%+v\n\n", klineAmplitude)
	AllKlineAmplitudeMap[symbol] = append(AllKlineAmplitudeMap[symbol], klineAmplitude...)
	return err
}
func SyncKlinesFutureClient(symbol string, wg *sync.WaitGroup) error {
	defer wg.Done()
	acount := inits.Config.Account
	apiKey := acount.ApiKey
	secretKey := acount.SecretKey
	ctx, cancel := context.WithTimeout(context.Background(), define.TimeoutBinanceAPI)
	defer cancel()
	// 获取当前时间
	now := time.Now()
	// 截断到分钟（去掉秒和纳秒部分）
	truncated := now.Truncate(time.Hour)
	// 开始时间
	startTime := truncated.Add(-25*time.Hour).Unix() * 1000
	// 前一分钟的结束时间（即当前分钟的第0秒，但通常用于表示一分钟的结束）
	endTime := truncated.Add(-1*time.Hour).Unix() * 1000
	kList, err := client.GetFutureClient(apiKey, secretKey).NewKlinesService().Symbol(symbol).
		Interval(inits.Config.Kline.Kine_interval).StartTime(startTime).EndTime(endTime).Do(ctx)
	if err != nil {
		logger.Errorf("获取数据失败：%v,error:%v", symbol, err)
	}
	mu.Lock()
	defer mu.Unlock()
	var klineAmplitude []datastruct.WsKlineAmplitude
	for _, kline := range kList { // 不需要索引index
		klineOpen, _ := decimal.NewFromString(kline.Open)
		klineClose, _ := decimal.NewFromString(kline.Close)
		diff := klineClose.Sub(klineOpen)
		// amplitudeStr, amplitude := getAmplitude(diff)
		_, amplitude := getAmplitude(diff)
		high := StrToFloat64(kline.High)    //最高价
		low := StrToFloat64(kline.Low)      //最低价
		close := StrToFloat64(kline.Close)  //收盘价
		oscillation := (high - low) / close //振幅
		klineAmplitude = append(klineAmplitude, datastruct.WsKlineAmplitude{
			Symbol:      symbol, // 使用函数参数中的 symbol
			StartTime:   getFormatTime(kline.OpenTime),
			EndTime:     getFormatTime(kline.CloseTime),
			Oscillation: oscillation,            //振幅
			QuoteVolume: kline.QuoteAssetVolume, //成交量
			Open:        kline.Open,
			Close:       kline.Close,
			// AmplitudeStr: amplitudeStr,
			Amplitude: amplitude, //ture 涨
		})

	}
	AllKlineAmplitudeMap[symbol] = append(AllKlineAmplitudeMap[symbol], klineAmplitude...)
	return err
}
func getAmplitude(diff decimal.Decimal) (string, bool) {
	if diff.LessThan(decimal.Zero) {
		return "跌", false
	}
	return "涨", true
}

func getFormatTime(timestamp int64) string {
	return time.Unix(timestamp/1000, 0).Format("2006-01-02 15:04:05")
}
