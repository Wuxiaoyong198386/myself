package datastruct

import (
	"sync"

	binance_connector "github.com/binance/binance-connector-go"
)

type WebsocketClientManager struct {
	Client chan *binance_connector.WebsocketAPIClient
	L      *sync.RWMutex
	Pool   *WebsocketClientManager
}

func NewConnectionPool(initialCap int) *WebsocketClientManager {
	return &WebsocketClientManager{
		Client: make(chan *binance_connector.WebsocketAPIClient, initialCap),
		L:      &sync.RWMutex{},
	}
}

func (s *WebsocketClientManager) SetPool(pool *WebsocketClientManager) {
	s.L.Lock()
	defer s.L.Unlock()
	s.Pool = pool
}

func (s *WebsocketClientManager) GetPool() (*WebsocketClientManager, error) {
	s.L.RLock()
	defer s.L.RUnlock()
	return s.Pool, nil
}

func (s *WebsocketClientManager) SetWsclient(conn *binance_connector.WebsocketAPIClient) {
	s.L.Lock()
	defer s.L.Unlock()
	s.Client <- conn
}

func (s *WebsocketClientManager) GetWsclient() *binance_connector.WebsocketAPIClient {
	// 加锁，确保并发安全
	s.L.RLock()
	defer s.L.RUnlock()
	// 检查连接池是否为空
	if len(s.Client) == 0 {
		return nil
	}
	// 将连接池中的第一个连接移除
	//p.connections = p.connections[1:]
	// 返回取出的连接
	return <-s.Client
}
