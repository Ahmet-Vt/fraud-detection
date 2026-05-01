package apidoc

type InvalidRequestBodyError struct {
	Code  string `json:"code" example:"INVALID_REQUEST_BODY" extensions:"x-order=0"`
	Error string `json:"error" example:"invalid request body" extensions:"x-order=1"`
}

type InvalidUserIDError struct {
	Code  string `json:"code" example:"INVALID_USER_ID" extensions:"x-order=0"`
	Error string `json:"error" example:"invalid user id" extensions:"x-order=1"`
}

type UserNotFoundError struct {
	Code  string `json:"code" example:"USER_NOT_FOUND" extensions:"x-order=0"`
	Error string `json:"error" example:"user not found" extensions:"x-order=1"`
}

type UserAlreadyExistsError struct {
	Code  string `json:"code" example:"USER_ALREADY_EXISTS" extensions:"x-order=0"`
	Error string `json:"error" example:"user already exists" extensions:"x-order=1"`
}

type InvalidTransactionIDError struct {
	Code  string `json:"code" example:"INVALID_TRANSACTION_ID" extensions:"x-order=0"`
	Error string `json:"error" example:"invalid transaction id" extensions:"x-order=1"`
}

type TransactionNotFoundError struct {
	Code  string `json:"code" example:"TRANSACTION_NOT_FOUND" extensions:"x-order=0"`
	Error string `json:"error" example:"transaction not found" extensions:"x-order=1"`
}

type TransactionAlreadyExistsError struct {
	Code  string `json:"code" example:"TRANSACTION_ALREADY_EXISTS" extensions:"x-order=0"`
	Error string `json:"error" example:"transaction already exists" extensions:"x-order=1"`
}

type ReferencedUserNotExistsError struct {
	Code  string `json:"code" example:"USER_NOT_EXISTS" extensions:"x-order=0"`
	Error string `json:"error" example:"referenced user does not exist" extensions:"x-order=1"`
}

type InternalServerError struct {
	Code  string `json:"code" example:"INTERNAL_ERROR" extensions:"x-order=0"`
	Error string `json:"error" example:"Internal Server Error" extensions:"x-order=1"`
}
