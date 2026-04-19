package migrations

import (
	"fraud-detection/internal/database"

	"gorm.io/gorm"
)

var CreateTransactionFraudRulesTable = database.MigrationEntry{
	Version: "20260417_001",
	Name:    "create_transaction_fraud_rules_table",
	Up: func(tx *gorm.DB) error {
		if err := tx.Exec(`
			DO $$
			BEGIN
				IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'fraud_rule_name') THEN
					CREATE TYPE fraud_rule_name AS ENUM ('amount', 'location', 'velocity');
				END IF;
			END
			$$;
		`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
			CREATE TABLE transaction_fraud_rules (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				transaction_id UUID NOT NULL,
				rule_name fraud_rule_name NOT NULL,
				matched BOOLEAN NOT NULL DEFAULT FALSE,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				CONSTRAINT fk_transaction_fraud_rules_transaction
					FOREIGN KEY (transaction_id) REFERENCES transactions(id)
					ON UPDATE CASCADE
					ON DELETE CASCADE,
				CONSTRAINT uq_transaction_fraud_rules_transaction_rule
					UNIQUE (transaction_id, rule_name)
			)
		`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`CREATE INDEX idx_transaction_fraud_rules_transaction_id ON transaction_fraud_rules(transaction_id)`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
			INSERT INTO transaction_fraud_rules (transaction_id, rule_name, matched)
			SELECT id, 'amount'::fraud_rule_name, amount_rule_matched FROM transactions
			UNION ALL
			SELECT id, 'location'::fraud_rule_name, location_rule_matched FROM transactions
			UNION ALL
			SELECT id, 'velocity'::fraud_rule_name, velocity_rule_matched FROM transactions
		`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`ALTER TABLE transactions DROP COLUMN amount_rule_matched`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`ALTER TABLE transactions DROP COLUMN location_rule_matched`).Error; err != nil {
			return err
		}

		return tx.Exec(`ALTER TABLE transactions DROP COLUMN velocity_rule_matched`).Error
	},
	Down: func(tx *gorm.DB) error {
		if err := tx.Exec(`ALTER TABLE transactions ADD COLUMN amount_rule_matched BOOLEAN NOT NULL DEFAULT FALSE`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`ALTER TABLE transactions ADD COLUMN location_rule_matched BOOLEAN NOT NULL DEFAULT FALSE`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`ALTER TABLE transactions ADD COLUMN velocity_rule_matched BOOLEAN NOT NULL DEFAULT FALSE`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
			UPDATE transactions
			SET amount_rule_matched = COALESCE((
				SELECT matched
				FROM transaction_fraud_rules
				WHERE transaction_id = transactions.id AND rule_name = 'amount'
			), FALSE)
		`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
			UPDATE transactions
			SET location_rule_matched = COALESCE((
				SELECT matched
				FROM transaction_fraud_rules
				WHERE transaction_id = transactions.id AND rule_name = 'location'
			), FALSE)
		`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
			UPDATE transactions
			SET velocity_rule_matched = COALESCE((
				SELECT matched
				FROM transaction_fraud_rules
				WHERE transaction_id = transactions.id AND rule_name = 'velocity'
			), FALSE)
		`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`DROP TABLE IF EXISTS transaction_fraud_rules`).Error; err != nil {
			return err
		}

		return tx.Exec(`DROP TYPE IF EXISTS fraud_rule_name`).Error
	},
}
