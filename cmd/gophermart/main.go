package main

import (
	"context"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/database"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/http"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/logger"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/services"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/utils"
	"log"
)

func main() {
	ctx := context.Background()
	config := NewConfig()

	if err := logger.Initialize(config.logLevel, config.env); err != nil {
		log.Fatalf("Logger wasn't initialized due to %s", err)
	}

	db, err := database.New(ctx, config.dsn)

	if err != nil {
		log.Fatalf("Database wasn't initialized due to %s", err)
	}

	log.Printf("Running server on %s\n", config.endpoint)

	jobQueueService := services.NewJobQueueService(ctx, 100, 2)
	accrualService := services.NewAccrualService(db, jobQueueService, config.accrualEndpoint)

	// todo add init jobs

	utils.HandleTerminationProcess(func() {
		jobQueueService.Shutdown()
	})

	router.New(
		router.Config{Endpoint: config.endpoint},
		services.NewAuthService(db),
		services.NewJWTService(config.authSecretKey),
		services.NewOrderService(db),
		accrualService,
	).Run()
}
