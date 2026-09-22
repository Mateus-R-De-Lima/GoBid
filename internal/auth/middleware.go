package auth

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				http.Error(w, "Token não informado", http.StatusUnauthorized)
				return
			}
			// Removes os espaços
			parts := strings.Split(authHeader, " ")

			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Token inválido", http.StatusUnauthorized)
				return
			}
			tokenString := parts[1]

			token, err := jwt.Parse(
				tokenString,
				func(t *jwt.Token) (any, error) {
					if t.Method != jwt.SigningMethodHS256 {
						return nil, jwt.ErrSignatureInvalid
					}
					return []byte(secret), nil
				},
			)

			if err != nil || !token.Valid {
				http.Error(w, "Token inválido ou expirado", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
