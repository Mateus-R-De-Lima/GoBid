package api

import (
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func (a *Api) BindRoutes() {
	a.Router.Use(middleware.RequestID, middleware.Recoverer, middleware.Logger, a.Sessions.LoadAndSave)
	a.Router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Route("/users", func(r chi.Router) {
				r.Post("/signup", a.handleSignupUser)
				r.Post("/login", a.handleLoginUser)
				r.With(a.AuthMiddleware).Post("/logout", a.handleLogoutUser)
			})
		})
	})

}
