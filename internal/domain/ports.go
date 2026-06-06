package domain

import (
	"context"
	"time"
)

// PaymentGatewayResolver picks the live gateway for a new payment (priority + Ping).
type PaymentGatewayResolver interface {
	Resolve(ctx context.Context) (PaymentGateway, error)
}

// CommunicationRepository persists gateway call audit rows.
type CommunicationRepository interface {
	Save(ctx context.Context, comm *PaymentGatewayCommunication) (*PaymentGatewayCommunication, error)
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

// TransactionRepository -  persists all transactions that are requested.
type TransactionRepository interface {
	Save(ctx context.Context, txn *Transaction) (*Transaction, error)
}

// GatewayPriorityRepository manages persisted gateway preference order.
type GatewayPriorityRepository interface {
	ListOrdered(ctx context.Context) ([]string, error)
	ReplaceAll(ctx context.Context, gateways []string) error
	RotateOnce(ctx context.Context) error
}

// GatewayFactory constructs a payment gateway by provider id.
type GatewayFactory func(name string) (PaymentGateway, error)
