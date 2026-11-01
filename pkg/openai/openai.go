package openai_client

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

type (
	openaiClient struct {
		client openai.Client
	}

	Turn struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	StreamChunk struct {
		Text string `json:"text"`
		Err  error  `json:"err"`
	}
)

func NewOpenAIClient() *openaiClient {
	client := openai.NewClient(
		option.WithAPIKey("My API Key"), // defaults to os.LookupEnv("OPENAI_API_KEY")
	)
	return &openaiClient{
		client: client,
	}
}

func (c *openaiClient) ChatGenerate(ctx context.Context, prompt string) (string, error) {
	resp, err := c.client.Responses.New(ctx, responses.ResponseNewParams{
		Model: "gpt-4o",
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(prompt),
		},
	})
	if err != nil {
		return "", err
	}
	return resp.OutputText(), nil
}

func (c *openaiClient) Stream(ctx context.Context, system string, history []Turn) (<-chan StreamChunk, error) {
	input := make(responses.ResponseInputParam, 0, len(history))
	for _, turn := range history {
		var role responses.EasyInputMessageRole
		switch turn.Role {
		case "user":
			role = responses.EasyInputMessageRoleUser
		case "assistant":
			role = responses.EasyInputMessageRoleAssistant
		default:
			return nil, fmt.Errorf("unknown turn role %q", turn.Role)
		}
		input = append(input, responses.ResponseInputItemParamOfMessage(turn.Content, role))
	}
	resp := c.client.Responses.NewStreaming(ctx, responses.ResponseNewParams{
		Model:        "gpt-4o",
		Instructions: openai.String(system),
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: input,
		},
	})
	out := make(chan StreamChunk)
	go func() {
		defer close(out)
		defer resp.Close()
		for resp.Next() {
			event := resp.Current()
			if event.Type != "response.output_text.delta" || event.Delta == "" {
				continue
			}
			select {
			case <-ctx.Done():
				return
			case out <- StreamChunk{Text: event.Delta}:
			}
		}
		if err := resp.Err(); err != nil && ctx.Err() == nil {
			select {
			case <-ctx.Done():
			case out <- StreamChunk{Err: err}:
			}
		}
	}()
	return out, nil
}
