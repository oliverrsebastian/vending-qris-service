package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"vending-qris-service/internal/config"
	"vending-qris-service/internal/controller"
	"vending-qris-service/internal/database"
	gwfactory "vending-qris-service/internal/gateway"
	"vending-qris-service/internal/repository"
	"vending-qris-service/internal/usecase"
	"vending-qris-service/internal/worker"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Open(cfg.Database)
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	bootstrap := cfg.PaymentGatewayPriority
	if len(bootstrap) == 0 {
		bootstrap = []string{cfg.PaymentGateway}
	}

	if err := database.SeedGatewayPriorities(ctx, db, bootstrap); err != nil {
		log.Fatalf("gateway priority seed: %v", err)
	}

	priorityRepo := repository.NewGatewayPriorityRepository(db)
	resolver := gwfactory.NewResolver(priorityRepo)
	routingUC := usecase.NewGatewayRoutingUsecase(priorityRepo, resolver)

	commRepo := repository.NewCommunicationRepository(db)
	qrisUC := usecase.NewQRISUsecase(resolver, commRepo)

	retryPolicy := usecase.RetryPolicy{
		Enabled:                   cfg.PaymentCommunicationRetry.Enabled,
		IntervalSeconds:           cfg.PaymentCommunicationRetry.IntervalSeconds,
		RetryableResponseStatuses: cfg.PaymentCommunicationRetry.RetryableResponseStatuses,
		MaxPollAttempts:           cfg.PaymentCommunicationRetry.MaxPollAttempts,
		BatchLimit:                cfg.PaymentCommunicationRetry.BatchLimit,
	}
	retryUC := usecase.NewCommunicationRetryUsecase(gwfactory.New, commRepo, retryPolicy)

	srv := controller.NewHTTPServer(controller.Deps{
		Health:  database.NewHealth(db),
		QRIS:    qrisUC,
		Retry:   retryUC,
		Routing: routingUC,
	})

	if cfg.PaymentCommunicationRetry.Enabled {
		interval := cfg.PaymentCommunicationRetry.IntervalSeconds
		log.Printf("payment communication retry poller enabled (every %ds)", interval)
		go worker.RunPaymentCommunicationPoll(ctx, retryPolicy, retryUC)
	}

	addr := ":" + cfg.HTTPPort
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		log.Fatal(fmt.Errorf("http: %w", err))
	}
}
