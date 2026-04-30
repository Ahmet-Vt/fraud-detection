package migrations

import (
	"fraud-detection/internal/database"

	"gorm.io/gorm"
)

var CreateTransactionsTable = database.MigrationEntry{
	Version: "20260414_002",
	Name:    "create_transactions_table",
	Up: func(tx *gorm.DB) error {
		if err := tx.Exec(`
			DO $$
			BEGIN
				IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'status') THEN
					CREATE TYPE status AS ENUM ('pending', 'completed', 'failed');
				END IF;
			END
			$$;
		`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
			CREATE TABLE transactions (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				user_id UUID NOT NULL,
				amount BIGINT NOT NULL,
				location VARCHAR(100) NOT NULL,
				transaction_status status NOT NULL DEFAULT 'pending',
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
		if err := tx.Exec(`DROP TABLE IF EXISTS transactions`).Error; err != nil {
			return err
		}

		return tx.Exec(`DROP TYPE IF EXISTS status`).Error
	},
}
