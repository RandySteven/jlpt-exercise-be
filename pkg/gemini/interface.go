package gemini_client

import "context"

type (
	GeminiClient interface {
		ChatGenerate(ctx context.Context, prompt string) (string, error)
		Stream(ctx context.Context, system string, history []Turn) (<-chan StreamChunk, error)
	}
)
