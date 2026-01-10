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
		outputMsg, err := handleGrafanaCopilot(callbackBody)
		if err != nil {
			outputMsg = fmt.Sprintf("Handle grafana copilot err: %s", err.Error())
			slog.Error(outputMsg)
		}
		NotifyUserResult(callbackBody, outputMsg)
	}
	return
}

func handleGrafanaCopilot(callbackBody *infoflow.CallbackBody) (msg string, err error) {
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
	markdownBuf.WriteString("|Title|URL|")
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
	// call openai to get resp
	llmOutput, err := ernie.GetErnieResponse(ctx, conf.AppConfig, llmInput)
	if err != nil {
		err = fmt.Errorf("get llm response err: %w", err)
		return
	}
	// return msg
	msg = llmOutput
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

func NotifyUserResult(callbackBody *infoflow.CallbackBody, outputMsg string) {
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
