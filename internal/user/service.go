package user

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

func (s *Service) CreateUser(user *domain.User) error {
	if user == nil {
		return fmt.Errorf("create user: %w", errors.New("nil user"))
	}

	if err := s.repository.Create(user); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (s *Service) GetUserByID(id uuid.UUID) (*domain.User, error) {
	user, err := s.repository.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func (s *Service) GetAllUsers() ([]*domain.User, error) {
	users, err := s.repository.FindAll()
	if err != nil {
		return nil, fmt.Errorf("get all users: %w", err)
	}

	return users, nil
}
