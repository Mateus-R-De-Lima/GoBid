package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	Secret string
}

func NewJwtService(secret string) *JWTService {
	return &JWTService{
		Secret: secret,
	}
}

func (j *JWTService) GenerateToken(userId string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userId,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString([]byte(j.Secret))

}
