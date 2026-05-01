package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserStatus string
type Status string
type FraudRuleName string

const (
	TransactionStatusPending   Status = "pending"
	TransactionStatusCompleted Status = "completed"
	TransactionStatusFailed    Status = "failed"
)

const (
	FraudRuleAmount   FraudRuleName = "amount"
	FraudRuleLocation FraudRuleName = "location"
	FraudRuleVelocity FraudRuleName = "velocity"
)

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusDeleted   UserStatus = "deleted"
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
	ID                uuid.UUID              `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID            uuid.UUID              `json:"user_id" gorm:"type:uuid;not null;index"`
	User              User                   `json:"user" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Amount            int64                  `json:"amount" gorm:"type:bigint;not null"`
	Location          string                 `json:"location" gorm:"type:varchar(100);not null"`
	FraudRules        []TransactionFraudRule `json:"fraud_rules,omitempty" gorm:"foreignKey:TransactionID"`
	TransactionStatus Status                 `json:"transaction_status" gorm:"type:status;not null;default:pending"`
	FraudStatus       bool                   `json:"fraud_status" gorm:"not null;default:false;index"`
	CreatedAt         time.Time              `json:"created_at" gorm:"not null;default:now()"`
}

type TransactionFraudRule struct {
	ID            uuid.UUID     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TransactionID uuid.UUID     `json:"transaction_id" gorm:"type:uuid;not null;index"`
	RuleName      FraudRuleName `json:"rule_name" gorm:"type:fraud_rule_name;not null"`
	Matched       bool          `json:"matched" gorm:"not null;default:false"`
	CreatedAt     time.Time     `json:"created_at" gorm:"not null;default:now()"`
}

func (t *Transaction) RefreshFraudStatus() {
	matchedRules := 0

	for _, fraudRule := range t.FraudRules {
		if fraudRule.Matched {
			matchedRules++
		}
	}

	t.FraudStatus = matchedRules >= 2
}
