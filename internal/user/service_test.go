package user

import (
	"errors"
	"testing"

	"github.com/Ahmet-Vt/fraud-detection/internal/domain"
	"github.com/google/uuid"
)

type stubRepository struct {
	createErr  error
	findUser   *domain.User
	findErr    error
	findAll    []*domain.User
	findAllErr error
}

func (s stubRepository) Create(_ *domain.User) error {
	return s.createErr
}

func (s stubRepository) FindByID(_ uuid.UUID) (*domain.User, error) {
	return s.findUser, s.findErr
}

func (s stubRepository) FindAll() ([]*domain.User, error) {
	return s.findAll, s.findAllErr
}

func TestCreateUserNilReturnsError(t *testing.T) {
	svc := NewService(stubRepository{})

	err := svc.CreateUser(nil)
	if err == nil {
		t.Fatal("expected error for nil user")
	}
}

func TestCreateUserAlreadyExistsWraps(t *testing.T) {
	svc := NewService(stubRepository{
		createErr: ErrUserAlreadyExists,
	})

	err := svc.CreateUser(&domain.User{Name: "test"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists in chain, got %v", err)
	}
}

func TestGetUserByIDNotFoundWraps(t *testing.T) {
	svc := NewService(stubRepository{
		findErr: ErrUserNotFound,
	})

	_, err := svc.GetUserByID(uuid.New())
	if err == nil {
		t.Fatal("expected error for missing user")
	}
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound in chain, got %v", err)
	}
}

func TestGetUserByIDSuccess(t *testing.T) {
	expected := &domain.User{Name: "Ahmet"}
	svc := NewService(stubRepository{
		findUser: expected,
	})

	user, err := svc.GetUserByID(uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Name != expected.Name {
		t.Fatalf("expected name %q, got %q", expected.Name, user.Name)
	}
}

func TestGetAllUsersReturnsError(t *testing.T) {
	svc := NewService(stubRepository{
		findAllErr: errors.New("db connection lost"),
	})

	_, err := svc.GetAllUsers()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetAllUsersSuccess(t *testing.T) {
	expected := []*domain.User{{Name: "A"}, {Name: "B"}}
	svc := NewService(stubRepository{
		findAll: expected,
	})

	users, err := svc.GetAllUsers()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != len(expected) {
		t.Fatalf("expected %d users, got %d", len(expected), len(users))
	}
}
