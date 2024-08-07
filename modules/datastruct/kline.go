package datastruct

import (
	"sync"

	"github.com/shopspring/decimal"
)

// 保存最近两根k线的boll信息
type KlineBollMap struct {
	L    *sync.RWMutex         `json:"-"`
	Boll map[string][]BollInfo `json:"boll"` //key：交易对
}
type BollInfo struct {
	Up decimal.Decimal `json:"up"`
	Ma decimal.Decimal `json:"ma"`
	Dn decimal.Decimal `json:"dn"`
}

func (b *KlineBollMap) GetBoll(symbol string) ([]BollInfo, bool) {
	b.L.RLock()
	defer b.L.RUnlock()
	binfo, ok := b.Boll[symbol]
	if !ok {
		return binfo, false
	}
	return binfo, true
}

func (b *KlineBollMap) SetBoll(symbol string, values BollInfo) {
	b.L.Lock()
	defer b.L.Unlock()
	//只保存三根k线
	if len(b.Boll[symbol]) >= 2 {
		BollInfo := b.Boll[symbol]
		b.Boll[symbol] = BollInfo[1:]
	}
	b.Boll[symbol] = append(b.Boll[symbol], values)
}

type KlineMacdMap struct {
	L *sync.RWMutex        `json:"-"`
	M map[string]KlineMacd `json:"m"` //key：交易对
}

type KlineMacd struct {
	MacdUpdateTime int64             `json:"macd_update_time"`
	Diff           []decimal.Decimal `json:"diff"`
	Eda            []decimal.Decimal `json:"dea"`
	Macd           []decimal.Decimal `json:"macd"`
}

func (sm *KlineMacdMap) ReInit(symbol string, m KlineMacd) {
	sm.L.Lock()
	defer sm.L.Unlock()
	sm.M[symbol] = m
}

func (sm *KlineMacdMap) GetBySymbol(symbol string) (KlineMacd, bool) {
	sm.L.RLock()
	defer sm.L.RUnlock()
	macd, ok := sm.M[symbol]
	if !ok {
		return macd, false
	}

	return macd, true
}

type SafeKlineMap struct {
	L *sync.RWMutex             `json:"-"`
	M map[string][]WsKlineEvent `json:"m"` //key：交易对
}

func (sm *SafeKlineMap) ReInit(symbol string, m []WsKlineEvent) {
	sm.L.Lock()
	defer sm.L.Unlock()
	sm.M[symbol] = m
}

func (sm *SafeKlineMap) GetBySymbol(symbol string) ([]WsKlineEvent, bool) {
	sm.L.RLock()
	defer sm.L.RUnlock()
	kline, ok := sm.M[symbol]
	if !ok {
		return kline, false
	}

	return kline, true
}

type WsKlineAmplitude struct {
	Symbol       string  `json:"s"`
	StartTime    string  `json:"StartTime"`
	EndTime      string  `json:"EndTime"`
	High         string  `json:"h"`
	Low          string  `json:"l"`
	Open         string  `json:"o"`
	Close        string  `json:"c"`
	QuoteVolume  string  `json:"q"`    //交易量
	Oscillation  float64 `json:"a"`    //振幅
	Amplitude    bool    `json:"A"`    //涨跌
	AmplitudeStr string  `json:"aStr"` //涨跌字符串
}

type WsKlineEvent struct {
	Event  string  `json:"e"`
	Time   int64   `json:"E"`
	Symbol string  `json:"s"`
	Kline  WsKline `json:"k"`
}

// WsKline define websocket kline
type WsKline struct {
	StartTime            int64  `json:"t"`
	EndTime              int64  `json:"T"`
	Symbol               string `json:"s"`
	Interval             string `json:"i"`
	FirstTradeID         int64  `json:"f"`
	LastTradeID          int64  `json:"L"`
	Open                 string `json:"o"`
	Close                string `json:"c"`
	High                 string `json:"h"`
	Low                  string `json:"l"`
	Volume               string `json:"v"`
	TradeNum             int64  `json:"n"`
	IsFinal              bool   `json:"x"`
	QuoteVolume          string `json:"q"`
	ActiveBuyVolume      string `json:"V"`
	ActiveBuyQuoteVolume string `json:"Q"`
}
