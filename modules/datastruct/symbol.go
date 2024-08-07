package datastruct

import (
	"fmt"
	"sync"

	"github.com/adshao/go-binance/v2"
)

var (
	// AllTradingSymbols stores all trading symbols
	AllTradingSymbols       = &SafeSymbolInfoMap{L: new(sync.RWMutex), M: make(map[string]SymbolInfo)}
	AllSymbolInfoPairs      = &SymbolInfoPairsMap{L: new(sync.RWMutex), M: make(map[string][]map[string][]SymbolInfo)}
	AllSymbolInfoPairs_Bool = &SymbolInfoPairsBoolMap{L: new(sync.RWMutex), M: make(map[string]bool)}
)

// SymbolPair specifies a symbol pair for auto trading
type SymbolPair struct {
	S1 SymbolInfo `json:"s1"`
	S2 SymbolInfo `json:"s2"`
	S3 SymbolInfo `json:"s3"`
}

type SafeSymbolInfoMap struct {
	L *sync.RWMutex
	M map[string]SymbolInfo // key: symbol, value: symbol info
}

func (si *SafeSymbolInfoMap) ReInit(m map[string]SymbolInfo) {
	si.L.Lock()
	defer si.L.Unlock()
	si.M = m
}

func (si *SafeSymbolInfoMap) GetBaseAndQuote(symbol string) (string, string, error) {
	si.L.RLock()
	defer si.L.RUnlock()
	info, ok := si.M[symbol]
	if !ok {
		return "", "", fmt.Errorf("unknow symbol '%s'", symbol)
	}
	return info.Base, info.Quote, nil
}

func (si *SafeSymbolInfoMap) Get() map[string]SymbolInfo {
	m := make(map[string]SymbolInfo)
	si.L.RLock()
	defer si.L.RUnlock()
	for symbol, symbolInfo := range si.M {
		m[symbol] = symbolInfo
	}
	return m
}

func (si *SafeSymbolInfoMap) GetBySymbol(symbol string) (SymbolInfo, bool) {
	si.L.RLock()
	defer si.L.RUnlock()
	info, ok := si.M[symbol]
	return info, ok
}

func (si *SafeSymbolInfoMap) IsExisted(symbol string) bool {
	si.L.RLock()
	defer si.L.RUnlock()
	_, ok := si.M[symbol]
	return ok
}

// SymbolInfo specifies symbol info
type SymbolInfo struct {
	Symbol string `json:"s"`     // symbol
	Base   string `json:"base"`  // base asset
	Quote  string `json:"quote"` // quote asset
}

type SymbolInfoBase struct {
	Base  string `json:"base"`  // base asset
	Quote string `json:"quote"` // quote asset
}

type SymbolInfoPairs struct {
	Symbol     string         `json:"s"` // symbol
	SymbolInfo [][]SymbolInfo `json:"symbol_info"`
}

type SymbolInfoPairsMap struct {
	L *sync.RWMutex
	M map[string][]map[string][]SymbolInfo // key: base asset
}

func (sm *SymbolInfoPairsMap) ReInit(m map[string][]map[string][]SymbolInfo) {
	sm.L.Lock()
	defer sm.L.Unlock()
	sm.M = m
}

func (sm *SymbolInfoPairsMap) Get() map[string][]map[string][]SymbolInfo {
	sm.L.RLock()
	defer sm.L.RUnlock()
	return sm.M
}

type SymbolInfoPairsBoolMap struct {
	L *sync.RWMutex
	M map[string]bool // key: base asset
}

func (sm *SymbolInfoPairsBoolMap) ReInit(m map[string]bool) {
	sm.L.Lock()
	defer sm.L.Unlock()
	sm.M = m
}

func (sm *SymbolInfoPairsBoolMap) Get() map[string]bool {
	sm.L.RLock()
	defer sm.L.RUnlock()
	return sm.M
}

func (sm *SymbolInfoPairsBoolMap) GetBySymbol(key string) bool {
	sm.L.RLock()
	defer sm.L.RUnlock()
	m := sm.M[key]
	if m {
		return true
	} else {
		return false
	}
}

type SymbolTradeInfo struct {
	UpdateID int64            `json:"u_id"`
	Symbol   string           `json:"symbol"`
	Base     string           `json:"-"`
	Quote    string           `json:"-"`
	SideType binance.SideType `json:"side_type"`
	Price    float64          `json:"price"`
	Quantity float64          `json:"quantity"`
	Value    float64          `json:"market_value"`
}

// SafeSymbolPairMap specifies a safe map for symbol pair
type SafeSymbolPairMap struct {
	L *sync.RWMutex
	M map[string][]SymbolPair // key: base asset
}

// ReInit reinits the map
func (sm *SafeSymbolPairMap) ReInit(m map[string][]SymbolPair) {
	sm.L.Lock()
	defer sm.L.Unlock()
	sm.M = m
}

// Get gets the map
func (sm *SafeSymbolPairMap) Get() map[string][]SymbolPair {
	sm.L.RLock()
	defer sm.L.RUnlock()
	return sm.M
}

// GetByKey gets symbol pairs with key
func (sm *SafeSymbolPairMap) GetByKey(baseAsset string) ([]SymbolPair, bool) {
	sm.L.RLock()
	defer sm.L.RUnlock()

	symbolPairs, ok := sm.M[baseAsset]
	return symbolPairs, ok
}

// GetSymbolPairs get all the symbol pairs in the map
func (sm *SafeSymbolPairMap) GetSymbolPairs() []SymbolPair {
	var symbolPairs []SymbolPair
	sm.L.RLock()
	defer sm.L.RUnlock()

	for _, symbolPairList := range sm.M {
		symbolPairs = append(symbolPairs, symbolPairList...)
	}

	return symbolPairs
}
