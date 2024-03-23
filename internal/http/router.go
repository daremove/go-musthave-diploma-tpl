package router

import (
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/logger"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/middlewares"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/models"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/services"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

type Config struct {
	Endpoint string
}

type Router struct {
	config       Config
	authService  services.AuthService
	jwtService   services.JWTService
	orderService services.OrderService
}

func New(
	config Config,
	authService services.AuthService,
	jwtService services.JWTService,
	orderService services.OrderService,
) *Router {
	return &Router{config, authService, jwtService, orderService}
}

func stub(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Stub"))
}

func (router *Router) get() chi.Router {
	r := chi.NewRouter()

	r.Use(
		middlewares.ServiceInjectorMiddleware(
			&router.authService,
			&router.jwtService,
			&router.orderService,
		),
		logger.RequestLogger,
		middlewares.AuthMiddleware().WithExcludedPaths(
			"/api/user/register",
			"/api/user/login",
		).Middleware,
	)
	//r.Use(middleware.NewCompressor(flate.DefaultCompression).Handler)
	//r.Use(dataintergity.NewMiddleware(dataintergity.DataIntegrityMiddlewareConfig{
	//	SigningKey: router.config.SigningKey,
	//}))
	//r.Use(gzipm.GzipMiddleware)

	r.Route("/api/user", func(r chi.Router) {
		r.With(middlewares.JSONMiddleware[models.User]).Post("/register", Register)
		r.With(middlewares.JSONMiddleware[models.User]).Post("/login", Login)

		r.With(middlewares.TextMiddleware).Post("/orders", PostOrders)
		r.Get("/orders", stub)

		r.Get("/balance", stub)
		r.Post("/balance/withdraw", stub)

		r.Post("/withdrawals", stub)
	})

	return r
}

func (router *Router) Run() {
	log.Fatal(http.ListenAndServe(router.config.Endpoint, router.get()))
}
