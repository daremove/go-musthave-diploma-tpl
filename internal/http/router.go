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
	balanceService *services.BalanceService
}

func New(
	config Config,
	authService *services.AuthService,
	jwtService *services.JWTService,
	orderService *services.OrderService,
	accrualService *services.AccrualService,
	balanceService *services.BalanceService,
) *Router {
	return &Router{
		config,
		authService,
		jwtService,
		orderService,
		accrualService,
		balanceService,
	}
}

func (router *Router) get() chi.Router {
	r := chi.NewRouter()

	r.Use(
		middlewares.ServiceInjectorMiddleware(
			router.authService,
			router.jwtService,
			router.orderService,
			router.accrualService,
			router.balanceService,
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
		r.With(middlewares.JSONMiddleware[models.UnknownUser]).Post("/register", Register)
		r.With(middlewares.JSONMiddleware[models.UnknownUser]).Post("/login", Login)

		r.With(middlewares.TextMiddleware).Post("/orders", CreateOrder)
		r.Get("/orders", GetOrders)

		r.Get("/balance", GetBalance)
		r.With(middlewares.JSONMiddleware[models.Withdrawal]).Post("/balance/withdraw", CreateWithdrawal)

		r.Get("/withdrawals", GetWithdrawals)
	})

	return r
}

func (router *Router) Run() {
	log.Fatal(http.ListenAndServe(router.config.Endpoint, router.get()))
}
