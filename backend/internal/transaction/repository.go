package transaction

import (
	"errors"
	"fmt"
	"fraud-detection/internal/domain"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type RepositoryInterface interface {
	Create(transaction *domain.Transaction) error
	UpdateFraudRules(transactionID uuid.UUID, rules []domain.TransactionFraudRule) error
	UpdateStatus(transactionID uuid.UUID, status domain.Status) error
	FindByID(id uuid.UUID) (*domain.Transaction, error)
	FindAll() ([]*domain.Transaction, error)
	FindByUserID(userID uuid.UUID) ([]*domain.Transaction, error)
	FindPreviousByUserID(userID uuid.UUID, before time.Time) (*domain.Transaction, bool, error)
	GetTripleAmountAverageLast24HoursByUserID(userID uuid.UUID, excludeTxID *uuid.UUID) (int64, error)
	Delete(id uuid.UUID) error
	GetLastLocationByUserID(userID uuid.UUID) (string, error)
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

func (r *Repository) UpdateFraudRules(transactionID uuid.UUID, rules []domain.TransactionFraudRule) error {
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
			return fmt.Errorf("update fraud rules: %w", err)
		}

		if len(normalizedRules) > 0 {
			if err := tx.Create(&normalizedRules).Error; err != nil {
				return fmt.Errorf("update fraud rules: %w", err)
			}
		}

		result := tx.Model(&domain.Transaction{}).
			Where("id = ?", transactionID).
			Updates(map[string]any{
				"fraud_status":       matchedRules >= 2,
				"transaction_status": domain.TransactionStatusCompleted,
			})
		if result.Error != nil {
			return fmt.Errorf("update fraud rules: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("update fraud rules: %w", ErrTransactionNotFound)
		}

		return nil
	})
}

func (r *Repository) UpdateStatus(transactionID uuid.UUID, status domain.Status) error {
	result := r.db.Model(&domain.Transaction{}).
		Where("id = ?", transactionID).
		Update("transaction_status", status)
	if result.Error != nil {
		return fmt.Errorf("update transaction status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("update transaction status: %w", ErrTransactionNotFound)
	}

	return nil
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

func (r *Repository) FindPreviousByUserID(userID uuid.UUID, before time.Time) (*domain.Transaction, bool, error) {
	var transaction domain.Transaction
	if err := r.db.
		Where("user_id = ? AND created_at < ?", userID, before).
		Order("created_at DESC").
		Take(&transaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("find previous transaction by user id: %w", err)
	}

	return &transaction, true, nil
}

func (r *Repository) GetTripleAmountAverageLast24HoursByUserID(userID uuid.UUID, excludeTxID *uuid.UUID) (int64, error) {
	var tripleAmountAverage int64

	query := r.db.Model(&domain.Transaction{}).
		Select("COALESCE((AVG(amount) * 3)::bigint, 0)").
		Where("user_id = ? AND created_at >= NOW() - INTERVAL '24 hours' AND fraud_status = ?", userID, false)
	if excludeTxID != nil {
		query = query.Where("id <> ?", *excludeTxID)
	}

	if err := query.Scan(&tripleAmountAverage).Error; err != nil {
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

func (r *Repository) GetLastLocationByUserID(userID uuid.UUID) (string, error) {
	var transaction domain.Transaction
	if err := r.db.Model(&domain.Transaction{}).
		Select("location").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Take(&transaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", fmt.Errorf("get last location by user transaction id: %w", err)
	}

	return transaction.Location, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
