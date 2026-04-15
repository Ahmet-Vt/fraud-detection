package user

import (
	"errors"
	"fmt"

	"github.com/Ahmet-Vt/fraud-detection/internal/domain"
	"github.com/Ahmet-Vt/fraud-detection/internal/secureerr"
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
		return secureerr.Internal(
			"USER_CREATE_FAILED",
			"failed to create user",
			errors.New("create user service: nil user"),
			nil,
		)
	}

	if err := s.repository.Create(user); err != nil {
		wrappedErr := fmt.Errorf("create user service: %w", err)
		if errors.Is(err, ErrUserAlreadyExists) {
			return secureerr.Conflict(
				"USER_ALREADY_EXISTS",
				"user already exists",
				wrappedErr,
				userMetadata(user),
			)
		}

		return secureerr.Internal(
			"USER_CREATE_FAILED",
			"failed to create user",
			wrappedErr,
			userMetadata(user),
		)
	}

	return nil
}

func (s *Service) GetUserByID(id string) (*domain.User, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, secureerr.BadRequest(
			"INVALID_USER_ID",
			"invalid user id",
			fmt.Errorf("parse user id: %w", errors.Join(ErrInvalidUserID, err)),
			map[string]string{"user_id": id},
		)
	}

	user, err := s.repository.FindByID(parsedID)
	if err != nil {
		wrappedErr := fmt.Errorf("get user by id service: %w", err)
		if errors.Is(err, ErrUserNotFound) {
			return nil, secureerr.NotFound(
				"USER_NOT_FOUND",
				"user not found",
				wrappedErr,
				map[string]string{"user_id": parsedID.String()},
			)
		}

		return nil, secureerr.Internal(
			"USER_RETRIEVAL_FAILED",
			"failed to retrieve user",
			wrappedErr,
			map[string]string{"user_id": parsedID.String()},
		)
	}

	return user, nil
}

func (s *Service) GetAllUsers() ([]*domain.User, error) {
	users, err := s.repository.FindAll()
	if err != nil {
		return nil, secureerr.Internal(
			"USER_LIST_FAILED",
			"failed to retrieve users",
			fmt.Errorf("get all users service: %w", err),
			nil,
		)
	}

	return users, nil
}

func userMetadata(user *domain.User) map[string]string {
	if user == nil {
		return nil
	}

	metadata := map[string]string{
		"user_status": string(user.UserStatus),
	}

	if user.ID != uuid.Nil {
		metadata["user_id"] = user.ID.String()
	}
	if user.Email != nil && *user.Email != "" {
		metadata["has_email"] = "true"
	}

	return metadata
}
