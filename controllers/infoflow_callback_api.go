package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/jemygraw/grafana-copilot/conf"
	"github.com/jemygraw/grafana-copilot/services/chatbot"
	"github.com/jemygraw/grafana-copilot/services/chatbot/infoflow"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

/*
ReceiveInfoflowRobotMessage 如流机器人消息回调接口，参考文档：https://qy.baidu.com/doc/index.html#/inner_serverapi/robot
此接口会接收到两种类型的请求，一种是回调地址验证请求，另一种是正常的机器人消息请求。
1. 回调地址验证请求的 content-type 是 www-form-urlencoded，参数通过 body 传递；
2. 机器人消息请求的参数通过 query string 传递，消息内容通过 body 传递；
*/
func ReceiveInfoflowRobotMessage(resp http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		// 回调地址验证请求，此处可以返回固定的字符串作为验证结果。
		handleVerify(resp, req)
	} else {
		// 机器人消息请求，此处可以处理接收到的消息。
		handleMessage(resp, req)
	}
}

// handleVerify 处理回调地址验证请求。
// 参数通过 POST Body 传递，例如：
// signature=3c36ec74d3c84d8f6bbd3a92a97fda37&rn=688087768&echostr=61121&timestamp=1722513394
// 验证通过之后, 提取 echostr 参数, 并返回这个固定的字符串作为响应体内容。
func handleVerify(resp http.ResponseWriter, req *http.Request) {
	_ = req.ParseForm()
	signature := req.Form.Get("signature")
	rn := req.Form.Get("rn")
	echostr := req.Form.Get("echostr")
	timestamp := req.Form.Get("timestamp")
	token := conf.AppConfig.InfoflowRobotToken
	localSignature := infoflow.CalcInfoflowVerifySignature(rn, timestamp, token)
	if localSignature != signature {
		slog.Error("signature not match")
		resp.WriteHeader(http.StatusUnauthorized)
		return
	}
	resp.WriteHeader(http.StatusOK)
	_, _ = resp.Write([]byte(echostr))
}

// handleMessage 处理机器人消息请求。
// 消息内容通过 POST Body 传递, 需要通过 aes 解密后使用.
func handleMessage(resp http.ResponseWriter, req *http.Request) {
	msgBodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		slog.Error(fmt.Sprintf("read body err: %v", err))
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	msgBody := infoflow.PaddingBase64String(string(msgBodyBytes))
	encryptedBytes, err := base64.URLEncoding.DecodeString(msgBody)
	if err != nil {
		slog.Error(fmt.Sprintf("invalid base64 err: %v", err))
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	aesKey := fmt.Sprintf("%s==", conf.AppConfig.InfoflowRobotEncodingAESKey)
	secret, err := base64.StdEncoding.DecodeString(aesKey)
	if err != nil {
		slog.Error(fmt.Sprintf("invalid aesKey err: %v", err))
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	srcMsgBytes, err := infoflow.DecryptInfoflowMessage(encryptedBytes, secret)
	if err != nil {
		slog.Error(fmt.Sprintf("decrypt message err: %v", err))
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	// parse src message
	slog.Debug(fmt.Sprintf("src message: %s", string(srcMsgBytes)))
	var callbackBody infoflow.CallbackBody
	err = json.Unmarshal(srcMsgBytes, &callbackBody)
	if err != nil {
		slog.Error(fmt.Sprintf("parse src message err: %v", err))
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	// run the async process and notify user when finished
	go chatbot.HandleUserInput(&callbackBody)
}
