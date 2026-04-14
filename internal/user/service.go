package user 

import (
	"github.com/Ahmet-Vt/fraud-detection/internal/domain"
	"github.com/google/uuid"
	"log"
)

type Service struct {
	repository RepositoryInterface
}

func NewService(repository RepositoryInterface) *Service {
	return &Service{repository: repository}
}

func (s *Service) CreateUser(user *domain.User) error {
	log.Printf("Creating user: %+v", user)
	if err := s.repository.Create(user); err != nil {
		return err
	}
	log.Printf("User created successfully: %+v", user)

	return nil
}

func (s *Service) GetUserByID(id uuid.UUID) (*domain.User, error) {
	log.Printf("Getting user by ID: %+v", id)
	user, err := s.repository.FindByID(id)
	if err != nil {
		return nil, err
	}
	log.Printf("User found: %+v", user)
	return user, nil
}

func (s *Service) GetAllUsers() ([]*domain.User, error) {
	log.Printf("Getting all users")
	users, err := s.repository.FindAll()
	if err != nil {
		return nil, err
	}
	log.Printf("Users found: %+v", users)
	return users, nil
}
