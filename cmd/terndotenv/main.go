package main

import (
	"fmt"
	"os/exec"

	"github.com/joho/godotenv"
)

func main() {

	// 1 - Carregar as variáveis de ambiente do arquivo .env
	if err := godotenv.Load(); err != nil {
		panic("Erro ao carregar o arquivo .env")
	}

	// 2 - Criar o comando para executar a migração usando o tern
	comando := exec.Command(
		"tern",
		"migrate",
		"--migrations",
		"./internal/store/pgstore/migrations",
		"--config",
		"./internal/store/pgstore/migrations/tern.conf",
	)
	// 3 - Executar o comando e capturar a saída
	output, err := comando.CombinedOutput()
	if err != nil {
		fmt.Println("Falha ao executar o comando: ", err)
		fmt.Println("Saída do comando: ", string(output))
		panic("Erro ao executar a migração: " + err.Error())
	}

	// 4 - Imprimir a saída do comando
	fmt.Println("Comando executado com sucesso: ", string(output))

}
