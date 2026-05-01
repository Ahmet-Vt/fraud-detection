package migrations

import "fraud-detection/internal/database"

func Register(migrator *database.Migrator) error {
	return migrator.Register(
		CreateUsersTable,
		CreateTransactionsTable,
		CreateTransactionFraudRulesTable,
	)
}
