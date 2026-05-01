package validator

import (
	"context"
	"strings"
)

// Interface que vai conter o método Valid, que retorna um Evaluator, que é um mapa de string para string, onde a chave é o nome do campo e o valor é a mensagem de erro.
type Validator interface {
	Valid(context.Context) Evaluator
}

type Evaluator map[string]string

// Função que adiciona um erro ao Evaluator, onde a chave é o nome do campo e o valor é a mensagem de erro. Se o Evaluator for nil, ele é inicializado como um mapa vazio.
func (e *Evaluator) AddError(field, message string) {
	if *e == nil {
		*e = make(map[string]string)
	}

	if _, exists := (*e)[field]; !exists {
		(*e)[field] = message
	}
}

// Função que verifica se um campo é válido, onde ok é um booleano que indica se o campo é válido ou não, key é o nome do campo e message é a mensagem de erro. Se ok for false, a mensagem de erro é adicionada ao Evaluator.
func (e *Evaluator) CheckField(ok bool, key, message string) {
	if !ok {
		e.AddError(key, message)
	}
}

// Função que verifica se o valor de uma string é diferente de vazio, removendo os espaços em branco. Retorna true se a string não for vazia, e false caso contrário.
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}
