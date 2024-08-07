package datastruct

import (
	"sync"
	"time"
)

var (
	// AllTradingSymbols stores all trading symbols
	BuySymbolsPairs  = &SafeRoot{L: new(sync.RWMutex), M: make(map[string][]OrderGroup)}
	SellSymbolsPairs = &SafeRoot{L: new(sync.RWMutex), M: make(map[string][]OrderGroup)}
)

// Root 定义整个JSON数据的根结构体
type Root struct {
	Symbol string       `json:"Symbol"`
	Data   []OrderGroup `json:"Data"`
}

// Order 定义订单的结构体
type Order struct {
	Symbol           string  `json:"symbol"`
	Tradid           string  `json:"tradid"`
	Side             string  `json:"side"`
	Type             string  `json:"type"`
	TimeInForce      string  `json:"timeInForce"`
	NewClientOrderID string  `json:"newClientOrderId"`
	OrderVolume      float64 `json:"ordervolume"` //成交价值，price*quantity
	Price            float64 `json:"price"`
	RatePrice        float64 `json:"rate_price"`
	Quantity         float64 `json:"quantity"`
	NewOrderRespType string  `json:"newOrderRespType"`
}

// OrderGroup 包含多个订单的结构体
type OrderGroup struct {
	S1 Order `json:"s1"`
	S2 Order `json:"s2"`
	S3 Order `json:"s3"`
}

type SafeRoot struct {
	L         *sync.RWMutex
	Timestamp int64                   // update timestamp
	M         map[string][]OrderGroup // key: symbol, value: price filter info
}

func (sm *SafeRoot) ReInit(m map[string][]OrderGroup) {
	sm.L.Lock()
	defer sm.L.Unlock()
	sm.Timestamp = time.Now().Unix()
	sm.M = m
}

func (sm *SafeRoot) Get() map[string][]OrderGroup {
	sm.L.RLock()
	defer sm.L.RUnlock()
	return sm.M

}

func (sm *SafeRoot) GetBySymbol(symbol string) ([]OrderGroup, bool) {
	sm.L.RLock()
	defer sm.L.RUnlock()
	OrderGroupInfo, ok := sm.M[symbol]
	if !ok {
		return []OrderGroup{}, false
	}

	return OrderGroupInfo, true
}
