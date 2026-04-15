package transaction

import (
	"errors"
	"fmt"

	"github.com/Ahmet-Vt/fraud-detection/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type RepositoryInterface interface {
	Create(transaction *domain.Transaction) error
	FindByID(id uuid.UUID) (*domain.Transaction, error)
	FindAll() ([]*domain.Transaction, error)
	FindByUserID(userID uuid.UUID) ([]*domain.Transaction, error)
	Delete(id uuid.UUID) error
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) RepositoryInterface {
	return &Repository{db: db}
}

func (r *Repository) Create(transaction *domain.Transaction) error {
	if err := r.db.Create(transaction).Error; err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("create transaction: %w", errors.Join(ErrTransactionAlreadyExists, err))
		}
		return fmt.Errorf("create transaction: %w", err)
	}
	return nil
}

func (r *Repository) FindByID(id uuid.UUID) (*domain.Transaction, error) {
	var transaction domain.Transaction
	if err := r.db.First(&transaction, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find transaction by id: %w", errors.Join(ErrTransactionNotFound, err))
		}
		return nil, fmt.Errorf("find transaction by id: %w", err)
	}
	return &transaction, nil
}

func (r *Repository) FindAll() ([]*domain.Transaction, error) {
	var transactions []*domain.Transaction
	if err := r.db.Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("find all transactions: %w", err)
	}
	return transactions, nil
}

func (r *Repository) FindByUserID(userID uuid.UUID) ([]*domain.Transaction, error) {
	var transactions []*domain.Transaction
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("find transactions by user id: %w", err)
	}
	return transactions, nil
}

func (r *Repository) Delete(id uuid.UUID) error {
	result := r.db.Delete(&domain.Transaction{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete transaction: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("delete transaction: %w", ErrTransactionNotFound)
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}