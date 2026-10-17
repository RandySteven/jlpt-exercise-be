package middlewares

import (
	"log"
	"net/http"
)

func (s *ServerMiddleware) CheckHealthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Check health middleware working fine")
		next.ServeHTTP(w, r)
	})
}
