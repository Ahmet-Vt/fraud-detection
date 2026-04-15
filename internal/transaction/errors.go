package transaction

import "errors"

var (
	ErrTransactionNotFound      = errors.New("transaction not found")
	ErrTransactionAlreadyExists = errors.New("transaction already exists")
	ErrUserNotExists            = errors.New("referenced user does not exist")
)
