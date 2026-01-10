package infoflow

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	DefaultTimeout = time.Second * 10 // 10 seconds
)

type Config struct {
	Timeout        int    `json:"timeout"`
	WebhookAddress string `json:"webhookAddress"`
}

// Client is a robot client to send infoflow messages
type Client struct {
	httpClient     *http.Client
	WebhookAddress string
}

func NewClient(cfg *Config) *Client {
	timeout := DefaultTimeout
	if cfg.Timeout > 0 {
		timeout = time.Duration(cfg.Timeout) * time.Second
	}
	return &Client{
		httpClient:     &http.Client{Timeout: timeout},
		WebhookAddress: cfg.WebhookAddress,
	}
}

func NewClientWithHttpClient(cfg *Config, httpClient *http.Client) *Client {
	return &Client{
		httpClient:     httpClient,
		WebhookAddress: cfg.WebhookAddress,
	}
}

func (c *Client) SendTextMessage(groupIds []int, content string, options *MessageOptions) (data ExtraData, err error) {
	body := make([]MessageBody, 0, 2)
	body = append(body, MessageBody{
		Type:    MessageBodyTypeText,
		Content: content,
	})
	// check the options
	if options != nil && options.IsAtEnabled() {
		body = append(body, options.CreateAtBody())
	}
	message := Message{
		Header: MessageHeader{ToId: groupIds},
		Body:   body,
	}
	return c.sendMessage(&message)
}

func (c *Client) SendLinkMessage(groupIds []int, link string, options *MessageOptions) (data ExtraData, err error) {
	body := make([]MessageBody, 0, 2)
	body = append(body, MessageBody{
		Type: MessageBodyTypeLink,
		Href: link,
	})
	// check the options
	if options != nil && options.IsAtEnabled() {
		body = append(body, options.CreateAtBody())
	}
	message := Message{
		Header: MessageHeader{ToId: groupIds},
		Body:   body,
	}
	return c.sendMessage(&message)
}

func (c *Client) SendImageMessage(groupIds []int, imageBytes []byte) (data ExtraData, err error) {
	message := Message{
		Header: MessageHeader{ToId: groupIds},
		Body: []MessageBody{
			{
				Type:    MessageBodyTypeImage,
				Content: base64.StdEncoding.EncodeToString(imageBytes),
			},
		},
	}
	return c.sendMessage(&message)
}

func (c *Client) SendMarkdownMessage(groupIds []int, content string) (data ExtraData, err error) {
	message := Message{
		Header: MessageHeader{ToId: groupIds},
		Body: []MessageBody{
			{
				Type:    MessageBodyTypeMarkdown,
				Content: content,
			},
		},
	}
	return c.sendMessage(&message)
}

func (c *Client) sendMessage(message *Message) (data ExtraData, err error) {
	reqMethod := http.MethodPost
	reqBody, mErr := json.Marshal(&RequestBody{Message: *message})
	if mErr != nil {
		err = fmt.Errorf("marshal request body error, %s", mErr.Error())
		return
	}
	req, newErr := http.NewRequest(reqMethod, c.WebhookAddress, bytes.NewReader(reqBody))
	if newErr != nil {
		err = fmt.Errorf("create request error, %s", newErr.Error())
		return
	}
	// set content-type
	req.Header.Set("Content-Type", "application/json")
	// fire request
	resp, callErr := c.httpClient.Do(req)
	if callErr != nil {
		err = fmt.Errorf("get response error, %s", callErr.Error())
		return
	}
	defer resp.Body.Close()
	// check status code
	if resp.StatusCode != http.StatusOK {
		// discard response body to reuse underline tcp connections
		_, _ = io.Copy(ioutil.Discard, resp.Body)
		err = fmt.Errorf("get response error, %s", resp.Status)
		return
	}
	var responseBody ResponseBody
	decoder := json.NewDecoder(resp.Body)
	decErr := decoder.Decode(&responseBody)
	if decErr != nil {
		err = fmt.Errorf("parse response error, %s", decErr.Error())
		return
	}
	// check logic code
	if responseBody.ErrorCode != ErrNone {
		err = fmt.Errorf("call api error, %d: %s", responseBody.ErrorCode, responseBody.ErrorMessage)
		return
	}
	data = responseBody.Data
	return
}
