package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserStatus string
type Status string

const (
	TransactionStatusPending Status = "pending"
	TransactionStatusCompleted Status = "completed"
	TransactionStatusFailed Status = "failed"
)

const (
	UserStatusActive   UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusDeleted UserStatus = "deleted"
)

type User struct {
	ID           uuid.UUID     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name         string        `json:"name" gorm:"type:varchar(100);not null"`
	Surname      string        `json:"surname" gorm:"type:varchar(100);not null"`
	Email        *string       `json:"email" gorm:"type:varchar(255);uniqueIndex"`
	UserStatus   UserStatus    `json:"status" gorm:"column:status;type:user_status;not null"`
	Transactions []Transaction `json:"transactions" gorm:"foreignKey:UserID"`
	CreatedAt    time.Time     `json:"created_at" gorm:"not null;default:now()"`
}

type Transaction struct {
	ID                  uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID              uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	User                User      `json:"user" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Amount              int64     `json:"amount" gorm:"type:bigint;not null"`
	City                string    `json:"city" gorm:"type:varchar(100);not null"`
	AmountRuleMatched   bool      `json:"amount_rule_matched" gorm:"not null;default:false"`
	LocationRuleMatched bool      `json:"location_rule_matched" gorm:"not null;default:false"`
	VelocityRuleMatched bool      `json:"velocity_rule_matched" gorm:"not null;default:false"`
	TransactionStatus   Status    `json:"transaction_status" gorm:"type:status;not null;default:pending"`
	FraudStatus         bool      `json:"fraud_status" gorm:"not null;default:false;index"`
	CreatedAt           time.Time `json:"created_at" gorm:"not null;default:now()"`
}

func (t *Transaction) BeforeCreate(_ *gorm.DB) error {
	t.RefreshFraudStatus()
	return nil
}

func (t *Transaction) BeforeUpdate(_ *gorm.DB) error {
	t.RefreshFraudStatus()
	return nil
}

func (t *Transaction) RefreshFraudStatus() {
	matchedRules := 0

	if t.AmountRuleMatched {
		matchedRules++
	}
	if t.LocationRuleMatched {
		matchedRules++
	}
	if t.VelocityRuleMatched {
		matchedRules++
	}

	t.FraudStatus = matchedRules >= 2
}
