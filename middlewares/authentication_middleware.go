package middlewares

import (
	"context"
	"net/http"

	jwt_client "github.com/RandySteven/go-cook/security/v2"
	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/enums"
	"github.com/RandySteven/jlpt-exercise-be/utils"
	"github.com/golang-jwt/jwt/v5"
)

func (c *ClientMiddleware) AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !c.middlewares.WhiteListed(r.Method, utils.ReplaceLastURLID(r.RequestURI), enums.AuthenticationMiddleware) {
			next.ServeHTTP(w, r)
			return
		}
		tokenStr, err := utils.GetAuthTokenFromHeader(r)
		if err != nil {
			utils.ResponseHandler(w, http.StatusUnauthorized, `Invalid token`, nil, nil, err)
			return
		}
		claims := &jwt_client.JWTAccessClaim{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(j *jwt.Token) (interface{}, error) {
			return jwt_client.JwtKey, nil
		})
		if err != nil || !token.Valid {
			invalidTokenErr := apperror.InvalidTokenErr()
			utils.ResponseHandler(w, http.StatusUnauthorized, invalidTokenErr.LogMessage, nil, nil, invalidTokenErr)
			return
		}
		ctx := context.WithValue(r.Context(), enums.UserID, claims.ID)
		ctx2 := context.WithValue(ctx, enums.RoleID, claims.Role)
		r = r.WithContext(ctx2)
		next.ServeHTTP(w, r)
	})
}
