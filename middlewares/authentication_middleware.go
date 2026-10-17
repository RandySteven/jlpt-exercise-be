package middlewares

import (
	"context"
	"net/http"

	"github.com/RandySteven/go-kopi/enums"
	jwt_client "github.com/RandySteven/go-cook/security"
	"github.com/RandySteven/go-kopi/utils"
	"github.com/golang-jwt/jwt/v5"
)

func (c *ClientMiddleware) AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !c.middlewares.WhiteListed(r.Method, utils.ReplaceLastURLID(r.RequestURI), enums.AuthenticationMiddleware) {
			next.ServeHTTP(w, r)
			return
		}
		auth := r.Header.Get("Authorization")
		if len(auth) == 0 || auth == "" {
			utils.ResponseHandler(w, http.StatusUnauthorized, `Invalid get token from auth`, nil, nil, nil)
			return
		}
		tokenStr := auth[len("Bearer "):]
		if tokenStr == "" {
			utils.ResponseHandler(w, http.StatusUnauthorized, `Invalid token failed to split from bearer`, nil, nil, nil)
			return
		}
		claims := &jwt_client.JWTAccessClaim{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(j *jwt.Token) (interface{}, error) {
			return jwt_client.JwtKey, nil
		})
		if err != nil || !token.Valid {
			utils.ResponseHandler(w, http.StatusUnauthorized, `Invalid token`, nil, nil, err)
			return
		}
		ctx := context.WithValue(r.Context(), enums.UserID, claims.UserID)
		//ctx2 := context.WithValue(ctx, enums.RoleID, claims.RoleID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
