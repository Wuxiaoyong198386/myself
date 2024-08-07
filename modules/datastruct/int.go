package datastruct

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"sync"
)

// SafeInt specifies a safe int object
type SafeInt struct {
	L     *sync.RWMutex
	Value int // the value of int
}

// Get gets the value
func (si *SafeInt) Get() int {
	si.L.RLock()
	defer si.L.RUnlock()

	return si.Value
}

// Set sets the value
func (si *SafeInt) Set(value int) {
	si.L.Lock()
	defer si.L.Unlock()

	si.Value = value
}

// Incr increases the value by 1
func (si *SafeInt) Incr() int {
	si.L.Lock()
	defer si.L.Unlock()

	if si.Value >= math.MaxInt {
		si.Value = 0
	}

	si.Value = si.Value + 1
	return si.Value
}

func (si *SafeInt) UpdateUsedWeight1m(header http.Header) error {
	valueStr := header.Get("X-Mbx-Used-Weight-1m")

	// do nothing
	if valueStr == "" {
		return nil
	}

	valueInt, err := strconv.Atoi(valueStr)
	if err != nil {
		return fmt.Errorf("failed to parse as int, err: %s", err.Error())
	}

	// update
	si.L.Lock()
	defer si.L.Unlock()
	si.Value = valueInt
	return nil
}

func (si *SafeInt) UpdateOrderCount10s(header http.Header) error {
	valueStr := header.Get("X-Mbx-Order-Count-10s")

	// do nothing
	if valueStr == "" {
		return nil
	}

	valueInt, err := strconv.Atoi(valueStr)
	if err != nil {
		return fmt.Errorf("failed to parse as int, err: %s", err.Error())
	}

	// update
	si.L.Lock()
	defer si.L.Unlock()
	si.Value = valueInt
	return nil
}

func (si *SafeInt) UpdateOrderCount1d(header http.Header) error {
	valueStr := header.Get("X-Mbx-Order-Count-1d")

	// do nothing
	if valueStr == "" {
		return nil
	}

	valueInt, err := strconv.Atoi(valueStr)
	if err != nil {
		return fmt.Errorf("failed to parse as int, err: %s", err.Error())
	}

	// update
	si.L.Lock()
	defer si.L.Unlock()
	si.Value = valueInt
	return nil
}

// SafeInt64 specifies a safe int object
type SafeInt64 struct {
	L     *sync.RWMutex
	Value int64 // the value of int
}

// Get gets the value
func (si *SafeInt64) Get() int64 {
	si.L.RLock()
	defer si.L.RUnlock()

	return si.Value
}

// Set sets the value
func (si *SafeInt64) Set(value int64) {
	si.L.Lock()
	defer si.L.Unlock()

	si.Value = value
}
