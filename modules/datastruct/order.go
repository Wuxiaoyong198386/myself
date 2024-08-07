package datastruct

import (
	"sync"

	"github.com/shopspring/decimal"
)

type OrderTongJi struct {
	L              *sync.RWMutex
	MoneySum       decimal.Decimal
	UCommissionSum decimal.Decimal
	OrderFail      decimal.Decimal
	DorderFail     decimal.Decimal
	KorderFail     decimal.Decimal
	DorderSuccess  decimal.Decimal
	KorderSuccess  decimal.Decimal
	OrderSuccess   decimal.Decimal
}

func (otj *OrderTongJi) WMoneySum(v decimal.Decimal) decimal.Decimal {
	otj.L.Lock()
	defer otj.L.Unlock()
	otj.MoneySum = otj.MoneySum.Add(v)
	return otj.MoneySum
}

func (otj *OrderTongJi) UsdtCommissionSum(v decimal.Decimal) decimal.Decimal {
	otj.L.Lock()
	defer otj.L.Unlock()
	otj.UCommissionSum = otj.UCommissionSum.Add(v)
	return otj.UCommissionSum
}

// NewOrderTongJi 创建一个新的OrderTongJi实例
// IncrementOrderFail 增加OrderFail计数
func (otj *OrderTongJi) WOrderFail() decimal.Decimal {
	otj.L.Lock()
	defer otj.L.Unlock()
	otj.OrderFail = otj.OrderFail.Add(decimal.NewFromInt(1))
	return otj.OrderFail
}
func (otj *OrderTongJi) GetOrderFail() decimal.Decimal {
	otj.L.Lock()
	defer otj.L.Unlock()
	return otj.OrderFail
}

// IncrementDorderFail 增加DorderFail计数
func (otj *OrderTongJi) WDorderFail() decimal.Decimal {
	otj.L.Lock()
	defer otj.L.Unlock()
	otj.DorderFail = otj.DorderFail.Add(decimal.NewFromInt(1))
	return otj.DorderFail
}
func (otj *OrderTongJi) GetDorderFail() decimal.Decimal {
	otj.L.Lock()
	defer otj.L.Unlock()
	return otj.DorderFail
}

// IncrementKorderFail 增加KorderFail计数
func (otj *OrderTongJi) WKorderFail() decimal.Decimal {
	otj.L.Lock()
	defer otj.L.Unlock()
	otj.KorderFail = otj.KorderFail.Add(decimal.NewFromInt(1))
	return otj.KorderFail
}

func (otj *OrderTongJi) GetKorderFail() decimal.Decimal {
	otj.L.Lock()
	defer otj.L.Unlock()
	return otj.KorderFail
}

// IncrementDorderSuccess 增加DorderSuccess计数
func (otj *OrderTongJi) WDorderSuccess() decimal.Decimal {
	otj.L.Lock()
	defer otj.L.Unlock()
	otj.DorderSuccess = otj.DorderSuccess.Add(decimal.NewFromInt(1))
	return otj.DorderSuccess
}

func (otj *OrderTongJi) GetDorderSuccess() decimal.Decimal {
	otj.L.Lock()
	defer otj.L.Unlock()
	return otj.DorderSuccess

}

// IncrementKorderSuccess 增加KorderSuccess计数
func (otj *OrderTongJi) WKorderSuccess() decimal.Decimal {
	otj.L.Lock()
	defer otj.L.Unlock()
	otj.KorderSuccess = otj.KorderSuccess.Add(decimal.NewFromInt(1))
	return otj.KorderSuccess
}
func (otj *OrderTongJi) GetKorderSuccess() decimal.Decimal {
	otj.L.Lock()
	defer otj.L.Unlock()
	return otj.KorderSuccess

}

// IncrementOrderSuccess 增加OrderSuccess计数
func (otj *OrderTongJi) WOrderSuccess() decimal.Decimal {
	otj.L.Lock()
	defer otj.L.Unlock()
	otj.OrderSuccess = otj.OrderSuccess.Add(decimal.NewFromInt(1))
	return otj.OrderSuccess
}

func (otj *OrderTongJi) GetOrderSuccess() decimal.Decimal {
	otj.L.Lock()
	defer otj.L.Unlock()
	return otj.OrderSuccess

}

// Root 定义整个JSON数据的根结构体
type OrderSymbols struct {
	Symbol          string `json:"Symbol"`
	CreateOrderTime int64  `json:"CreateOrderTime"`
	OrderStatus     string `json:"OrderStatus"`
	ClientOrderID   string `json:"clientOrderID"`
	OrderSide       string `json:"OrderSide"`
}

type SafeOrderInfo struct {
	L *sync.RWMutex
	M map[string]OrderSymbols // key: symbol, value: price filter info
}

func (sm *SafeOrderInfo) ReInit(key string, m OrderSymbols) {
	sm.L.Lock()
	defer sm.L.Unlock()
	sm.M[key] = m
}

func (sm *SafeOrderInfo) Get() map[string]OrderSymbols {
	sm.L.RLock()
	defer sm.L.RUnlock()
	return sm.M

}

func (sm *SafeOrderInfo) GetBySymbol(symbol string) (OrderSymbols, bool) {
	sm.L.RLock()
	defer sm.L.RUnlock()
	OrderInfo, ok := sm.M[symbol]
	if !ok {
		return OrderSymbols{}, false
	}

	return OrderInfo, true
}

func (s *SafeOrderInfo) Delete(key string) {
	s.L.Lock()
	defer s.L.Unlock()
	delete(s.M, key)
}
