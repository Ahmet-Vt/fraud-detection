package transaction

import "time"

type TransactionCreatedEvent struct {
	TransactionID string
	UserID        string
	Amount        int64
	Location      string
	CreatedAt     time.Time
}

type TransactionEvaluatedEvent struct {
	TransactionID     string    `json:"transaction_id"`
	UserID            string    `json:"user_id"`
	UserName          string    `json:"user_name"`
	UserSurname       string    `json:"user_surname"`
	Amount            int64     `json:"amount"`
	Location          string    `json:"location"`
	TransactionStatus string    `json:"transaction_status"`
	FraudStatus       bool      `json:"fraud_status"`
	CreatedAt         time.Time `json:"created_at"`
	EvaluatedAt       time.Time `json:"evaluated_at"`
}
