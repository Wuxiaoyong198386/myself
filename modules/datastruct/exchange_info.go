package datastruct

import (
	"sync"
	"time"
)

type SafePriceFilterInfoMap struct {
	L         *sync.RWMutex
	Timestamp int64                     // update timestamp
	M         map[string]SpotFilterInfo // key: symbol, value: price filter info
}

type SpotFilterInfo struct {
	Price         PriceFilterInfo   `json:"price"`           // filter info for price
	LotSize       LotSizeInfo       `json:"lot_size"`        // lot size for limit order
	MarketLotSize MarketLotSizeInfo `json:"market_lot_size"` // lot size for market order
	Notional      NotionalInfo      `json:"notional"`        // lot size for market order
}

type PriceFilterInfo struct {
	MinPrice string `json:"min_price"`
	MaxPrice string `json:"max_price"`
	TickSize string `json:"tick_size"`
}

type LotSizeInfo struct {
	MinQty   string `json:"min_qty"`
	MaxQty   string `json:"max_qty"`
	StepSize string `json:"step_size"`
}

type MarketLotSizeInfo struct {
	MinQty   string `json:"min_qty"`
	MaxQty   string `json:"max_qty"`
	StepSize string `json:"step_size"`
}

type NotionalInfo struct {
	MinNotional      string `json:"min_notional"`
	ApplyMinToMarket string `json:"apply_min_to_market"`
	MaxNotional      string `json:"max_notional"`
}

func (sm *SafePriceFilterInfoMap) ReInit(m map[string]SpotFilterInfo) {
	sm.L.Lock()
	defer sm.L.Unlock()
	sm.Timestamp = time.Now().Unix()
	sm.M = m
}

func (sm *SafePriceFilterInfoMap) GetStepSizeBySymbol(symbol string) (string, bool) {
	sm.L.RLock()
	defer sm.L.RUnlock()
	priceFilterInfo, ok := sm.M[symbol]
	if !ok {
		return "", false
	}

	// use step size in lot size is ok
	return priceFilterInfo.LotSize.StepSize, true
}

func (sm *SafePriceFilterInfoMap) GetTickSizeBySymbol(symbol string) (string, bool) {
	sm.L.RLock()
	defer sm.L.RUnlock()
	priceFilterInfo, ok := sm.M[symbol]
	if !ok {
		return "", false
	}

	// use step size in lot size is ok
	return priceFilterInfo.Price.TickSize, true
}

// GetBySymbol gets price filter info with symbol
func (sm *SafePriceFilterInfoMap) GetBySymbol(symbol string) (SpotFilterInfo, bool) {
	sm.L.RLock()
	defer sm.L.RUnlock()
	priceFilterInfo, ok := sm.M[symbol]
	if !ok {
		return SpotFilterInfo{}, false
	}

	return priceFilterInfo, true
}

// Get gets the map
func (sm *SafePriceFilterInfoMap) Get() map[string]SpotFilterInfo {
	sm.L.RLock()
	defer sm.L.RUnlock()
	return sm.M
}

// GetCnt gets count of elements in the map
func (sm *SafePriceFilterInfoMap) GetCnt() int {
	sm.L.RLock()
	defer sm.L.RUnlock()

	return len(sm.M)
}

const (
	RateLimitTypeRequestWeight string = "REQUEST_WEIGHT"
)

// RateLimitInfo specifies rate limit info
type RateLimitInfo struct {
	RateLimitType string `json:"rateLimitType"`
	Interval      string `json:"interval"`
	IntervalNum   int64  `json:"intervalNum"`
	Limit         int64  `json:"limit"`
}
