package router

import (
	"errors"
	"fmt"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/middlewares"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/models"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/services"
	"net/http"
)

func Register(w http.ResponseWriter, r *http.Request) {
	data := middlewares.GetParsedJSONData[models.User](w, r)
	authService := middlewares.GetServiceFromContext[services.AuthService](w, r, middlewares.AuthServiceKey)
	jwtService := middlewares.GetServiceFromContext[services.JWTService](w, r, middlewares.JwtServiceKey)

	if data.Login == nil || data.Password == nil {
		http.Error(w, "Request doesn't contain login or password", http.StatusBadRequest)
		return
	}

	if err := authService.Register(r.Context(), data); err != nil {
		if errors.Is(err, services.ErrUserIsAlreadyRegistered) {
			http.Error(w, "User is already registered", http.StatusConflict)
			return
		}

		http.Error(w, fmt.Sprintf("Error occurred during registration: %s", err.Error()), http.StatusBadRequest)
		return
	}

	// todo think about atomicity
	token, err := jwtService.GenerateJWT(*data.Login)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error occurred during generating jwt token: %s", err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
}
