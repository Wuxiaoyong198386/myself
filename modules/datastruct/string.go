package datastruct

import (
	"sync"
)

// SafeString defines a safe string object
type SafeString struct {
	L     *sync.RWMutex
	Value string
}

// Set sets the value
func (ss *SafeString) Set(value string) {
	ss.L.Lock()
	defer ss.L.Unlock()
	ss.Value = value
}

// Get gets the value
func (ss *SafeString) Get() string {
	ss.L.RLock()
	defer ss.L.RUnlock()
	return ss.Value
}
