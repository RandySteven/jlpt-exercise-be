package middlewares

import (
	"context"
	"net/http"

	"github.com/RandySteven/go-kopi/enums"
	redis_client "github.com/RandySteven/go-cook/redis"
	"github.com/RandySteven/go-kopi/utils"
	ip "github.com/vikram1565/request-ip"
)

func (c *ClientMiddleware) RateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !c.middlewares.WhiteListed(r.Method, utils.ReplaceLastURLID(r.RequestURI), enums.RateLimiterMiddleware) {
			next.ServeHTTP(w, r)
			return
		}
		clientIp := ip.GetClientIP(r)
		ctx := context.WithValue(r.Context(), enums.ClientIP, clientIp)
		if err := redis_client.RateLimiter(ctx); err != nil {
			w.WriteHeader(http.StatusTooManyRequests)
			utils.ResponseHandler(w, http.StatusTooManyRequests, `too many request`, nil, nil, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}
