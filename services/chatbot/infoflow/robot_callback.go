package infoflow

import (
	"bytes"
	"github.com/duoland/base/crypto/ecb"
	"github.com/duoland/base/hash"
	"strings"
)

type CallbackBody struct {
	EventType string          `json:"eventtype"`
	AgentId   int             `json:"agentid"`
	GroupId   int             `json:"groupid"`
	CorpId    string          `json:"corpid"`
	Message   CallbackMessage `json:"message"`
	Time      int64           `json:"time"`
}

type CallbackMessage struct {
	Header CallbackMessageHeader `json:"header"`
	Body   []CallbackMessageBody `json:"body"`
}

// GetUserInput 返回用户输入的文字信息
func (m CallbackMessage) GetUserInput() string {
	buf := bytes.NewBuffer(nil)
	for _, body := range m.Body {
		if body.Type == "LINK" && body.Label != "" {
			buf.WriteString(body.Label)
			buf.WriteString("\n")
		}
		if body.Type == "TEXT" && body.Content != "" {
			buf.WriteString(body.Content)
		}
	}
	return buf.String()
}

// GetUserCommand 返回用户触发的斜杠命令
func (m CallbackMessage) GetUserCommand() string {
	for _, body := range m.Body {
		if body.Type == "command" {
			return body.CommandName
		}
	}
	return ""
}

type CallbackMessageHeader struct {
	FromUserId    string            `json:"fromuserid"`
	ToId          int               `json:"toid"`
	ToType        string            `json:"totype"`
	MsgType       string            `json:"msgtype"`
	ClientMsgId   int64             `json:"clientmsgid"`
	MessageId     int64             `json:"messageid"`
	MsgSeqId      string            `json:"msgseqid"`
	At            CallbackMessageAt `json:"at"`
	Compatible    string            `json:"compatible"`
	OfflineNotify string            `json:"offlinenotify"`
	Extra         string            `json:"extra"`
	ServerTime    int64             `json:"servertime"`
	ClientTime    int64             `json:"clienttime"`
	UpdateTime    int64             `json:"updatetime"`
}

type CallbackMessageBody struct {
	// Type can be "TEXT","LINK","command"
	Type string `json:"type"`
	// CommandName is set when Type is "command"
	CommandName string `json:"commandname"`
	// Content is set when Type is "TEXT"
	Content string `json:"content"`
	// Label is set when Type is "LINK"
	Label       string `json:"label"`
	DownloadURL string `json:"downloadurl"`
	RobotId     int    `json:"robotid"`
	// UserID is the username who send the message to the robot
	UserID string `json:"userid"`
	Name   string `json:"name"`
}

type CallbackMessageAt struct {
	AtRobotIds []int    `json:"atrobotids"`
	AtUserIds  []string `json:"atuserids"`
}

// CalcInfoflowVerifySignature 计算如流机器人验证地址和发送消息时的签名
func CalcInfoflowVerifySignature(rn, timestamp, token string) string {
	srcStr := strings.Join([]string{rn, timestamp, token}, "")
	return hash.Md5HexString([]byte(srcStr))
}

func DecryptInfoflowMessage(encrytedMsg []byte, secret []byte) ([]byte, error) {
	return ecb.AESDecrypt(encrytedMsg, secret)
}

// PaddingBase64String 填充base64字符串，使其长度为4的倍数
func PaddingBase64String(src string) string {
	trailingCnt := len(src) % 4
	if trailingCnt > 0 {
		src += strings.Repeat("=", 4-trailingCnt)
	}
	return src
}
