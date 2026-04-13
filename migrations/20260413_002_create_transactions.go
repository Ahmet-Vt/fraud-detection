package migrations

import (
	"github.com/Ahmet-Vt/fraud-detection/internal/database"

	"gorm.io/gorm"
)

var CreateTransactionsTable = database.MigrationEntry{
	Version: "20260413_002",
	Name:    "create_transactions_table",
	Up: func(tx *gorm.DB) error {
		if err := tx.Exec(`
			CREATE TABLE transactions (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				user_id UUID NOT NULL,
				amount BIGINT NOT NULL,
				latitude DOUBLE PRECISION NOT NULL,
				longitude DOUBLE PRECISION NOT NULL,
				city VARCHAR(100) NOT NULL,
				amount_rule_matched BOOLEAN NOT NULL DEFAULT FALSE,
				location_rule_matched BOOLEAN NOT NULL DEFAULT FALSE,
				velocity_rule_matched BOOLEAN NOT NULL DEFAULT FALSE,
				fraud_status BOOLEAN NOT NULL DEFAULT FALSE,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				CONSTRAINT fk_transactions_user
					FOREIGN KEY (user_id) REFERENCES users(id)
					ON UPDATE CASCADE
					ON DELETE CASCADE
			)
		`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`CREATE INDEX idx_transactions_user_id ON transactions(user_id)`).Error; err != nil {
			return err
		}

		return tx.Exec(`CREATE INDEX idx_transactions_fraud_status ON transactions(fraud_status)`).Error
	},
	Down: func(tx *gorm.DB) error {
		return tx.Exec(`DROP TABLE IF EXISTS transactions`).Error
	},
}
