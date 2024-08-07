package notice

import (
	"fmt"
	"strings"

	"go_code/myselfgo/client"
	"go_code/myselfgo/define"
	"go_code/myselfgo/utils"

	json "github.com/json-iterator/go"
	"github.com/open-binance/logger"
)

const (
	DingTalkMsgTypeText    string = "text"
	DingTalkErrCodeSucceed int    = 0
)

type ReqDingTalk struct {
	Msgtype string   `json:"msgtype,omitempty"`
	At      AtInfo   `json:"at,omitempty"`
	Text    TextInfo `json:"text,omitempty"`
}

type AtInfo struct {
	IsAtAll   bool     `json:"isAtAll,omitempty"`
	AtMobiles []string `json:"atMobiles,omitempty"`
	AtUserIDs []string `json:"atUserIds,omitempty"`
}

type TextInfo struct {
	Content string `json:"content,omitempty"`
}

type ResDingTalk struct {
	ErrCode int    `json:"errcode"` // 0 means success
	ErrMsg  string `json:"errmsg"`
}

type DingTalkSender struct {
	Enable  bool   `json:"enable"`
	Name    string `json:"name"`
	Webhook string `json:"webhook"`
	Keyword string `json:"keyword"`
}

func NewDingTalkSender(enable bool, name, webhook, keyword string) *DingTalkSender {
	sender := &DingTalkSender{
		Enable:  enable,
		Name:    name,
		Webhook: webhook,
		Keyword: keyword,
	}

	return sender
}

func (ns *DingTalkSender) SendText(content string) error {
	if ns == nil {
		return fmt.Errorf("sender of ding talk is nil")
	}

	if !ns.Enable {
		return nil
	}

	if ns.Keyword != "" {
		content = utils.JoinStrWithSep(define.SepDoubleEnter, utils.GetChinaStandTimeStr(), ns.Keyword, content)
	}

	atInfo := AtInfo{
		// IsAtAll:   true,
		AtMobiles: nil,
		AtUserIDs: nil,
	}
	textInfo := TextInfo{
		Content: content,
	}
	req := ReqDingTalk{
		Msgtype: DingTalkMsgTypeText,
		At:      atInfo,
		Text:    textInfo,
	}

	reqJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	respBody, _, _, err := client.CommonClient.PostJSON(ns.Webhook, reqJSON)
	if err != nil {
		return err
	}

	var res ResDingTalk
	if err := json.Unmarshal(respBody, &res); err != nil {
		return err
	}
	if res.ErrCode != DingTalkErrCodeSucceed {
		return fmt.Errorf("error code from ding talk: %d, error message: %s", res.ErrCode, res.ErrMsg)
	}

	if strings.Contains(content, "stop") {
		logger.Infof("msg=succeed to send stop ding talk message||name=%s", ns.Name)
	}

	return nil
}
