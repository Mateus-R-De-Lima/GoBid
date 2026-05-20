package jsonutils

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Mateus-R-De-Lima/GoBid/internal/validator"
)

// Função que recebe um http.ResponseWriter, um http.Request, um status code e um dado genérico, e escreve o dado como json no response. Se ocorrer um erro ao codificar o json, a função retorna um erro.
func EncodeJson[T any](w http.ResponseWriter, r *http.Request, statusCode int, data T) error {
	// Define o header Content-Type como application/json, escreve o status code no response e codifica o dado como json. Se ocorrer um erro ao codificar o json, a função retorna um erro.
	w.Header().Set("Content-Type", "application/json")
	// Escreve o status code no response e codifica o dado como json. Se ocorrer um erro ao codificar o json, a função retorna um erro.
	w.WriteHeader(statusCode)
	//Verifica se ocorreu um erro ao codificar o json, e caso tenha ocorrido, retorna um erro com a mensagem "failed to encode json" e o erro original.
	if err := json.NewEncoder(w).Encode(data); err != nil {
		return fmt.Errorf("failed to encode json %w", err)
	}
	// Se o json foi codificado com sucesso, a função retorna nil.
	return nil
}

// Função que recebe um http.Request e um tipo genérico que implementa a interface Validator, e retorna o dado decodificado, um mapa de erros de validação e um erro. A função decodifica o json do request para o tipo genérico, e se ocorrer um erro ao decodificar, retorna um erro. Em seguida, a função chama o método Valid do tipo genérico para validar os dados, e se houver erros de validação, retorna o mapa de erros. Se os dados forem válidos, a função retorna o dado decodificado e nil para os erros de validação e o erro.
func DecodeJson[T validator.Validator](r *http.Request) (T, map[string]string, error) {
	// Declara uma variável do tipo genérico T, que será usada para armazenar o dado decodificado do json.
	var data T
	//Verifica se ocorreu um erro ao decodificar o json do request para a variável data, e caso tenha ocorrido, retorna um erro com a mensagem "failed to decode json" e o erro original.
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return data, nil, fmt.Errorf("failed to decode json %w", err)
	}
	// Chama o método Valid do tipo genérico para validar os dados, e se houver erros de validação, retorna o mapa de erros. Se os dados forem válidos, a função retorna o dado decodificado e nil para os erros de validação e o erro.
	if problems := data.Valid(r.Context()); len(problems) > 0 {
		return data, problems, fmt.Errorf("invalid %T: %d problems", data, len(problems))
	}

	// Se os dados forem válidos, a função retorna o dado decodificado e nil para os erros de validação e o erro.
	return data, nil, nil
}
