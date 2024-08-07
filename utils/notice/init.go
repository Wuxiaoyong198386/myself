package notice

import (
	"go_code/myselfgo/inits"

	"github.com/open-binance/logger"
)

// all ding talk sender
var (
	StartStopSender *DingTalkSender
	InfoLogSender   *DingTalkSender
)

// InitHTTPConnPool initializes all ding talk sender
func Init(cfg inits.NoticeInfo) {
	dingTalkfg := cfg.DingTalk
	enable := dingTalkfg.Enable

	// start and stop sender
	sssCfg := dingTalkfg.StartAndStop
	sss := NewDingTalkSender(enable, sssCfg.Name, sssCfg.Webhook, sssCfg.Keyword) // sss is short for start stop sender
	StartStopSender = sss
	logger.Infof("msg=%s||enable=%t||name=%s||webhook=%s||keyword=%s",
		"succeed to init start and stop sender", enable, sssCfg.Name, sss.Webhook, sss.Keyword)

	var ifCfg inits.DingTalkConfig
	// error and warn sender
	if inits.Config.Symbol.Type == 1 {
		ifCfg = dingTalkfg.InfoLog1
	} else {
		ifCfg = dingTalkfg.InfoLog2
	}

	ils := NewDingTalkSender(enable, ifCfg.Name, ifCfg.Webhook, ifCfg.Keyword) // ils is short for info log sender
	InfoLogSender = ils
	logger.Infof("msg=%s||enable=%t||name=%s||webhook=%s||keyword=%s",
		"succeed to init error and warn sender", enable, ifCfg.Name, ils.Webhook, ils.Keyword)
}
