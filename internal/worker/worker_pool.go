package worker

import (
	"context"
	"errors"
	"fmt"
	"log"

	"fraud-detection/internal/domain"
	"fraud-detection/internal/fraud"
	"fraud-detection/internal/messaging/rabbitmq"

	"github.com/google/uuid"
)

type transactionService interface {
	EvaluateTransaction(ctx context.Context, transactionID uuid.UUID, userID uuid.UUID) ([]domain.TransactionFraudRule, error)
}

type Pool struct {
	transactionService transactionService
}

func NewPool(transactionService transactionService) *Pool {
	return &Pool{
		transactionService: transactionService,
	}
}

func (p *Pool) HandleTransactionCreated(ctx context.Context, evt rabbitmq.TransactionCreatedMessage) error {
	if p == nil || p.transactionService == nil {
		return fmt.Errorf("worker pool: nil transaction service")
	}

	transactionID, err := uuid.Parse(evt.TransactionID)
	if err != nil {
		return fmt.Errorf("worker pool: %w", errors.Join(fraud.ErrInvalidTransactionID, err))
	}

	userID, err := uuid.Parse(evt.UserID)
	if err != nil {
		return fmt.Errorf("worker pool: %w", errors.Join(fraud.ErrInvalidUserID, err))
	}

	rules, err := p.transactionService.EvaluateTransaction(ctx, transactionID, userID)
	if err != nil {
		return fmt.Errorf("worker pool: evaluate transaction: %w", err)
	}

	matchedRules := 0
	for _, rule := range rules {
		if rule.Matched {
			matchedRules++
		}
		log.Printf(
			"worker: rule result transaction_id=%s user_id=%s rule=%s matched=%t",
			transactionID,
			userID,
			rule.RuleName,
			rule.Matched,
		)
	}
	fraudStatus := matchedRules >= 2
	log.Printf(
		"worker: fraud evaluation transaction_id=%s user_id=%s matched_rules=%d fraud_status=%t",
		transactionID,
		userID,
		matchedRules,
		fraudStatus,
	)

	return nil
}
