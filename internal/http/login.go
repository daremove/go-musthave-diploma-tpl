package router

import (
	"errors"
	"fmt"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/middlewares"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/models"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/services"
	"net/http"
)

// todo add invalidation prev token
// todo remove duplication
func Login(w http.ResponseWriter, r *http.Request) {
	data := middlewares.GetParsedJSONData[models.User](w, r)
	authService := middlewares.GetServiceFromContext[services.AuthService](w, r, middlewares.AuthServiceKey)
	jwtService := middlewares.GetServiceFromContext[services.JWTService](w, r, middlewares.JwtServiceKey)

	if data.Login == nil || data.Password == nil {
		http.Error(w, "Request doesn't contain login or password", http.StatusBadRequest)
		return
	}

	if err := authService.Login(r.Context(), data); err != nil {
		if errors.Is(err, services.ErrUserIsNotExist) {
			http.Error(w, fmt.Sprintf("Login %s is not exist", *data.Login), http.StatusUnauthorized)
			return
		}

		if errors.Is(err, services.ErrPasswordIsIncorrect) {
			http.Error(w, "Password is not correct", http.StatusUnauthorized)
			return
		}

		http.Error(w, fmt.Sprintf("Error occurred during login: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	token, err := jwtService.GenerateJWT(*data.Login)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error occurred during generating jwt token: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
}
