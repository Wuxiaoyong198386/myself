package notice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go_code/myselfgo/inits"
	"io/ioutil"
	"net/http"
)

type DingTalkMessage struct {
	Msgtype string `json:"msgtype"` // 消息类型，这里以text为例
	Text    struct {
		Content string `json:"content"` // 文本消息内容
	} `json:"text"`
	// 如果有其他类型的消息，比如markdown、link等，可以添加相应的字段
}

// Send 是一个方法，用于设置DingTalkMessage的内容
func (d *DingTalkMessage) Send(content string) {
	d.Msgtype = "text"
	d.Text.Content = content
}

// SendDingTalk 发送文本消息到钉钉
func SendDingTalk(content string) error {
	if !inits.Config.Notice.DingTalk.Enable {
		return nil
	}
	dingTalkMessage := DingTalkMessage{}
	dingTalkMessage.Send(content)

	// 将DingTalkMessage实例序列化为JSON
	data, err := json.Marshal(dingTalkMessage)
	if err != nil {
		return err
	}

	// 发送POST请求到Webhook URL
	var webhook string
	if inits.Config.Symbol.Type == 2 {
		webhook = inits.Config.Notice.DingTalk.InfoLog1.Webhook
	} else {
		webhook = inits.Config.Notice.DingTalk.InfoLog2.Webhook
	}
	resp, err := http.Post(webhook, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 读取响应体
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// 你可以在这里打印或检查响应内容
	//fmt.Println("Response:", string(body))

	// 假设钉钉返回200状态码表示成功，但你可能需要根据实际响应来处理
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to send message to DingTalk, status code: %d", resp.StatusCode)
	}

	return nil
}
