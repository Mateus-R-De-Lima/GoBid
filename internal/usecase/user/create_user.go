package user

import (
	"context"

	"github.com/Mateus-R-De-Lima/GoBid/internal/validator"
)

type CreateUserRequest struct {
	UserName string `json:"user_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Bio      string `json:"bio"`
}

func (req CreateUserRequest) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator

	eval.CheckField(validator.NotBlank(req.UserName), "user_name", "user_name cannot be blank")
	eval.CheckField(validator.NotBlank(req.Email), "email", "email cannot be blank")
	eval.CheckField(validator.Matches(req.Email, validator.EmailRX), "email", "email is not valid")
	eval.CheckField(validator.NotBlank(req.Password), "password", "password cannot be blank")

	eval.CheckField(
		validator.MinChars(req.Password, 8) && validator.MaxChars(req.Password, 255),
		"password",
		"password must have a length between 8 and 255",
	)

	eval.CheckField(
		validator.MinChars(req.Bio, 10) && validator.MaxChars(req.Bio, 255),
		"bio",
		"this field must have a length between 10 and 255",
	)
	return eval

}
