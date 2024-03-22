package middlewares

import (
	"errors"
	"fmt"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/services"
	"net/http"
	"strings"
)

func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authService := GetServiceFromContext[services.AuthService](w, r, AuthServiceKey)
		jwtService := GetServiceFromContext[services.JWTService](w, r, JwtServiceKey)

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		if tokenString == "" {
			http.Error(w, "Bearer token is empty", http.StatusUnauthorized)
			return
		}

		token, err := jwtService.ValidateToken(tokenString)

		if err != nil {
			if errors.Is(err, services.ErrTokenIsInvalid) {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			if errors.Is(err, services.ErrTokenIsExpired) {
				http.Error(w, "Token is expired", http.StatusUnauthorized)
				return
			}

			http.Error(w, fmt.Sprintf("Error occurred during validating token: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		login, err := token.Claims.GetSubject()

		if err != nil {
			http.Error(w, fmt.Sprintf("Error occurred during reading sub field: %s", err.Error()), http.StatusUnauthorized)
			return
		}

		isValid, err := authService.IsLoginValid(r.Context(), login)

		if err != nil {
			http.Error(w, fmt.Sprintf("Error occurred during validation user login: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		if !isValid {
			http.Error(w, fmt.Sprintf("User login %s doesn't exist", login), http.StatusConflict)
			return
		}

		next.ServeHTTP(w, r)
	}
}
