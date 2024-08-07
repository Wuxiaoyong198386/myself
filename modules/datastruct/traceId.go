package datastruct

import (
	"fmt"
	"sync"
)

type SafeTraceId struct {
	L *sync.RWMutex     `json:"-"`
	T map[string]string `json:"t"`
}

func (s *SafeTraceId) SetValue(key string, value string) {
	s.L.Lock()
	defer s.L.Unlock()
	s.T[key] = value
}

func (s *SafeTraceId) GetValue(key string) string {
	s.L.Lock()
	defer s.L.Unlock()
	if trace_id, isExists := s.T[key]; isExists {
		return trace_id
	}
	return fmt.Errorf("no get value for key:%s", key).Error()

}
