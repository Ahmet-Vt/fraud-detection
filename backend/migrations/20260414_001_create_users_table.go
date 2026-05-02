package migrations 

import (
	"fraud-detection/internal/database"

	"gorm.io/gorm"
)

var CreateUsersTable = database.MigrationEntry{
	Version: "20260414_001",
	Name: "create_users_table",
	Up: func(tx *gorm.DB) error {
		if err := tx.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
			DO $$
			BEGIN
				IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_status') THEN
					CREATE TYPE user_status AS ENUM ('active', 'suspended', 'deleted');
				END IF;
			END
			$$;
		`).Error; err != nil {
			return err
		}

		return tx.Exec(`
			CREATE TABLE users (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				name VARCHAR(100) NOT NULL,
				surname VARCHAR(100) NOT NULL,
				email VARCHAR(255) UNIQUE,
				status user_status NOT NULL DEFAULT 'active',
				risk_rate DOUBLE PRECISION NOT NULL DEFAULT 0,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			)
		`).Error
	},
	Down: func(tx *gorm.DB) error {
		if err := tx.Exec(`DROP TABLE IF EXISTS users`).Error; err != nil {
			return err
		}

		return tx.Exec(`DROP TYPE IF EXISTS user_status`).Error
	},
}