package openai_client

import "context"

type (
	OpenAIClient interface {
		ChatGenerate(ctx context.Context, prompt string) (string, error)
	}
)
