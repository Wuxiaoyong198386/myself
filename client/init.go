package client

import (
	"crypto/tls"
	"net/http"
	"time"

	"go_code/myselfgo/inits"

	"github.com/open-binance/logger"
)

func Init(cfg *inits.ServerConfig) error {
	// init http connection pool
	InitHTTPConnPool(cfg.HTTPClient)
	return nil
}

// client链接池
var (
	CommonClient *HTTPClient
)

// 初始化client链接池
func InitHTTPConnPool(cfg inits.HTTPClientInfo) {
	cc := NewHTTPConnPool(cfg.Common)
	CommonClient = cc
	logger.Infof(inits.SuccessInitHttpPoolLogmsg, inits.SuccessInitHttpPool, cc.Timeout, cc.MaxConnsPerHost, cc.MaxIdleConnsPerHost, cc.API)
}

// 使用config创建http连接池
func NewHTTPConnPool(cfg inits.HTTPClientPoolConfig) *HTTPClient {
	client := NewHTTPClient(cfg.Timeout, cfg.MaxConnsPerHost, cfg.MaxIdleConnsPerHost)

	return &HTTPClient{
		Timeout:             cfg.Timeout,
		MaxConnsPerHost:     cfg.MaxConnsPerHost,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		API:                 cfg.API,
		Client:              client,
	}
}

// 创建超时的http连接池
// @timeout: unit: ms
// @maxConnsPerHost: 限制每个主机的连接总数，包括处于拨号活动和空闲状态的连接
// @maxIdleConnsPerHost: 控制每个主机要保持的最大空闲（保持活动）连接，默认使用2
func NewHTTPClient(timeout int, maxConnsPerHost, maxIdleConnsPerHost int) *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			MaxConnsPerHost:     maxConnsPerHost,                       // 设置了到每个主机的最大连接数。当达到这个限制时，将不会建立新的连接，直到一些连接关闭。如果设置为0，则表示没有限制
			MaxIdleConnsPerHost: maxIdleConnsPerHost,                   // 设置了每个主机的最大空闲连接数。空闲连接是那些当前没有被使用，但仍然保持打开状态的连接。默认情况下，Go 的 HTTP 客户端为每个主机保持2空闲连接
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true}, //客户端在TLS握手时将不会验证服务器的证书
		},
		Timeout: time.Duration(timeout) * time.Millisecond,
	}
	return client
}
