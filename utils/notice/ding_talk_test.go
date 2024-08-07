package notice

import (
	"fmt"
	"os"
	"testing"

	"go_code/myselfgo/client"
	"go_code/myselfgo/inits"

	"github.com/open-binance/logger"
)

func init() {
	// load service config
	if err := inits.LoadConfig("../../conf/cfg.test.yaml"); err != nil {
		fmt.Printf("msg=failed to load config||err=%s\n", err.Error())
		os.Exit(1)
	}

	// init logger
	logCfg := logger.NewDefaultCfg()
	// logCfg.Level = "debug"
	if err := logger.Init(logCfg); err != nil {
		fmt.Printf("msg=failed to init logger||err=%s\n", err.Error())
		os.Exit(1)
	}
	defer logger.Sync() // flushes buffer, called in main function is reasonable

	// init http connection pool
	client.InitHTTPConnPool(inits.Config.HTTPClient)
}

func TestDingTalkSenderSendText(t *testing.T) {
	webhook := "https://oapi.dingtalk.com/robot/send?access_token=f2cdcfc62374174397ae0edd8d3e91f2ec87492d2dbd113b696ad45ec76e8071"
	sender := NewDingTalkSender(true, "golang", webhook, "WhoAreYou")
	err := sender.SendText("who are you")
	if err != nil {
		t.Errorf("msg=failed to send ding talk text message||err=%s", err.Error())
		return
	}
}
