package api

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
)

func (a *Api) BindRoutes() {
	a.Router.Use(middleware.RequestID, middleware.Recoverer, middleware.Logger, a.Sessions.LoadAndSave)

	/*csrfMiddleware := csrf.Protect(
		[]byte(os.Getenv("GOBID_CSRF_KEY")),
		csrf.Secure(false), // DEV ONLY
		csrf.Path("/api/v1"),
		csrf.SameSite(csrf.SameSiteLaxMode),
	)

	// In local development over HTTP, gorilla/csrf requires the request to be marked as plaintext.
	// Otherwise it assumes TLS and enforces strict Referer checks.
	a.Router.Use(a.PlaintextHTTP)
	a.Router.Use(csrfMiddleware) */

	a.Router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			//r.Get("/csrftoken", a.HandleGetCSRFToken)
			r.Route("/users", func(r chi.Router) {
				r.Post("/signup", a.handleSignupUser)
				r.Post("/login", a.handleLoginUser)
				r.With(a.AuthMiddleware).Post("/logout", a.handleLogoutUser)
			})
		})
	})

}

func (a *Api) PlaintextHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, csrf.PlaintextHTTPRequest(r))
	})
}
