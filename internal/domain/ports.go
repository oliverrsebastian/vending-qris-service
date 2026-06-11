package domain

import (
	"context"
	"time"
)

// PaymentGatewayResolver picks the live gateway for a new payment (priority + Ping).
type PaymentGatewayResolver interface {
	Resolve(ctx context.Context) (PaymentGateway, error)
	ResolveByName(ctx context.Context, name string) (PaymentGateway, error)
}

// CommunicationRepository persists gateway call audit rows.
type CommunicationRepository interface {
	Save(ctx context.Context, comm *PaymentGatewayCommunication) (*PaymentGatewayCommunication, error)
	FindLatestByTransactionAndOperation(ctx context.Context, transactionID int64, operation string) (*PaymentGatewayCommunication, error)
	UpdateGatewayResponse(
		ctx context.Context,
		id int64,
		gatewayName string,
		responseJSON string,
		responseStatus int,
		responseTimestamp time.Time,
	) error
	ListRetryableByResponseStatus(
		ctx context.Context,
		responseStatuses []string,
		maxPollAttempts int,
		limit int,
	) ([]PaymentGatewayCommunication, error)
	UpdateAfterStatusPoll(
		ctx context.Context,
		id int64,
		responseJSON []byte,
		responseStatus string,
		responseTimestamp time.Time,
		pollAttempts int,
	) error
}

// TransactionRepository persists all transactions that are requested.
type TransactionRepository interface {
	Save(ctx context.Context, txn *Transaction) (*Transaction, error)
	FindByID(ctx context.Context, id int64) (*Transaction, error)
	FindByInvoiceNumber(ctx context.Context, invoiceNumber string) (*Transaction, error)
}

// Transactor runs a callback inside a single database transaction.
// Orchestration lives in use cases; repositories participate via context-bound connections.
type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// GatewayPriorityRepository manages persisted gateway preference order.
type GatewayPriorityRepository interface {
	ListOrdered(ctx context.Context) ([]string, error)
	ReplaceAll(ctx context.Context, gateways []string) error
	RotateOnce(ctx context.Context) error
}

// GatewayFactory constructs a payment gateway by provider id.
type GatewayFactory func(name string) (PaymentGateway, error)
