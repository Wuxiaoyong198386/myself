package client

import (
	"encoding/json"
	"os"
	"time"

	"go_code/myselfgo/inits"

	"github.com/open-binance/logger"
)

// GracefulExit tries to exit gracefully
func GracefulExit(message string) {
	inits.FlagGracefulExit.Set(true)
	maxTry := 66
	for i := 0; i < maxTry; i++ {
		if trading := inits.FlagTrade.GetStatus(); trading {
			logger.Infof("msg=尝试退出,message:%s", message)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		logger.Infof("成功退出,message:%s", message)
		//logAndNotice(false, message)
		os.Exit(0)
	}
	logger.Infof("强制退出成功,message:%s", message)

	os.Exit(0)
}

func DisplayInfo(first bool, errMsg string) {
	withErr := errMsg != ""
	if withErr {
		logger.Infof("请稍等以显示信息,等待时间=%ds", 11)
		time.Sleep(11 * time.Second)
		logger.Infof("开始显示信息")
	}

	spotAccountInfo := inits.SpotAsset2AssetInfo.Get()
	sJSON, err := json.Marshal(spotAccountInfo)
	if err == nil {
		logger.Infof("显示账户信息,first:%t,info:%s", first, string(sJSON))
	} else {
		logger.Errorf("格式化账户信息失败，失败原因:%s", err.Error())
	}

	totalValue := inits.AllAssetValueSum.Get()
	logger.Infof("显示账户总额,total_value:%v", totalValue)

	if withErr {
		logger.Infof("退出原因,err_msg:%s", errMsg)
	}

}
