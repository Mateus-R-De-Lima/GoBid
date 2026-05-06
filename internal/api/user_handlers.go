package api

import (
	"errors"
	"net/http"

	"github.com/Mateus-R-De-Lima/GoBid/internal/jsonutils"
	"github.com/Mateus-R-De-Lima/GoBid/internal/services"
	"github.com/Mateus-R-De-Lima/GoBid/internal/usecase/user"
)

func (a *Api) handleSignupUser(w http.ResponseWriter, r *http.Request) {

	data, problems, err := jsonutils.DecodeJson[user.CreateUserRequest](r)

	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, problems)
		return
	}

	id, err := a.UserService.CreateUser(
		r.Context(),
		data.UserName,
		data.Email,
		data.Password,
		data.Bio)

	if err != nil {
		errors.Is(err, services.ErrDuplicatedEmailOrPassword)
		_ = jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, map[string]any{
			"error": "username or email already exists",
		})
		return
	}

	_ = jsonutils.EncodeJson(w, r, http.StatusCreated, map[string]any{
		"id": id,
	})
}
