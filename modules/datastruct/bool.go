package datastruct

import (
	"sync"
)

type AutoOrderTraceFlag struct {
	L *sync.RWMutex
	M map[string]bool
}

func (aotf *AutoOrderTraceFlag) Set(traceID string, v bool) {
	aotf.L.Lock()
	defer aotf.L.Unlock()
	if v {
		aotf.M[traceID] = true
	}
}

func (aotf *AutoOrderTraceFlag) Get(traceID string) bool {
	aotf.L.RLock()
	defer aotf.L.RUnlock()
	return aotf.M[traceID]
}

type AutoTradeFlag struct {
	L       *sync.RWMutex
	Trading bool
	TraceID string
}

// TryDoing sets the value of Trading as true
// return true means the value if trading is false before assignment
func (atf *AutoTradeFlag) Doing(traceID string) bool {
	atf.L.Lock()
	defer atf.L.Unlock()
	//第一次进来，Trading为false，可以进行交易
	if atf.Trading {
		return false
	}
	atf.TraceID = traceID
	atf.Trading = true

	return true
}

func (atf *AutoTradeFlag) Done() {
	atf.L.Lock()
	defer atf.L.Unlock()
	atf.Trading = false
}

func (atf *AutoTradeFlag) GetStatus() bool {
	atf.L.RLock()
	defer atf.L.RUnlock()
	return atf.Trading
}

func (atf *AutoTradeFlag) GetTraceID() string {
	atf.L.RLock()
	defer atf.L.RUnlock()
	return atf.TraceID
}

// SafeBool specifies a safe bool
type SafeBool struct {
	L     *sync.RWMutex
	Value bool
}

func (sb *SafeBool) Set(v bool) {
	sb.L.Lock()
	defer sb.L.Unlock()

	sb.Value = v
}

func (sb *SafeBool) Get() bool {
	sb.L.Lock()
	defer sb.L.Unlock()

	return sb.Value
}
