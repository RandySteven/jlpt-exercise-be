package openai_client

import (
	"context"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

type (
	openaiClient struct {
		client openai.Client
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

