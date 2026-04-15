package transaction

import (
	"errors"
	"fmt"

	"github.com/Ahmet-Vt/fraud-detection/internal/domain"
	"github.com/google/uuid"
)

type Service struct {
	repository RepositoryInterface
}

func NewService(repository RepositoryInterface) *Service {
	return &Service{repository: repository}
}

func (s *Service) CreateTransaction(transaction *domain.Transaction) error {
	if transaction == nil {
		return fmt.Errorf("create transaction: %w", errors.New("nil transaction"))
	}

	if err := s.repository.Create(transaction); err != nil {
		return fmt.Errorf("create transaction: %w", err)
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