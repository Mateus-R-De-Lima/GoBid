package user

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testsLoginUserReq = []struct {
	name               string
	req                LoginUserReq
	wantErrors         []string
	wantErrorsMessages []string
}{
	{
		name: "dados validos",
		req: LoginUserReq{
			Email:    "mateus@example.com",
			Password: "senha123",
		},
		wantErrors:         []string{},
		wantErrorsMessages: []string{},
	},
	{
		name: "password vazio",
		req: LoginUserReq{
			Email:    "mateus@example.com",
			Password: "",
		},
		wantErrors:         []string{"password"},
		wantErrorsMessages: []string{"this field cannot be blank"},
	},
	{
		name: "Email  inválido",
		req: LoginUserReq{
			Email:    "email-invalido",
			Password: "senha123",
		},
		wantErrors:         []string{"email"},
		wantErrorsMessages: []string{"must be a valid email"},
	},
	{
		name: "Email  inválido e password vazio",
		req: LoginUserReq{
			Email:    "email-invalido",
			Password: "",
		},
		wantErrors:         []string{"email", "password"},
		wantErrorsMessages: []string{"must be a valid email", "this field cannot be blank"},
	},
}

func TestLoginUserReq_Valid(t *testing.T) {

	for _, tt := range testsLoginUserReq {
		t.Run(tt.name, func(t *testing.T) {
			eval := tt.req.Valid(context.Background())

			if len(eval) != len(tt.wantErrors) {
				t.Fatalf(
					"quantidade de erros = %d; esperava %d. Erros: %v",
					len(eval),
					len(tt.wantErrors),
					eval,
				)
			}

			for _, field := range tt.wantErrors {
				if _, exists := eval[field]; !exists {
					t.Errorf(
						"esperava erro no campo %q, mas ele não foi encontrado. a",
						field,
					)
				}
			}

			for index, message := range tt.wantErrorsMessages {
				msg, _ := eval[tt.wantErrors[index]]
				fmt.Println("Messagem : ", msg)
				assert.Equal(t, message, msg)
			}
		})
	}
}
