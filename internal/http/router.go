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
	config         Config
	authService    *services.AuthService
	jwtService     *services.JWTService
	orderService   *services.OrderService
	accrualService *services.AccrualService
}

func New(
	config Config,
	authService *services.AuthService,
	jwtService *services.JWTService,
	orderService *services.OrderService,
	accrualService *services.AccrualService,
) *Router {
	return &Router{config, authService, jwtService, orderService, accrualService}
}

func stub(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Stub"))
}

func (router *Router) get() chi.Router {
	r := chi.NewRouter()

	r.Use(
		middlewares.ServiceInjectorMiddleware(
			router.authService,
			router.jwtService,
			router.orderService,
			router.accrualService,
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

		r.With(middlewares.TextMiddleware).Post("/orders", CreateOrder)
		r.Get("/orders", GetOrders)

		r.Get("/balance", stub)
		r.Post("/balance/withdraw", stub)

		r.Post("/withdrawals", stub)
	})

	return r
}

func (router *Router) Run() {
	log.Fatal(http.ListenAndServe(router.config.Endpoint, router.get()))
}
