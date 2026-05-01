package user

import "errors"

var (
	ErrInvalidUserID     = errors.New("invalid user id")
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)
