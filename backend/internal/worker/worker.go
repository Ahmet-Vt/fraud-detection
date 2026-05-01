package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"fraud-detection/internal/domain"
	"fraud-detection/internal/fraud"
	"fraud-detection/internal/messaging/rabbitmq"

	"github.com/google/uuid"
)

type fraudService interface {
	EvaluateTransaction(ctx context.Context, transactionID uuid.UUID, userID uuid.UUID) ([]domain.TransactionFraudRule, error)
}

type transactionStatusUpdater interface {
	UpdateTransactionStatus(transactionID uuid.UUID, status domain.Status) error
}

type FraudWorker struct {
	fraudService       fraudService
	transactionService transactionStatusUpdater
}

func NewFraudWorker(fraudService fraudService, transactionService transactionStatusUpdater) *FraudWorker {
	return &FraudWorker{
		fraudService:       fraudService,
		transactionService: transactionService,
	}
}

func (p *FraudWorker) HandleTransactionCreated(ctx context.Context, evt rabbitmq.TransactionCreatedMessage) error {
	if p == nil || p.fraudService == nil {
		return fmt.Errorf("worker FraudWorker: nil transaction service")
	}

	transactionID, err := uuid.Parse(evt.TransactionID)
	if err != nil {
		return fmt.Errorf("worker FraudWorker: %w", errors.Join(fraud.ErrInvalidTransactionID, err))
	}

	userID, err := uuid.Parse(evt.UserID)
	if err != nil {
		p.markTransactionFailed(transactionID)
		return fmt.Errorf("worker FraudWorker: %w", errors.Join(fraud.ErrInvalidUserID, err))
	}

	evaluationStartedAt := time.Now()
	rules, err := p.fraudService.EvaluateTransaction(ctx, transactionID, userID)
	evaluationDurationMs := time.Since(evaluationStartedAt).Milliseconds()
	if err != nil {
		p.markTransactionFailed(transactionID)
		log.Printf(
			"worker: fraud evaluation failed transaction_id=%s user_id=%s evaluate_transaction_duration_ms=%d error=%v",
			transactionID,
			userID,
			evaluationDurationMs,
			err,
		)
		return fmt.Errorf("worker FraudWorker: evaluate transaction: %w", err)
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
		"worker: fraud evaluation transaction_id=%s user_id=%s matched_rules=%d fraud_status=%t evaluate_transaction_duration_ms=%d",
		transactionID,
		userID,
		matchedRules,
		fraudStatus,
		evaluationDurationMs,
	)

	return nil
}

func (p *FraudWorker) markTransactionFailed(transactionID uuid.UUID) {
	if p == nil || p.transactionService == nil {
		return
	}
	if err := p.transactionService.UpdateTransactionStatus(transactionID, domain.TransactionStatusFailed); err != nil {
		log.Printf("worker: mark transaction failed transaction_id=%s: %v", transactionID, err)
	}
}
