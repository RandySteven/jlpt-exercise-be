package middlewares

import (
	"fmt"
	"net/http"

	"github.com/RandySteven/go-kopi/enums"
)

type (
	Middlewares struct {
		middlewares map[enums.Middleware]map[string]bool
	}

	ClientMiddlewares interface {
		AuthenticationMiddleware(next http.Handler) http.Handler
		RateLimiterMiddleware(next http.Handler) http.Handler
	}

	ServerMiddlewares interface {
		TimeoutMiddleware(next http.Handler) http.Handler
		LoggingMiddleware(next http.Handler) http.Handler
		CorsMiddleware(next http.Handler) http.Handler
		CheckHealthMiddleware(next http.Handler) http.Handler
	}
)

func NewMiddlewares() *Middlewares {
	return &Middlewares{
		middlewares: make(map[enums.Middleware]map[string]bool),
	}
}

func (m *Middlewares) RegisterMiddleware(prefix enums.RouterPrefix, method string, path string, middlewares []enums.Middleware) {
	if m == nil {
		_ = NewMiddlewares()
	}
	if middlewares == nil {
		return
	}
	whitelist := fmt.Sprintf("%s|%s%s", method, prefix.ToString(), path)
	for _, middleware := range middlewares {
		if m.middlewares[middleware] == nil {
			m.middlewares[middleware] = make(map[string]bool)
		}
		m.middlewares[middleware][whitelist] = true
	}
}

func (m *Middlewares) WhiteListed(method string, uri string, middleware enums.Middleware) bool {
	whiteList := fmt.Sprintf("%s|%s", method, uri)
	return m.middlewares[middleware][whiteList]
}
