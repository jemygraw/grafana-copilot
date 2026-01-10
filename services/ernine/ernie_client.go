package ernie

import (
	"context"
	"fmt"
	"github.com/jemygraw/grafana-copilot/conf"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func GetErnieResponse(ctx context.Context, appConfig *conf.Config, messages []llms.MessageContent) (llmOutput string, err error) {
	var client *openai.LLM
	client, err = openai.New(openai.WithBaseURL(appConfig.OpenAIAPIBase),
		openai.WithModel(appConfig.OpenAIModel),
		openai.WithToken(appConfig.OpenAIAPIKey),
	)
	if err != nil {
		err = fmt.Errorf("create openai client err: %v", err)
		return
	}
	llmResp, err := client.GenerateContent(ctx, messages, llms.WithTemperature(0.7))
	if err != nil {
		err = fmt.Errorf("call openai err: %v", err)
		return
	}
	llmOutput = llmResp.Choices[0].Content
	return
}
