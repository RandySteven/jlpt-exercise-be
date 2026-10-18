package deepseek_client

import (
	"context"

	deepseek_go "github.com/cohesion-org/deepseek-go"
)

type (
	deepSeek struct {
		deepseekGoClient *deepseek_go.Client
	}

	DeepSeekConfig struct {
		APIKey string
	}
)

func NewDeepSeek(config *DeepSeekConfig) (*deepSeek, error) {
	deepseekGoClient := deepseek_go.NewClient(config.APIKey, "https://deepseek.com")
	return &deepSeek{
		deepseekGoClient: deepseekGoClient,
	}, nil
}

func (d *deepSeek) ChatGenerate(ctx context.Context, prompt string) (string, error) {
	response, err := d.deepseekGoClient.CreateChatCompletion(ctx, &deepseek_go.ChatCompletionRequest{
		Model: "deepseek-chat",
		Messages: []deepseek_go.ChatCompletionMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	})
	if err != nil {
		return "", err
	}
	return response.Choices[0].Message.Content, nil
}
