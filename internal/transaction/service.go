package transaction

import (
	"context"
	"errors"
	"fmt"
	"log"

	"fraud-detection/internal/domain"
	"github.com/google/uuid"
)

type transactionCreatedPublisher interface {
	PublishTransactionCreated(ctx context.Context, transaction *domain.Transaction) error
}

type Service struct {
	repository RepositoryInterface
	publisher  transactionCreatedPublisher
}

func NewService(repository RepositoryInterface, publisher transactionCreatedPublisher) *Service {
	return &Service{repository: repository, publisher: publisher}
}

func (s *Service) CreateTransaction(ctx context.Context, transaction *domain.Transaction) error {
	if transaction == nil {
		return fmt.Errorf("create transaction: %w", errors.New("nil transaction"))
	}

	if err := s.repository.Create(transaction); err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}

	if s.publisher != nil {
		if err := s.publisher.PublishTransactionCreated(ctx, transaction); err != nil {
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

func (s *Service) GetTransactionsByUserID(userID uuid.UUID) ([]*domain.Transaction, error) {
	transactions, err := s.repository.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("get transactions by user id: %w", err)
	}

	return transactions, nil
}

func (s *Service) DeleteTransaction(id uuid.UUID) error {
	if err := s.repository.Delete(id); err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}

	return nil
}