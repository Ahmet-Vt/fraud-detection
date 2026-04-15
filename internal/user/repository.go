package user

import (
	"errors"
	"fmt"

	"github.com/Ahmet-Vt/fraud-detection/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type RepositoryInterface interface {
	Create(user *domain.User) error
	FindByID(id uuid.UUID) (*domain.User, error)
	FindAll() ([]*domain.User, error)
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) RepositoryInterface {
	return &Repository{db: db}
}

func (r *Repository) Create(user *domain.User) error {
	if err := r.db.Create(user).Error; err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("create user: %w", errors.Join(ErrUserAlreadyExists, err))
		}
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (r *Repository) FindByID(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	if err := r.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find user by id: %w", errors.Join(ErrUserNotFound, err))
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &user, nil
}

func (r *Repository) FindAll() ([]*domain.User, error) {
	var users []*domain.User
	if err := r.db.Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("find all users: %w", err)
	}
	return users, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
