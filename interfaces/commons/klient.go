package commons

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type (
	Klient interface {
		Call(ctx context.Context, method string, url string, body any, headers map[string]string) (any, error)
	}

	KlientOptions struct {
		BaseURL string
		Headers map[string]string
		Timeout time.Duration
	}

	klient struct {
		baseURL string
		headers map[string]string
		timeout time.Duration
	}
)

func (k *klient) Call(ctx context.Context, method string, url string, body any, headers map[string]string) (any, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response any
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func NewKlient(options KlientOptions) Klient {
	return &klient{
		baseURL: options.BaseURL,
		headers: options.Headers,
		timeout: options.Timeout,
	}
}
