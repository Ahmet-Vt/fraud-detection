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
	FindByUserID(userID uuid.UUID, from *time.Time, to *time.Time, fraudOnly bool) ([]*domain.Transaction, error)
	GetUserTransactionSummary(
		userID uuid.UUID,
		from time.Time,
		to time.Time,
		bucket string,
		fraudOnly bool,
	) ([]UserTransactionBucketSummary, error)
	FindPreviousByUserID(userID uuid.UUID, before time.Time) (*domain.Transaction, bool, error)
	GetTripleAmountAverageLast24HoursByUserID(userID uuid.UUID, excludeTxID *uuid.UUID) (int64, error)
	Delete(id uuid.UUID) error
	FindFraudByCreatedAtBetween(from time.Time, to time.Time, limit int) ([]*domain.Transaction, error)
}

type Repository struct {
	db *gorm.DB
}

type UserTransactionBucketSummary struct {
	BucketStart  time.Time `gorm:"column:bucket_start"`
	RequestCount int64     `gorm:"column:request_count"`
	AnomalyCount int64     `gorm:"column:anomaly_count"`
	AmountSum    int64     `gorm:"column:amount_sum"`
	AmountAvg    float64   `gorm:"column:amount_avg"`
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

		var userIDs []uuid.UUID
		if err := tx.Model(&domain.Transaction{}).
			Where("id = ?", transactionID).
			Limit(1).
			Pluck("user_id", &userIDs).Error; err != nil {
			return fmt.Errorf("update fraud rules: %w", err)
		}
		if len(userIDs) == 0 {
			return fmt.Errorf("update fraud rules: %w", ErrTransactionNotFound)
		}
		userID := userIDs[0]

		if err := tx.Exec(`
			UPDATE users SET risk_rate = COALESCE((
				SELECT LEAST(100::float8, GREATEST(0::float8,
					(100.0 * COUNT(*) FILTER (WHERE fraud_status)) / NULLIF(COUNT(*)::float8, 0)
				))
				FROM transactions WHERE user_id = ?
			), 0)
			WHERE id = ?
		`, userID, userID).Error; err != nil {
			return fmt.Errorf("update fraud rules: %w", err)
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
	if err := r.db.Preload("FraudRules").Preload("User").First(&transaction, id).Error; err != nil {
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

func (r *Repository) FindByUserID(userID uuid.UUID, from *time.Time, to *time.Time, fraudOnly bool) ([]*domain.Transaction, error) {
	var transactions []*domain.Transaction
	q := r.db.Preload("FraudRules").Where("user_id = ?", userID)
	if fraudOnly {
		q = q.Where("fraud_status = ?", true)
	}
	if from != nil {
		q = q.Where("created_at >= ?", *from)
	}
	if to != nil {
		q = q.Where("created_at <= ?", *to)
	}
	if err := q.Order("created_at DESC").Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("find transactions by user id: %w", err)
	}
	return transactions, nil
}

func (r *Repository) GetUserTransactionSummary(
	userID uuid.UUID,
	from time.Time,
	to time.Time,
	bucket string,
	fraudOnly bool,
) ([]UserTransactionBucketSummary, error) {
	summaries := make([]UserTransactionBucketSummary, 0)
	whereFraud := ""
	if fraudOnly {
		whereFraud = " AND fraud_status = TRUE"
	}
	query := `SELECT
			date_trunc(?, created_at) AS bucket_start,
			COUNT(*) AS request_count,
			COALESCE(SUM(CASE WHEN fraud_status THEN 1 ELSE 0 END), 0) AS anomaly_count,
			COALESCE(SUM(amount), 0) AS amount_sum,
			COALESCE(AVG(amount), 0) AS amount_avg
		FROM transactions
		WHERE user_id = ? AND created_at >= ? AND created_at <= ?` + whereFraud + `
		GROUP BY 1
		ORDER BY 1 ASC`
	if err := r.db.Raw(
		query,
		bucket,
		userID,
		from,
		to,
	).Scan(&summaries).Error; err != nil {
		return nil, fmt.Errorf("get user transaction summary: %w", err)
	}

	return summaries, nil
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

func (r *Repository) FindFraudByCreatedAtBetween(from time.Time, to time.Time, limit int) ([]*domain.Transaction, error) {
	var transactions []*domain.Transaction
	q := r.db.Preload("FraudRules").
		Where("fraud_status = ?", true).
		Where("created_at >= ? AND created_at <= ?", from, to).
		Order("created_at DESC").
		Limit(limit)
	if err := q.Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("find fraud by created at between: %w", err)
	}
	return transactions, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
