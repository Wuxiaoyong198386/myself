package datastruct

import (
	"sync"

	"github.com/adshao/go-binance/v2"
)

type TimeInForceType string

type SafeWsOrderUpdate struct {
	L *sync.RWMutex                    `json:"-"`
	M map[string]binance.WsOrderUpdate `json:"m"` //key：交易对
}

func (sm *SafeWsOrderUpdate) ReInit(ClientOrderId string, data binance.WsOrderUpdate) {
	sm.L.Lock()
	defer sm.L.Unlock()
	sm.M[ClientOrderId] = data
}

func (sm *SafeWsOrderUpdate) Get() map[string]binance.WsOrderUpdate {
	sm.L.Lock()
	defer sm.L.Unlock()
	return sm.M
}

func (sm *SafeWsOrderUpdate) GetStateByClientOrderId(ClientOrderId string) string {
	sm.L.Lock()
	defer sm.L.Unlock()
	return sm.M[ClientOrderId].Status
}

// NewAllState 初始化allState结构体
type AllState struct {
	L *sync.RWMutex  `json:"-"`
	M map[string]int `json:"m"` //key：交易对

}

// SetState 更新指定pendingSymbol的状态
func (as *AllState) SetState(ps string, newState int) {
	as.L.Lock()
	defer as.L.Unlock()
	as.M[ps] = newState
}

// GetState
// 获取指定Symbol的状态
func (as *AllState) GetStateBypendingSymbol(ps string) (int, bool) {
	as.L.RLock()
	defer as.L.RUnlock()
	state, exists := as.M[ps]
	return state, exists
}
func (as *AllState) GetState() map[string]int {
	as.L.RLock()
	defer as.L.RUnlock()
	return as.M

}

