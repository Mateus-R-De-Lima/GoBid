package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/Mateus-R-De-Lima/GoBid/internal/api"
	"github.com/Mateus-R-De-Lima/GoBid/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Carregando as variáveis de ambiente do arquivo .env e tratando erros caso o arquivo não seja encontrado ou haja problemas na leitura
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	// Criando um contexto para a aplicação, que pode ser usado para controlar o tempo de vida de operações assíncronas e cancelamento
	ctx := context.Background()

	// Criando uma conexão com o banco de dados PostgreSQL usando pgxpool, utilizando as variáveis de ambiente para configurar a conexão
	pool, err := pgxpool.New(ctx, fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		os.Getenv("GOBID_DATABASE_USER"),
		os.Getenv("GOBID_DATABASE_PASSWORD"),
		os.Getenv("GOBID_DATABASE_HOST"),
		os.Getenv("GOBID_DATABASE_PORT"),
		os.Getenv("GOBID_DATABASE_NAME")))

	// Verificando se ocorreu um erro durante a criação da conexão com o banco de dados e tratando o erro caso necessário
	if err != nil {
		panic(err)
	}

	// Garantindo que a conexão com o banco de dados seja fechada quando a função main terminar sua execução
	defer pool.Close()

	// Verificando se a conexão com o banco de dados está ativa, enviando um ping para o banco e tratando o erro caso a conexão não esteja ativa
	if err := pool.Ping(ctx); err != nil {
		panic(err)
	}
	// Criando uma instância da API, configurando o roteador e os serviços necessários para a aplicação
	api := api.Api{
		Router:      chi.NewMux(),
		UserService: services.NewUserService(pool),
	}
	// Vinculando as rotas da API aos manipuladores de requisições correspondentes
	api.BindRoutes()

	fmt.Println("Server is running on port 8080")

	if err := http.ListenAndServe(":8080", api.Router); err != nil {
		panic(err)
	}
}
