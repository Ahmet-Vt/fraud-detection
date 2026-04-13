package migrations 

import (
	"github.com/Ahmet-Vt/fraud-detection/internal/database"

	"gorm.io/gorm"
)

var CreateUsersTable = database.MigrationEntry{
	Version: "20260413_001",
	Name: "create_users_table",
	Up: func(tx *gorm.DB) error {
		if err := tx.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`CREATE TYPE user_status AS ENUM ('active', 'suspended', 'deleted')`).Error; err != nil {
			return err
		}

		return tx.Exec(`
			CREATE TABLE users (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				name VARCHAR(100) NOT NULL,
				surname VARCHAR(100) NOT NULL,
				email VARCHAR(255) UNIQUE NOT NULL,
				status user_status NOT NULL DEFAULT 'active',
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