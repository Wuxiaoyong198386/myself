package datastruct

import (
	"sync"

	"github.com/adshao/go-binance/v2"
)

// SafeRateLimit specifies rate limit info
type SafeRateLimit struct {
	L         *sync.RWMutex       `json:"-"`
	RateLimit []binance.RateLimit `json:"rate_limit"`
}

// ReInit reinits the rate limit info
func (sl *SafeRateLimit) ReInit(rateLimits []binance.RateLimit) {
	sl.L.Lock()
	defer sl.L.Unlock()

	sl.RateLimit = rateLimits
}

// Get gets the rate limit info
// hint: shallow copy is ok
func (sl *SafeRateLimit) Get() []binance.RateLimit {
	sl.L.RLock()
	defer sl.L.RUnlock()

	return sl.RateLimit
}
