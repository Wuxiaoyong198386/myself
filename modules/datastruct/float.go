package datastruct

import (
	"sync"
)

// SafeFloat64 specifies a safe float64 object
type SafeFloat64 struct {
	L     *sync.RWMutex
	Value float64 // the value of float64
}

// Get gets the value
func (sf *SafeFloat64) Get() float64 {
	sf.L.RLock()
	defer sf.L.RUnlock()

	return sf.Value
}

// Set sets the value
func (sf *SafeFloat64) Set(value float64) {
	sf.L.Lock()
	defer sf.L.Unlock()

	sf.Value = value
}
