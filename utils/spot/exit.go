package spot

import (
	"os"
	"time"

	"go_code/myselfgo/define"
	"go_code/myselfgo/inits"
	"go_code/myselfgo/utils"
	"go_code/myselfgo/utils/notice"

	json "github.com/json-iterator/go"
	"github.com/open-binance/logger"
)

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
		os.Exit(0)
	}
	logger.Infof("强制退出成功,message:%s", message)

	os.Exit(0)
}

/*
 * @description:
 * @fileName: exit.go
 * @author: vip120@126.com
 * @date: 2024-03-30 11:44:47
 */
func logAndNotice(first bool, message string) {
	dingTalkMsgList := make([]string, 0, 5)
	DisplayInfo(false, message)
	dingTalkMsgList = append(dingTalkMsgList, "stop")

	dingTalkMsg := inits.BuildDingTalkMsg()
	dingTalkMsgList = append(dingTalkMsgList, dingTalkMsg)

	returnRate, deltaValue := inits.CumulativeReturnRate.GetReturnRate()
	returnRateVIP9 := returnRate.Sub(define.Decimal1Point8)
	returnRateStr := returnRate.StringFixed(define.DecimalPlaces3)
	returnRateVIP9Str := returnRateVIP9.StringFixed(define.DecimalPlaces3)
	msg11 := utils.JoinStrWithSep(": ", "return rate", returnRateStr)
	msg12 := utils.JoinStrWithSep(": ", "VIP9", returnRateVIP9Str)
	returnRateMsg := utils.JoinStrWithSep("  ", msg11, msg12)
	dingTalkMsgList = append(dingTalkMsgList, returnRateMsg)

	msg21 := utils.JoinStrWithSep(": ", "delta value", deltaValue.StringFixed(define.DecimalPlaces3))
	cumulative := utils.Float64ToDecimal(inits.QuoteValueInfo.GetCumulativeQuoteValue())
	msg22 := utils.JoinStrWithSep(": ", "VIP9", cumulative.Mul(returnRateVIP9).Div(define.Decimal10000).StringFixed(define.DecimalPlaces3))
	deltaValueMsg := utils.JoinStrWithSep("  ", msg21, msg22)
	dingTalkMsgList = append(dingTalkMsgList, deltaValueMsg)

	exitReason := utils.JoinStrWithSep("", "exit reason: ", message)
	dingTalkMsgList = append(dingTalkMsgList, exitReason)
	msg := utils.JoinStrWithSep(define.SepDoubleEnter, dingTalkMsgList...)
	notice.StartStopSender.SendText(msg)
}

/*
 * @description: 显示重要信息
 * @fileName: exit.go
 * @author: vip120@126.com
 * @date: 2024-03-29 19:02:34
 */
func DisplayInfo(first bool, errMsg string) {
	withErr := errMsg != ""
	if withErr {
		sleepTime := inits.Config.Custom.SleepBeforeExit
		if sleepTime > 0 {
			logger.Infof("等待显示信息,duration=%ds", sleepTime)
			time.Sleep(time.Duration(sleepTime) * time.Second)
		}
		logger.Infof("开始显示信息。")
	}

	spotAccountInfo := inits.SpotAsset2AssetInfo.Get()
	sJSON, err := json.Marshal(spotAccountInfo)
	if err == nil {
		logger.Infof("显示现货账户信息:%s", string(sJSON))
	} else {
		logger.Errorf("Json格式化账户信息失败。失败原因:%s", err.Error())
	}

	fundingAccountInfo := inits.FundingAsset2AssetInfo.Get()
	fJSON, err := json.Marshal(fundingAccountInfo)
	if err == nil {
		logger.Infof("显示资金账户信息:%s", string(fJSON))
	} else {
		logger.Errorf("Json格式化资金账户失败。失败原因:%s", err.Error())
	}

	if withErr {
		logger.Infof("退出原因：%s", errMsg)
	}
}
