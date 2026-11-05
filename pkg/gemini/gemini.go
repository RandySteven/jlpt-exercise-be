package gemini_client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	defaultGeminiModel = "gemini-3.1-flash-lite"
	geminiStreamURL    = "https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse"
	geminiGenerateURL  = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent"
)

type (
	geminiClient struct {
		apiKey string
		model  string
		http   *http.Client
	}

	Turn struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	StreamChunk struct {
		Text string `json:"text"`
		Err  error  `json:"err"`
	}

	geminiRequest struct {
		SystemInstruction *geminiContent  `json:"systemInstruction,omitempty"`
		Contents          []geminiContent `json:"contents"`
	}

	geminiContent struct {
		Role  string       `json:"role,omitempty"`
		Parts []geminiPart `json:"parts"`
	}

	geminiPart struct {
		Text string `json:"text"`
	}

	geminiResponse struct {
		Candidates []struct {
			Content geminiContent `json:"content"`
		} `json:"candidates"`
		Error *struct {
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"error"`
	}
)

func NewGeminiClient() *geminiClient {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		panic("GEMINI_API_KEY is not set")
	}
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = defaultGeminiModel
	}
	return &geminiClient{
		apiKey: apiKey,
		model:  model,
		http:   &http.Client{},
	}
}

func (c *geminiClient) ChatGenerate(ctx context.Context, prompt string) (string, error) {
	body, err := json.Marshal(geminiRequest{
		Contents: []geminiContent{{
			Role:  "user",
			Parts: []geminiPart{{Text: prompt}},
		}},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf(geminiGenerateURL, c.model), bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	c.setHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	text, err := geminiText(payload)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("gemini status %d: %s", resp.StatusCode, text)
	}
	return text, nil
}

func (c *geminiClient) Stream(ctx context.Context, system string, history []Turn) (<-chan StreamChunk, error) {
	contents := make([]geminiContent, 0, len(history))
	for _, turn := range history {
		role, err := geminiRole(turn.Role)
		if err != nil {
			return nil, err
		}
		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: turn.Content}},
		})
	}

	request := geminiRequest{Contents: contents}
	if strings.TrimSpace(system) != "" {
		request.SystemInstruction = &geminiContent{
			Parts: []geminiPart{{Text: system}},
		}
	}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf(geminiStreamURL, c.model), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	out := make(chan StreamChunk)
	go func() {
		defer close(out)
		defer resp.Body.Close()

		if resp.StatusCode >= 300 {
			payload, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
			sendChunk(ctx, out, StreamChunk{Err: fmt.Errorf("gemini status %d: %s", resp.StatusCode, geminiErrorText(payload))})
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || !strings.HasPrefix(line, "data:") {
				continue
			}
			payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if payload == "[DONE]" {
				return
			}
			text, err := geminiText([]byte(payload))
			if err != nil {
				sendChunk(ctx, out, StreamChunk{Err: err})
				return
			}
			if text == "" {
				continue
			}
			if !sendChunk(ctx, out, StreamChunk{Text: text}) {
				return
			}
		}
		if err := scanner.Err(); err != nil && ctx.Err() == nil {
			sendChunk(ctx, out, StreamChunk{Err: err})
		}
	}()
	return out, nil
}

func (c *geminiClient) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.apiKey)
}

func geminiRole(role string) (string, error) {
	switch role {
	case "user":
		return "user", nil
	case "assistant":
		return "model", nil
	default:
		return "", fmt.Errorf("unknown turn role %q", role)
	}
}

func geminiText(payload []byte) (string, error) {
	if len(bytes.TrimSpace(payload)) == 0 {
		return "", nil
	}
	var parsed geminiResponse
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return "", fmt.Errorf("gemini response: %w", err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return "", fmt.Errorf("gemini %s: %s", parsed.Error.Status, parsed.Error.Message)
	}
	var text strings.Builder
	for _, candidate := range parsed.Candidates {
		for _, part := range candidate.Content.Parts {
			text.WriteString(part.Text)
		}
	}
	return text.String(), nil
}

func geminiErrorText(payload []byte) string {
	text, err := geminiText(payload)
	if err != nil {
		message := err.Error()
		if len(payload) > 0 && !strings.Contains(message, "gemini ") {
			return strings.TrimSpace(string(payload))
		}
		return message
	}
	if text != "" {
		return text
	}
	return strings.TrimSpace(string(payload))
}

func sendChunk(ctx context.Context, out chan<- StreamChunk, chunk StreamChunk) bool {
	select {
	case <-ctx.Done():
		return false
	case out <- chunk:
		return true
	}
}
