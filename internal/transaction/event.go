package transaction

import "time"

type TransactionCreatedEvent struct {
	TransactionID string
	UserID        string
	Amount        int64
	Location      string
	CreatedAt     time.Time
}
