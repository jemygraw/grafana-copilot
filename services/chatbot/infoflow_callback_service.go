package chatbot

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jemygraw/grafana-copilot/conf"
	"github.com/jemygraw/grafana-copilot/services/chatbot/infoflow"
	ernie "github.com/jemygraw/grafana-copilot/services/ernine"
	"github.com/jemygraw/grafana-copilot/services/grafana"
	"log/slog"
	"os"
	"strings"
	"text/template"
)

const (
	GrafanaCmd = "grafana"
)

type GrafanaCopilotContext struct {
	GrafanaDashboards string
	UserInput         string
}

func HandleUserInput(callbackBody *infoflow.CallbackBody) {
	// check whether triggered by slash command
	userCmd := callbackBody.Message.GetUserCommand()
	if userCmd == "" {
		// TODO support understanding by LLM later
		return
	}
	if userCmd == GrafanaCmd {
		// handle grafana dashboard matching
		suggestedDashboards, err := handleGrafanaCopilot(callbackBody)
		if err != nil || len(suggestedDashboards) == 0 {
			var errMsg string
			if err != nil {
				errMsg = fmt.Sprintf("Handle grafana copilot err: %s", err.Error())
			} else {
				errMsg = "没有找到匹配的仪表盘，请尝试其他问题"
			}
			slog.Error(errMsg)
			NotifyUserError(callbackBody, errMsg)
		} else {
			NotifyUserResult(callbackBody, suggestedDashboards)
		}
	}
	return
}

func handleGrafanaCopilot(callbackBody *infoflow.CallbackBody) (suggestedDashboards []grafana.Dashboard, err error) {
	ctx := context.Background()
	// collect user message
	userInput := callbackBody.Message.GetUserInput()
	if userInput == "" {
		// notify error
		err = fmt.Errorf("no user input")
		return
	}
	// list the grafana dashboard metas
	dashboardMetas, err := grafana.ListDashboardMeta("")
	if err != nil {
		// notify error
		err = fmt.Errorf("list grafana dashboard metas err: %v", err)
		return
	}
	// covert dashboard metas to markdown table
	markdownBuf := bytes.NewBuffer(nil)
	markdownBuf.WriteString("|Title|Path|")
	markdownBuf.WriteString("|---|---|")
	for _, dashboardMeta := range dashboardMetas {
		markdownBuf.WriteString("|")
		markdownBuf.WriteString(dashboardMeta.Title)
		markdownBuf.WriteString("|")
		markdownBuf.WriteString(dashboardMeta.URL)
		markdownBuf.WriteString("|")
	}
	// prepare render context
	renderCtx := GrafanaCopilotContext{
		GrafanaDashboards: markdownBuf.String(),
		UserInput:         userInput,
	}
	// prepare llm input
	llmInput, err := RenderTemplate("prompts/grafana_copilot_prompt.md", renderCtx)
	if err != nil {
		err = fmt.Errorf("render template err: %w", err)
		return
	}
	slog.Debug(fmt.Sprintf("llm input:\n %s", llmInput))
	// call openai to get resp
	llmOutput, err := ernie.GetErnieResponse(ctx, conf.AppConfig, llmInput)
	if err != nil {
		err = fmt.Errorf("get llm response err: %w", err)
		return
	}
	// return msg
	slog.Debug(fmt.Sprintf("llm output:\n %s", llmOutput))
	textOutput := ernie.GetResponseTextContent(llmOutput)
	textLines := strings.Split(textOutput, "\n")
	suggestedDashboards = make([]grafana.Dashboard, 0, 2)
	grafanaHost := strings.TrimSuffix(conf.AppConfig.GrafanaHost, "/")
	for _, line := range textLines {
		slog.Debug(fmt.Sprintf("get llm text line, %s", line))
		items := strings.SplitN(line, "=", 2)
		if len(items) != 2 {
			continue
		}
		title := items[0]
		path := items[1]
		suggestedDashboards = append(suggestedDashboards, grafana.Dashboard{
			Title: title,
			URL:   fmt.Sprintf("%s%s", grafanaHost, path),
		})
	}
	return
}

func RenderTemplate(promptPath string, renderCtx any) (msg string, err error) {
	grafanaPromptTemplate, err := os.ReadFile(promptPath)
	if err != nil {
		err = fmt.Errorf("read %s err: %w", promptPath, err)
		return
	}
	tpl, err := template.New(promptPath).Parse(string(grafanaPromptTemplate))
	if err != nil {
		err = fmt.Errorf("parse %s err: %w", promptPath, err)
		return
	}
	llmInputBuf := bytes.NewBuffer(nil)
	err = tpl.Execute(llmInputBuf, &renderCtx)
	if err != nil {
		err = fmt.Errorf("parse %s err: %w", promptPath, err)
		return
	}
	msg = llmInputBuf.String()
	return
}

func NotifyUserError(callbackBody *infoflow.CallbackBody, outputMsg string) {
	// send the reply
	client := infoflow.NewClient(&infoflow.Config{
		WebhookAddress: conf.AppConfig.InfoflowRobotWebhookAddress,
	})
	groupId := callbackBody.GroupId
	fromUserId := callbackBody.Message.Header.FromUserId
	options := infoflow.MessageOptions{AtUserIds: []string{fromUserId}}
	_, err := client.SendTextMessage([]int{groupId}, outputMsg, &options)
	if err != nil {
		slog.Error(fmt.Sprintf("send message error: %v", err))
	}
}

func NotifyUserResult(callbackBody *infoflow.CallbackBody, suggestedDashboards []grafana.Dashboard) {
	// send the reply
	client := infoflow.NewClient(&infoflow.Config{
		WebhookAddress: conf.AppConfig.InfoflowRobotWebhookAddress,
	})
	groupId := callbackBody.GroupId
	fromUserId := callbackBody.Message.Header.FromUserId
	body := make([]infoflow.MessageBody, 0, 2)
	body = append(body, infoflow.MessageBody{
		Type:    infoflow.MessageBodyTypeText,
		Content: "为您找到如下看板:\n",
	})
	// add the suggested kanban links
	for _, dashboard := range suggestedDashboards {
		body = append(body, infoflow.MessageBody{
			Type:    infoflow.MessageBodyTypeText,
			Content: fmt.Sprintf("%s: ", dashboard.Title),
		})
		body = append(body, infoflow.MessageBody{
			Type: infoflow.MessageBodyTypeLink,
			Href: dashboard.URL,
		})
	}
	// check the options
	options := infoflow.MessageOptions{AtUserIds: []string{fromUserId}}
	body = append(body, options.CreateAtBody())
	message := infoflow.Message{
		Header: infoflow.MessageHeader{ToId: []int{groupId}},
		Body:   body,
	}
	_, err := client.SendMessage(&message)
	if err != nil {
		slog.Error(fmt.Sprintf("send message error: %v", err))
	}
}
