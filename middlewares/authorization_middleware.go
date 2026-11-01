package middlewares

import (
	"net/http"

	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/enums"
	"github.com/RandySteven/jlpt-exercise-be/utils"
)

func (c *ClientMiddleware) AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//fetch onboarding role, whitelisted
		role := r.Context().Value(enums.RoleID).(string)

		switch role {
		case enums.RegisterAsAdmin.ToString(), enums.RegisterAsUser.ToString(), enums.RegisterAsSensei.ToString():
			if !allowedRoles[role][r.URL.Path] {
				forbiddenErr := apperror.ForbiddenErr()
				utils.ResponseHandler(w, http.StatusForbidden, forbiddenErr.LogMessage, nil, nil, forbiddenErr)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

var allowedRoles = map[string]map[string]bool{
	enums.RegisterAsAdmin.ToString():  map[string]bool{},
	enums.RegisterAsUser.ToString():   map[string]bool{},
	enums.RegisterAsSensei.ToString(): map[string]bool{},
}
