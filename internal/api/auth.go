package api

import (
	"net/http"

	"github.com/Mateus-R-De-Lima/GoBid/internal/jsonutils"
	"github.com/gorilla/csrf"
)

func (api *Api) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !api.Sessions.Exists(r.Context(), "AuthenticatedUserId") {
			jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]any{
				"message": "must be logged in",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (api *Api) HandleGetCSRFToken(w http.ResponseWriter, r *http.Request) {
	token := csrf.Token(r)
	w.Header().Set("X-CSRF-Token", token)
	jsonutils.EncodeJson(w, r, http.StatusOK, map[string]any{
		"csrf_token":  token,
		"csrf_header": "X-CSRF-Token",
	})
}
