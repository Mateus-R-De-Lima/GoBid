package api

import (
	"github.com/Mateus-R-De-Lima/GoBid/internal/services"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

type Api struct {
	Router      *chi.Mux
	UserService services.UsersService
	Sessions    *scs.SessionManager
}
