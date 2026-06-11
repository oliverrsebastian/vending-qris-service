package repository

import (
	"context"
	"errors"
	"time"

	"vending-qris-service/internal/domain"

	"gorm.io/gorm"
)

type CommunicationRepository struct {
	db *gorm.DB
}

func NewCommunicationRepository(db *gorm.DB) *CommunicationRepository {
	return &CommunicationRepository{db: db}
}

func (r *CommunicationRepository) Save(ctx context.Context, paymentGatewayCommunication *domain.PaymentGatewayCommunication) (*domain.PaymentGatewayCommunication, error) {
	if err := DBFromContext(ctx, r.db).Save(paymentGatewayCommunication).Error; err != nil {
		return nil, err
	}

	return paymentGatewayCommunication, nil
}

func (r *CommunicationRepository) FindLatestByTransactionAndOperation(
	ctx context.Context,
	transactionID int64,
	operation string,
) (*domain.PaymentGatewayCommunication, error) {
	var comm domain.PaymentGatewayCommunication
	err := DBFromContext(ctx, r.db).
		Where("transaction_id = ? AND operation = ?", transactionID, operation).
		Order("id DESC").
		First(&comm).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &comm, nil
}

func (r *CommunicationRepository) UpdateGatewayResponse(
	ctx context.Context,
	id int64,
	gatewayName string,
	responseJSON string,
	responseStatus int,
	responseTimestamp time.Time,
) error {
	updates := map[string]any{
		"response_json":      responseJSON,
		"response_status":    responseStatus,
		"response_timestamp": responseTimestamp,
	}
	if gatewayName != "" {
		updates["gateway_name"] = gatewayName
	}
	return DBFromContext(ctx, r.db).Model(&domain.PaymentGatewayCommunication{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *CommunicationRepository) ListRetryableByResponseStatus(
	ctx context.Context,
	responseStatuses []string,
	maxPollAttempts int,
	limit int,
) ([]domain.PaymentGatewayCommunication, error) {
	var rows []domain.PaymentGatewayCommunication
	q := r.db.WithContext(ctx).Model(&domain.PaymentGatewayCommunication{}).
		Where("response_status IN ?", responseStatuses).
		Where("poll_attempts < ?", maxPollAttempts).
		Order("id ASC").
		Limit(limit)
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *CommunicationRepository) UpdateAfterStatusPoll(
	ctx context.Context,
	id int64,
	responseJSON []byte,
	responseStatus string,
	responseTimestamp time.Time,
	pollAttempts int,
) error {
	return DBFromContext(ctx, r.db).Model(&domain.PaymentGatewayCommunication{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"response_json":      responseJSON,
			"response_status":    responseStatus,
			"response_timestamp": responseTimestamp,
			"poll_attempts":      pollAttempts,
		}).Error
}
