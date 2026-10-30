package gemini_client

import (
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type (
	geminiClient struct {
		client openai.Client
	}
)

func NewGeminiClient() *geminiClient {
	client := openai.NewClient(
		option.WithAPIKey("My API Key"), // defaults to os.LookupEnv("OPENAI_API_KEY")
	)
	return &geminiClient{
		client: client,
	}
}
