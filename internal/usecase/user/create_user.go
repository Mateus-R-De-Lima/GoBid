package user

import (
	"context"

	"github.com/Mateus-R-De-Lima/GoBid/internal/validator"
)

type CreateUserRequest struct {
	UserName     string `json:"user_name"`
	Email        string `json:"email"`
	PasswordHash []byte `json:"password_hash"`
	Bio          string `json:"bio"`
}

func (req CreateUserRequest) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator

	eval.CheckField(validator.NotBlank(req.UserName), "user_name", "user_name cannot be blank")
	eval.CheckField(validator.NotBlank(req.Email), "email", "email cannot be blank")
	eval.CheckField(len(req.PasswordHash) > 0, "password_hash", "password_hash cannot be blank")

	return eval

}
