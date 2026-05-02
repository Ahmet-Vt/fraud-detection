package transaction

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"fraud-detection/internal/domain"

	"github.com/google/uuid"
)

type transactionCreatedPublisher interface {
	PublishTransactionCreated(ctx context.Context, evt TransactionCreatedEvent) error
}

type transactionEvaluatedPublisher interface {
	PublishTransactionEvaluated(ctx context.Context, evt TransactionEvaluatedEvent) error
}

type transactionStateCache interface {
	GetLastLocationByUserID(ctx context.Context, userID uuid.UUID) (string, bool, error)
	SetLastLocationByUserID(ctx context.Context, userID uuid.UUID, location string) error
	SetLastTransactionCreatedAtByUserID(ctx context.Context, userID uuid.UUID, createdAt time.Time) error
	IncrementAndGetRequestCountByUserID(ctx context.Context, userID uuid.UUID, requestTime time.Time) (int, error)
}

type Service struct {
	repository         RepositoryInterface
	createdPublisher   transactionCreatedPublisher
	evaluatedPublisher transactionEvaluatedPublisher
	cache              transactionStateCache
}

const requestCountFraudThreshold = 5

func NewService(
	repository RepositoryInterface,
	createdPublisher transactionCreatedPublisher,
	evaluatedPublisher transactionEvaluatedPublisher,
	cache transactionStateCache,
) *Service {
	return &Service{
		repository:         repository,
		createdPublisher:   createdPublisher,
		evaluatedPublisher: evaluatedPublisher,
		cache:              cache,
	}
}

func (s *Service) CreateTransaction(ctx context.Context, transaction *domain.Transaction) error {
	if transaction == nil {
		return fmt.Errorf("create transaction: %w", errors.New("nil transaction"))
	}
	if transaction.TransactionStatus == "" {
		transaction.TransactionStatus = domain.TransactionStatusPending
	}

	if err := s.repository.Create(transaction); err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}

	if err := s.SetLastLocationByUserID(ctx, transaction.UserID, transaction.Location); err != nil {
		log.Printf("set last location cache after create transaction: %v", err)
	}
	if s.cache != nil {
		if err := s.cache.SetLastTransactionCreatedAtByUserID(ctx, transaction.UserID, transaction.CreatedAt); err != nil {
			log.Printf("set last transaction created at cache after create transaction: %v", err)
		}
	}
	if s.cache != nil {
		if requestCount, err := s.cache.IncrementAndGetRequestCountByUserID(ctx, transaction.UserID, transaction.CreatedAt); err != nil {
			log.Printf("increment request count cache after create transaction: %v", err)
		} else if requestCount > requestCountFraudThreshold {
			log.Printf("user %s reached %d requests in last 1 minute", transaction.UserID, requestCount)
		}
	}

	if s.createdPublisher != nil {
		evt := TransactionCreatedEvent{
			TransactionID: transaction.ID.String(),
			UserID:        transaction.UserID.String(),
			Amount:        transaction.Amount,
			Location:      transaction.Location,
			CreatedAt:     transaction.CreatedAt,
		}
		if err := s.createdPublisher.PublishTransactionCreated(ctx, evt); err != nil {
			log.Printf("publish transaction created event: %v", err)
		}
	}

	return nil
}

func (s *Service) UpdateFraudRules(transactionID uuid.UUID, rules []domain.TransactionFraudRule) error {
	if err := s.repository.UpdateFraudRules(transactionID, rules); err != nil {
		return fmt.Errorf("replace fraud rules: %w", err)
	}

	return nil
}

func (s *Service) UpdateTransactionStatus(transactionID uuid.UUID, status domain.Status) error {
	if err := s.repository.UpdateStatus(transactionID, status); err != nil {
		return fmt.Errorf("update transaction status: %w", err)
	}

	return nil
}

func (s *Service) GetTransactionByID(id uuid.UUID) (*domain.Transaction, error) {
	transaction, err := s.repository.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("get transaction by id: %w", err)
	}

	return transaction, nil
}

func (s *Service) GetAllTransactions() ([]*domain.Transaction, error) {
	transactions, err := s.repository.FindAll()
	if err != nil {
		return nil, fmt.Errorf("get all transactions: %w", err)
	}

	return transactions, nil
}

func (s *Service) GetTransactionsByUserID(userID uuid.UUID, from *time.Time, to *time.Time, fraudOnly bool) ([]*domain.Transaction, error) {
	if (from == nil) != (to == nil) {
		return nil, fmt.Errorf("get transactions by user id: %w", ErrInvalidTimeRange)
	}
	if from != nil && to.Before(*from) {
		return nil, fmt.Errorf("get transactions by user id: %w", ErrInvalidTimeRange)
	}

	transactions, err := s.repository.FindByUserID(userID, from, to, fraudOnly)
	if err != nil {
		return nil, fmt.Errorf("get transactions by user id: %w", err)
	}

	return transactions, nil
}

func (s *Service) GetUserTransactionSummary(
	userID uuid.UUID,
	from *time.Time,
	to *time.Time,
	bucket string,
	fraudOnly bool,
) ([]UserTransactionBucketSummary, error) {
	if (from == nil) != (to == nil) {
		return nil, fmt.Errorf("get user transaction summary: %w", ErrInvalidTimeRange)
	}

	now := time.Now().UTC()
	var effFrom, effTo time.Time
	allTime := from == nil && to == nil
	if allTime {
		effFrom = time.Unix(0, 0).UTC()
		effTo = now
	} else {
		effFrom = *from
		effTo = *to
		if effTo.Before(effFrom) {
			return nil, fmt.Errorf("get user transaction summary: %w", ErrInvalidTimeRange)
		}
	}

	normalizedBucket := strings.ToLower(strings.TrimSpace(bucket))
	if normalizedBucket == "" {
		if allTime {
			normalizedBucket = "day"
		} else {
			normalizedBucket = "minute"
		}
	}

	switch normalizedBucket {
	case "minute", "hour", "day":
	default:
		return nil, fmt.Errorf("get user transaction summary: invalid bucket %q: %w", bucket, ErrInvalidBucket)
	}

	summary, err := s.repository.GetUserTransactionSummary(userID, effFrom, effTo, normalizedBucket, fraudOnly)
	if err != nil {
		return nil, fmt.Errorf("get user transaction summary: %w", err)
	}

	return summary, nil
}

func (s *Service) GetPreviousTransactionByUserID(userID uuid.UUID, before time.Time) (*domain.Transaction, bool, error) {
	transaction, found, err := s.repository.FindPreviousByUserID(userID, before)
	if err != nil {
		return nil, false, fmt.Errorf("get previous transaction by user id: %w", err)
	}

	return transaction, found, nil
}

func (s *Service) GetTripleAmountAverageLast24HoursByUserID(userID uuid.UUID, excludeTxID *uuid.UUID) (int64, error) {
	tripleAverage, err := s.repository.GetTripleAmountAverageLast24HoursByUserID(userID, excludeTxID)
	if err != nil {
		return 0, fmt.Errorf("get triple amount average last 24 hours by user id: %w", err)
	}

	return tripleAverage, nil
}

func (s *Service) GetLastLocationByUserID(ctx context.Context, userID uuid.UUID) (string, error) {
	if s.cache != nil {
		location, found, err := s.cache.GetLastLocationByUserID(ctx, userID)
		if err == nil {
			if found {
				return location, nil
			}
		} else {
			log.Printf("get last location cache: %v", err)
		}
	}

	location, err := s.repository.GetLastLocationByUserID(userID)
	if err != nil {
		return "", fmt.Errorf("get last location by user transaction id: %w", err)
	}

	if location != "" {
		if err := s.SetLastLocationByUserID(ctx, userID, location); err != nil {
			log.Printf("set last location cache after db fetch: %v", err)
		}
	}

	return location, nil
}

func (s *Service) SetLastLocationByUserID(ctx context.Context, userID uuid.UUID, location string) error {
	if s.cache == nil {
		return nil
	}
	if err := s.cache.SetLastLocationByUserID(ctx, userID, location); err != nil {
		return fmt.Errorf("set last location by user id: %w", err)
	}

	return nil
}

func (s *Service) DeleteTransaction(id uuid.UUID) error {
	if err := s.repository.Delete(id); err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}

	return nil
}

func (s *Service) PublishTransactionEvaluated(ctx context.Context, transaction *domain.Transaction, evaluatedAt time.Time) error {
	if transaction == nil {
		return fmt.Errorf("publish transaction evaluated: %w", errors.New("nil transaction"))
	}
	if s.evaluatedPublisher == nil {
		return nil
	}

	evt := TransactionEvaluatedEvent{
		TransactionID: transaction.ID.String(),
		UserID:        transaction.UserID.String(),
		UserName:      transaction.User.Name,
		UserSurname:   transaction.User.Surname,
		Amount:        transaction.Amount,
		Location:      transaction.Location,
		FraudStatus:   transaction.FraudStatus,
		CreatedAt:     transaction.CreatedAt,
		EvaluatedAt:   evaluatedAt,
	}

	if err := s.evaluatedPublisher.PublishTransactionEvaluated(ctx, evt); err != nil {
		return fmt.Errorf("publish transaction evaluated: %w", err)
	}

	return nil
}
