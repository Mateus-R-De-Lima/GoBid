package api

import (
	"github.com/Mateus-R-De-Lima/GoBid/internal/services"
	"github.com/go-chi/chi/v5"
)

type Api struct {
	Router      *chi.Mux
	UserService services.UsersService //
}
