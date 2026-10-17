package deepseek

import "net/http"

type (
	deepSeek struct {
		client *http.Client
	}
)
