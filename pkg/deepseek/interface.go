package deepseek_client

import "context"

type (
	DeepSeek interface {
		ChatGenerate(ctx context.Context, prompt string) (string, error)
	}
)
