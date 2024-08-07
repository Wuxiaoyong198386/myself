package client

import (
	"context"

	"go_code/myselfgo/define"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
)

// GetClient gets binance client instance with API key and secret key.
// @apiKey: api key from binance
// @secretKey: secret key from binance
func GetClient(apiKey, secretKey string) *binance.Client {
	return binance.NewClient(apiKey, secretKey)
}

// GetFutureClient gets binance future client instance with API key and secret key.
// @apiKey: api key from binance
// @secretKey: secret key from binance
func GetFutureClient(apiKey, secretKey string) *futures.Client {
	return futures.NewClient(apiKey, secretKey)
}

// GetServerTime 获取服务器时间
// 参数：
// apiKey: string类型，Binance API的密钥
// secretKey: string类型，Binance API的密钥对应的私钥
//
// 返回值：
// int64类型，表示服务器当前时间的毫秒时间戳
// error类型，如果请求过程中出现错误，则返回相应的错误信息
func GetServerTime(apiKey, secretKey string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), define.TimeoutBinanceAPI)
	defer cancel()

	return GetFutureClient(apiKey, secretKey).NewServerTimeService().Do(ctx)
}
