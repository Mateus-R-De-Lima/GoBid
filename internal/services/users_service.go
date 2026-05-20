package services

import (
	"context"
	"errors"

	"github.com/Mateus-R-De-Lima/GoBid/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type UsersService struct {
	pool    *pgxpool.Pool
	queries *pgstore.Queries
}

// Mensagem de erro para indicar que o nome de usuário ou email já existe
var (
	ErrDuplicatedEmailOrUsername = errors.New("username or email already exists")
	ErrInvalidCredentials        = errors.New("invalid credentials")
)

// NewUserService é uma função que cria e retorna uma nova instância de UsersService
func NewUserService(pool *pgxpool.Pool) UsersService {
	return UsersService{
		pool:    pool,
		queries: pgstore.New(pool),
	}
}

func (us *UsersService) CreateUser(ctx context.Context, name, email, password, bio string) (uuid.UUID, error) {
	// Gerando um hash da senha usando bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)

	// Verificando se ocorreu um erro durante a geração do hash
	if err != nil {
		return uuid.UUID{}, err
	}

	// Criando um objeto CreateUserParams com os dados do usuário
	args := pgstore.CreateUserParams{
		UserName:     name,
		Email:        email,
		PasswordHash: hash,
		Bio:          pgtype.Text{String: bio, Valid: true},
	}

	// Chamando o método CreateUser do pgstore para inserir o usuário no banco de dados
	id, err := us.queries.CreateUser(ctx, args)

	// Verificando se ocorreu um erro durante a criação do usuário
	if err != nil {
		// Verificando se o erro é um erro de violação de chave única (código 23505) do PostgreSQL
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return uuid.UUID{}, ErrDuplicatedEmailOrUsername
		}

		return uuid.UUID{}, err
	}
	// Retornando o ID do usuário criado e nil para indicar que não houve erros
	return id, nil
}

func (us *UsersService) AuthenticateUser(ctx context.Context, email, password string) (uuid.UUID, error) {
	user, err := us.queries.GetUserByEmail(ctx, email)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.UUID{}, ErrInvalidCredentials
		}

		return uuid.UUID{}, err
	}

	return user.ID, nil

}
