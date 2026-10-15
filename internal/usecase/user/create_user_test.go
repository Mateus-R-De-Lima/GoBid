package user

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testsCreateUserRequest = []struct {
	name               string
	req                CreateUserRequest
	wantErrors         []string
	wantErrorsMessages []string
}{
	{
		name: "dados válidos",
		req: CreateUserRequest{
			UserName: "mateus",
			Email:    "mateus@example.com",
			Password: "senha123",
			Bio:      "Desenvolvedor backend Go",
		},
		wantErrors:         []string{},
		wantErrorsMessages: []string{},
	},
	{
		name: "nome vazio",
		req: CreateUserRequest{
			UserName: "",
			Email:    "mateus@example.com",
			Password: "senha123",
			Bio:      "Desenvolvedor backend Go",
		},
		wantErrors:         []string{"user_name"},
		wantErrorsMessages: []string{"user_name cannot be blank"},
	},
	{
		name: "e-mail inválido",
		req: CreateUserRequest{
			UserName: "mateus",
			Email:    "email-invalido",
			Password: "senha123",
			Bio:      "Desenvolvedor backend Go",
		},
		wantErrors:         []string{"email"},
		wantErrorsMessages: []string{"email is not valid"},
	},
	{
		name: "e-mail vazio",
		req: CreateUserRequest{
			UserName: "mateus",
			Email:    "",
			Password: "senha123",
			Bio:      "Desenvolvedor backend Go",
		},
		wantErrors:         []string{"email"},
		wantErrorsMessages: []string{"email cannot be blank"},
	},
}

func TestCreateUserRequest_Valid(t *testing.T) {

	for _, tt := range testsCreateUserRequest {
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
						"esperava erro no campo %q, mas ele não foi encontrado",
						field,
					)
				}
			}

			for index, message := range tt.wantErrorsMessages {
				msg, _ := eval[tt.wantErrors[index]]

				assert.Equal(t, message, msg)
			}
		})
	}
}
