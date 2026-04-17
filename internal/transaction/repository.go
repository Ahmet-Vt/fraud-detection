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
	ReplaceFraudRules(transactionID uuid.UUID, rules []domain.TransactionFraudRule) error
	FindByID(id uuid.UUID) (*domain.Transaction, error)
	FindAll() ([]*domain.Transaction, error)
	FindByUserID(userID uuid.UUID) ([]*domain.Transaction, error)
	GetTripleAmountAverageLast24HoursByUserID(userID uuid.UUID) (int64, error)
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
		if isForeignKeyViolation(err) {
			return fmt.Errorf("create transaction: %w", errors.Join(ErrUserNotExists, err))
		}
		return fmt.Errorf("create transaction: %w", err)
	}
	return nil
}

func (r *Repository) ReplaceFraudRules(transactionID uuid.UUID, rules []domain.TransactionFraudRule) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		normalizedRules := make([]domain.TransactionFraudRule, 0, len(rules))
		matchedRules := 0

		for _, rule := range rules {
			rule.TransactionID = transactionID
			if rule.Matched {
				matchedRules++
			}
			normalizedRules = append(normalizedRules, rule)
		}

		if err := tx.Where("transaction_id = ?", transactionID).Delete(&domain.TransactionFraudRule{}).Error; err != nil {
			return fmt.Errorf("replace fraud rules: %w", err)
		}

		if len(normalizedRules) > 0 {
			if err := tx.Create(&normalizedRules).Error; err != nil {
				return fmt.Errorf("replace fraud rules: %w", err)
			}
		}

		result := tx.Model(&domain.Transaction{}).
			Where("id = ?", transactionID).
			Update("fraud_status", matchedRules >= 2)
		if result.Error != nil {
			return fmt.Errorf("replace fraud rules: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("replace fraud rules: %w", ErrTransactionNotFound)
		}

		return nil
	})
}

func (r *Repository) FindByID(id uuid.UUID) (*domain.Transaction, error) {
	var transaction domain.Transaction
	if err := r.db.Preload("FraudRules").First(&transaction, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find transaction by id: %w", errors.Join(ErrTransactionNotFound, err))
		}
		return nil, fmt.Errorf("find transaction by id: %w", err)
	}
	return &transaction, nil
}

func (r *Repository) FindAll() ([]*domain.Transaction, error) {
	var transactions []*domain.Transaction
	if err := r.db.Preload("FraudRules").Order("created_at DESC").Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("find all transactions: %w", err)
	}
	return transactions, nil
}

func (r *Repository) FindByUserID(userID uuid.UUID) ([]*domain.Transaction, error) {
	var transactions []*domain.Transaction
	if err := r.db.Preload("FraudRules").Where("user_id = ?", userID).Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("find transactions by user id: %w", err)
	}
	return transactions, nil
}

func (r *Repository) GetTripleAmountAverageLast24HoursByUserID(userID uuid.UUID) (int64, error) {
	var tripleAmountAverage int64

	if err := r.db.Model(&domain.Transaction{}).
		Select("COALESCE(AVG(amount), 0) * 3").
		Where("user_id = ? AND created_at >= NOW() - INTERVAL '24 hours'", userID).
		Scan(&tripleAmountAverage).Error; err != nil {
		return 0, fmt.Errorf("get triple amount average last 24 hours by user id: %w", err)
	}

	return tripleAmountAverage, nil
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

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}