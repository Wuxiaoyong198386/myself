package datastruct

import (
	"sync"

	"go_code/myselfgo/define"
)

// SafeStrInt64Map defines a safe map whose type of key is string, and type of value is int64
type SafeStrInt64Map struct {
	L *sync.RWMutex
	M map[string]int64
}

func (sm *SafeStrInt64Map) Set(key string, value int64) {
	sm.L.Lock()
	defer sm.L.Unlock()
	if sm.M == nil {
		sm.M = make(map[string]int64)
	}
	sm.M[key] = value
}

// Del deletes the key from the map
// return true if succeed to delete the key
func (sm *SafeStrInt64Map) Del(key string) bool {
	sm.L.Unlock()
	defer sm.L.Unlock()

	if _, ok := sm.M[key]; !ok {
		return false
	}

	delete(sm.M, key)
	return true
}

// Drop 从 SafeStrInt64Map 中删除一个键值对
// 如果键不存在，则添加键值对
// 如果值大于已存在的值，则更新该键值对
// 如果删除成功，则返回 true，否则返回 false
func (ss *SafeStrInt64Map) Drop(key string, value int64) bool {
	ss.L.Lock()
	defer ss.L.Unlock()

	v, ok := ss.M[key]
	if !ok {
		ss.M[key] = value
		return false
	}

	if value > v {
		ss.M[key] = value
		return false
	}

	// drop the record
	return true
}

// SafeStrBoolMap defines a safe map whose type of key is string, and type of value is bool
type SafeStrBoolMap struct {
	L *sync.RWMutex
	M map[string]bool
}

// ReInit reinits the map
func (sbm *SafeStrBoolMap) ReInit(m map[string]bool) {
	sbm.L.Lock()
	defer sbm.L.Unlock()
	sbm.M = m
}

// Keys gets all keys in the map
func (sbm *SafeStrBoolMap) Keys() []string {
	sbm.L.RLock()
	defer sbm.L.RUnlock()
	keys := make([]string, 0, len(sbm.M))

	for key := range sbm.M {
		keys = append(keys, key)
	}
	return keys
}

// GetCnt gets count of elements in the map
func (sbm *SafeStrBoolMap) GetCnt() int {
	sbm.L.RLock()
	defer sbm.L.RUnlock()

	return len(sbm.M)
}

// Get gets the map
func (ss *SafeStrBoolMap) Get() map[string]bool {
	m := make(map[string]bool)
	ss.L.RLock()
	defer ss.L.RUnlock()
	for k, v := range ss.M {
		m[k] = v
	}
	return m
}

// IsExisted checks whether the key is existed in the map or not
func (sbm *SafeStrBoolMap) IsExisted(key string) bool {
	sbm.L.RLock()
	defer sbm.L.RUnlock()

	_, ok := sbm.M[key]
	return ok
}

// GetKeys returns all keys in the map
func (sbm *SafeStrBoolMap) GetKeys() []string {
	sbm.L.RLock()
	defer sbm.L.RUnlock()

	keys := make([]string, 0, len(sbm.M))
	for key := range sbm.M {
		keys = append(keys, key)
	}
	return keys
}

type SafeNetworkDelayMap struct {
	L        *sync.RWMutex
	Adaptive bool
	Index    int
	Cnt      int
	MaxCnt   int
	Host     string               // the host with min network delay
	Delay    float64              // network delay of the host
	M        map[string][]float64 // key: host, value: network delay of the host
}

// NewNetworkDelay creates a new object to store network delay
func NewNetworkDelay(adaptive bool, customHost string) *SafeNetworkDelayMap {
	maxCnt := define.MaxNetworkDelayCnt
	m := make(map[string][]float64)
	host := define.BinanceHost // ust it by default
	if !adaptive {
		host = customHost
		m[host] = make([]float64, maxCnt)
	} else {
		m[define.BinanceHost] = make([]float64, maxCnt)
		m[define.BinanceHost1] = make([]float64, maxCnt)
		m[define.BinanceHost2] = make([]float64, maxCnt)
		m[define.BinanceHost3] = make([]float64, maxCnt)
		m[define.BinanceHost4] = make([]float64, maxCnt)
	}

	return &SafeNetworkDelayMap{
		L:        new(sync.RWMutex),
		Adaptive: adaptive,
		Index:    0,
		Cnt:      0,
		MaxCnt:   maxCnt,
		Host:     host,
		Delay:    0,
		M:        m,
	}
}

// UpdateHostAndDelay updates the host and delay
func (sm *SafeNetworkDelayMap) UpdateHostAndDelay() {
	sm.L.Lock()
	defer sm.L.Unlock()

	cnt := sm.Cnt
	if cnt == 0 {
		return
	}

	if !sm.Adaptive {
		// only update the network delay
		delays, ok := sm.M[sm.Host]
		if !ok {
			return
		}
		delaySum := float64(0)
		for _, delay := range delays {
			delaySum += delay
		}
		sm.Delay = delaySum / float64(cnt)
	}

	var hostSelect string
	var minDelaySum float64
	for host, delays := range sm.M {
		delaySum := float64(0)
		for _, delay := range delays {
			delaySum += delay
		}
		hostSelect = host
		minDelaySum = delaySum
		break
	}

	for host, delays := range sm.M {
		if host == hostSelect {
			continue
		}
		delaySum := float64(0)
		for _, delay := range delays {
			delaySum += delay
		}

		if delaySum < minDelaySum {
			minDelaySum = delaySum
			hostSelect = host
		}
	}

	sm.Host = hostSelect
	sm.Delay = minDelaySum / float64(cnt)
}

// GetHostAndDelay gets the host and delay
func (sm *SafeNetworkDelayMap) GetHostAndDelay() (string, float64) {
	sm.L.RLock()
	defer sm.L.RUnlock()

	return sm.Host, sm.Delay
}

// GetHostAndGetAllHostAndDelayDelay gets all the host and delay
func (sm *SafeNetworkDelayMap) GetAllHostAndDelay() map[string][]float64 {
	sm.L.RLock()
	defer sm.L.RUnlock()

	m := make(map[string][]float64)
	for host, delays := range sm.M {
		newDelays := make([]float64, len(delays))
		copy(newDelays, delays)
		m[host] = newDelays
	}

	return m
}

// StoreDelay stores delay in the map
func (sm *SafeNetworkDelayMap) StoreDelay(m map[string]float64) {
	sm.L.Lock()
	defer sm.L.Unlock()

	if sm.Cnt < sm.MaxCnt {
		sm.Cnt++
	}

	index := sm.Index
	for host, delay := range m {
		if _, ok := sm.M[host]; !ok {
			sm.M[host] = make([]float64, sm.MaxCnt)
		}
		sm.M[host][index] = delay
	}
	sm.Index = (index + 1) % sm.MaxCnt
}
