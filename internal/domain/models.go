package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name         string        `gorm:"type:varchar(100);not null"`
	Surname      string        `gorm:"type:varchar(100);not null"`
	Email        string        `gorm:"type:varchar(255);uniqueIndex;not null"`
	CreatedAt    time.Time     `gorm:"not null;default:now()"`
	Transactions []Transaction `gorm:"foreignKey:UserID"`
}

type Transaction struct {
	ID                  uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID              uuid.UUID `gorm:"type:uuid;not null;index"`
	User                User      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Amount              int64     `gorm:"type:bigint;not null"`
	Latitude            float64   `gorm:"type:double precision;not null"`
	Longitude           float64   `gorm:"type:double precision;not null"`
	City                string    `gorm:"type:varchar(100);not null"`
	Status              string    `gorm:"type:varchar(50);not null"`
	AmountRuleMatched   bool      `gorm:"not null;default:false"`
	LocationRuleMatched bool      `gorm:"not null;default:false"`
	VelocityRuleMatched bool      `gorm:"not null;default:false"`
	FraudStatus         bool      `gorm:"not null;default:false;index"`
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
