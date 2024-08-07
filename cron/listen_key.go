package cron

import (
	"fmt"
	"sync"
	"time"

	"go_code/myselfgo/client"
	"go_code/myselfgo/inits"

	"github.com/open-binance/logger"
)

/*
 * @description: 延时密钥
 * @fileName: listen_key.go
 * @author: vip120@126.com
 * @date: 2024-03-23 09:28:17
 */
func syncListenKey(interval int, wg *sync.WaitGroup, ch chan error) {
	defer wg.Done()

	apiKey := inits.Config.Account.ApiKey
	secretKey := inits.Config.Account.SecretKey

	listenKey, err := getListenKey(apiKey, secretKey)
	if err != nil {
		logger.Errorf(inits.ErrorListenkey, inits.ErrorListenkeyMsg, apiKey, secretKey, err.Error())
		ch <- err
		return
	}
	inits.ListenKey.Set(listenKey)
	logger.Infof(inits.SuccessListenkey, inits.SuccessListenkeyMsg, apiKey, secretKey, listenKey)
	go func() {
		logger.Infof(inits.MsgProlongListenKey, interval)
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		for range ticker.C {
			go prolongListenKeyOnce(apiKey, secretKey, listenKey)
		}
	}()
}

/*
 * @description: 定时获取密钥
 * @fileName: listen_key.go
 * @author: vip120@126.com
 * @date: 2024-03-23 09:29:58
 */
func prolongListenKeyOnce(apiKey, secretKey, listenKey string) {
	if err := client.ProlongListenKey(apiKey, secretKey, listenKey); err != nil {
		logger.Errorf(inits.ErrorListenkeyDelayed, listenKey, err.Error())
		return
	}
	logger.Infof(inits.SuccessListenkeyDelayed, listenKey)
}

/*
 * @description: 重试获取密钥
 * @fileName: listen_key.go
 * @author: vip120@126.com
 * @date: 2024-03-23 09:28:59
 */
func getListenKey(apiKey, secretKey string) (string, error) {
	maxCnt := 10
	for i := 0; i < maxCnt; i++ {
		listenKey, err := client.GetListenKey(apiKey, secretKey)
		if err != nil {
			logger.Errorf(inits.ErrorListenkey, inits.ErrorListenkeyMsg, apiKey, secretKey, err.Error())
			time.Sleep(1 * time.Second)
			continue
		}
		return listenKey, nil
	}

	return "", fmt.Errorf(inits.ErrorGetListenTime, maxCnt)
}
