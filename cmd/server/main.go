package main

import (
	"context"
	"log"
	"net/http"
	"vending-qris-service/internal/logger"

	"vending-qris-service/internal/config"
	"vending-qris-service/internal/controller"
	"vending-qris-service/internal/database"
	gwfactory "vending-qris-service/internal/gateway"
	"vending-qris-service/internal/repository"
	"vending-qris-service/internal/usecase"

	"github.com/shopspring/decimal"
)

func main() {
	decimal.MarshalJSONWithoutQuotes = true

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	sync, err := logger.Init()
	if err != nil {
		log.Fatal(err)
	}

	defer sync()

	db, err := database.Open(cfg.Database)
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	if err := database.RunMigrations(db); err != nil {
		logger.Error("db migration failed: %v", err)

		return
	}

	bootstrap := cfg.PaymentGatewayPriority
	if len(bootstrap) == 0 {
		bootstrap = []string{cfg.PaymentGateway}
	}

	if err := database.SeedGatewayPrioritiesIfNotExist(ctx, db, bootstrap); err != nil {
		logger.Error("gateway priority seed: %v", err)

		return
	}

	priorityRepo := repository.NewGatewayPriorityRepository(db)
	resolver := gwfactory.NewResolver(priorityRepo)
	routingUC := usecase.NewGatewayRoutingUsecase(priorityRepo, resolver)

	commRepo := repository.NewCommunicationRepository(db)
	txnRepo := repository.NewTransactionRepository(db)

	transactor := database.NewTransactor(db)
	qrisUC := usecase.NewQRISUsecase(resolver, commRepo, txnRepo, transactor)

	srv := controller.NewHTTPServer(controller.Deps{
		Health:  database.NewHealth(db),
		QRIS:    qrisUC,
		Routing: routingUC,
		AuthKey: cfg.AuthenticationKey,
	})

	addr := ":" + cfg.HTTPPort
	logger.Info("listening on %v", addr)

	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		logger.Error("http error: %v", err)
	}
}
