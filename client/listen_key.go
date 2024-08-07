package client

import (
	"context"

	"go_code/myselfgo/define"
)

// GetListenKey gets listen key from binance
func GetListenKey(apiKey, secretKey string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), define.TimeoutBinanceAPI)
	defer cancel()

	listenKey, err := GetFutureClient(apiKey, secretKey).NewStartUserStreamService().Do(ctx)
	return listenKey, err
}

// ProlongListenKey prolongs the valid time of the listen key
func ProlongListenKey(apiKey, secretKey, listenKey string) error {
	ctx, cancel := context.WithTimeout(context.Background(), define.TimeoutBinanceAPI)
	defer cancel()

	err := GetFutureClient(apiKey, secretKey).NewKeepaliveUserStreamService().ListenKey(listenKey).Do(ctx)
	return err
}
