package transaction

import "errors"

var (
	ErrTransactionNotFound      = errors.New("transaction not found")
	ErrTransactionAlreadyExists = errors.New("transaction already exists")
	ErrUserNotExists            = errors.New("referenced user does not exist")
	ErrInvalidTimeRange         = errors.New("invalid time range")
	ErrInvalidBucket            = errors.New("invalid bucket")
)
