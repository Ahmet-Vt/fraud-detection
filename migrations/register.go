package migrations

import "github.com/Ahmet-Vt/fraud-detection/internal/database"

func Register(migrator *database.Migrator) error {
	return migrator.Register(
		CreateUsersTable,
		CreateTransactionsTable,
	)
}
