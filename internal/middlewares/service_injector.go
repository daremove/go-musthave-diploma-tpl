package middlewares

import (
	"context"
	"fmt"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/services"
	"net/http"
)

type key int

const (
	AuthServiceKey key = iota
	JwtServiceKey
)

func ServiceInjectorMiddleware(authService *services.AuthService, jwtService *services.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), AuthServiceKey, authService)
			ctx = context.WithValue(ctx, JwtServiceKey, jwtService)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetServiceFromContext[Service interface{}](w http.ResponseWriter, r *http.Request, serviceKey key) *Service {
	foundService, ok := r.Context().Value(serviceKey).(*Service)

	if !ok {
		http.Error(w, fmt.Sprintf("Service wasn't found in context by key %v", serviceKey), http.StatusInternalServerError)
		return nil
	}

	return foundService
}
