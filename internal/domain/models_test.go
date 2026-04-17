package domain

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/Ahmet-Vt/fraud-detection/internal/database"
	"github.com/Ahmet-Vt/fraud-detection/migrations"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func logFraudRules(t *testing.T, rules []TransactionFraudRule) {
	t.Helper()

	for _, rule := range rules {
		t.Logf("transaction_fraud_rules row -> rule_name=%s matched=%t", rule.RuleName, rule.Matched)
	}
}

func logFraudStateAsJSON(t *testing.T, transaction *Transaction) {
	t.Helper()

	payload := map[string]any{
		"transaction": map[string]any{
			"fraud_status":       transaction.FraudStatus,
			"transaction_status": transaction.TransactionStatus,
		},
		"transaction_fraud_rules": transaction.FraudRules,
	}

	prettyJSON, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatalf("marshal fraud state json: %v", err)
	}

	t.Logf("\n%s", prettyJSON)
}

func logStepAsJSON(t *testing.T, step string, payload any) {
	t.Helper()

	prettyJSON, err := json.MarshalIndent(map[string]any{
		"step": step,
		"data": payload,
	}, "", "  ")
	if err != nil {
		t.Fatalf("marshal step json: %v", err)
	}

	t.Logf("\n%s", prettyJSON)
}

func integrationTestDSN(t *testing.T) string {
	t.Helper()

	env := map[string]string{
		"POSTGRES_HOST":     os.Getenv("POSTGRES_HOST"),
		"POSTGRES_PORT":     os.Getenv("POSTGRES_PORT"),
		"POSTGRES_USER":     os.Getenv("POSTGRES_USER"),
		"POSTGRES_PASSWORD": os.Getenv("POSTGRES_PASSWORD"),
		"POSTGRES_DB":       os.Getenv("POSTGRES_DB"),
		"POSTGRES_SSLMODE":  os.Getenv("POSTGRES_SSLMODE"),
	}

	missingKey := ""
	for key, value := range env {
		if value == "" {
			missingKey = key
			break
		}
	}

	if missingKey != "" {
		envPath := filepath.Join("..", "..", ".env")
		fileEnv, err := godotenv.Read(envPath)
		if err != nil {
			t.Fatalf("load integration env from %s: %v", envPath, err)
		}

		for key := range env {
			if env[key] == "" {
				env[key] = fileEnv[key]
			}
		}
	}

	for key, value := range env {
		if value == "" {
			t.Fatalf("missing required integration env var: %s", key)
		}
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		env["POSTGRES_HOST"],
		env["POSTGRES_PORT"],
		env["POSTGRES_USER"],
		env["POSTGRES_PASSWORD"],
		env["POSTGRES_DB"],
		env["POSTGRES_SSLMODE"],
	)
}

func integrationTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := database.NewPostgres(integrationTestDSN(t))
	if err != nil {
		t.Fatalf("connect integration database: %v", err)
	}

	migrator := database.NewMigrator(db)
	if err := migrations.Register(migrator); err != nil {
		t.Fatalf("register migrations: %v", err)
	}
	if err := migrator.Run(); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	return db
}

func TestTransactionFraudRuleRuleNameUsesEnumMapping(t *testing.T) {
	field, ok := reflect.TypeOf(TransactionFraudRule{}).FieldByName("RuleName")
	if !ok {
		t.Fatal("expected RuleName field on TransactionFraudRule")
	}

	gormTag := field.Tag.Get("gorm")
	t.Logf("RuleName gorm tag: %s", gormTag)

	expectedTag := "type:fraud_rule_name;not null"
	if gormTag != expectedTag {
		t.Fatalf("expected RuleName gorm tag %q, got %q", expectedTag, gormTag)
	}
}

func TestFraudRuleNamesEnumValues(t *testing.T) {
	ruleNames := []FraudRuleName{
		FraudRuleAmount,
		FraudRuleLocation,
		FraudRuleVelocity,
	}
	expectedValues := []FraudRuleName{"amount", "location", "velocity"}

	seen := make(map[FraudRuleName]struct{}, len(ruleNames))
	for i, ruleName := range ruleNames {
		t.Logf("fraud rule enum value: %s", ruleName)
		if ruleName != expectedValues[i] {
			t.Fatalf("expected fraud rule enum value %q at index %d, got %q", expectedValues[i], i, ruleName)
		}
		if _, exists := seen[ruleName]; exists {
			t.Fatalf("duplicate fraud rule enum value detected: %s", ruleName)
		}
		seen[ruleName] = struct{}{}
	}
}

func TestTransactionRefreshFraudStatus(t *testing.T) {
	transaction := &Transaction{
		FraudRules: []TransactionFraudRule{
			{RuleName: FraudRuleAmount, Matched: true},
			{RuleName: FraudRuleLocation, Matched: true},
			{RuleName: FraudRuleVelocity, Matched: false},
		},
	}

	logFraudRules(t, transaction.FraudRules)

	transaction.RefreshFraudStatus()
	t.Logf("transactions row -> fraud_status=%t", transaction.FraudStatus)

	if !transaction.FraudStatus {
		t.Fatal("expected fraud status to be true when at least two rules are matched")
	}
}

func TestTransactionRefreshFraudStatusBelowThreshold(t *testing.T) {
	transaction := &Transaction{
		FraudRules: []TransactionFraudRule{
			{RuleName: FraudRuleAmount, Matched: true},
			{RuleName: FraudRuleLocation, Matched: false},
		},
	}

	logFraudRules(t, transaction.FraudRules)

	transaction.RefreshFraudStatus()
	t.Logf("transactions row -> fraud_status=%t", transaction.FraudStatus)

	if transaction.FraudStatus {
		t.Fatal("expected fraud status to be false when fewer than two rules are matched")
	}
}

func TestTransactionFraudStateAsJSON(t *testing.T) {
	transaction := &Transaction{
		TransactionStatus: TransactionStatusPending,
		FraudRules: []TransactionFraudRule{
			{RuleName: FraudRuleAmount, Matched: false},
			{RuleName: FraudRuleLocation, Matched: true},
			{RuleName: FraudRuleVelocity, Matched: true},
		},
	}

	transaction.RefreshFraudStatus()
	logFraudStateAsJSON(t, transaction)

	if !transaction.FraudStatus {
		t.Fatal("expected fraud status to be true when two rules are matched")
	}
}

func TestTransactionFraudFlowIntegration(t *testing.T) {
	db := integrationTestDB(t)

	user := &User{
		Name:       "Integration",
		Surname:    "TestUser",
		Email:      ptr(fmt.Sprintf("integration-%d@example.com", time.Now().UnixNano())),
		UserStatus: UserStatusActive,
	}

	var transactionID uuid.UUID
	var userID uuid.UUID

	defer func() {
		if transactionID != uuid.Nil {
			result := db.Where("transaction_id = ?", transactionID).Delete(&TransactionFraudRule{})
			logStepAsJSON(t, "cleanup_fraud_rules", map[string]any{
				"transaction_id": transactionID,
				"rows_affected":  result.RowsAffected,
				"error":          errorString(result.Error),
			})

			result = db.Delete(&Transaction{}, transactionID)
			logStepAsJSON(t, "cleanup_transaction", map[string]any{
				"transaction_id": transactionID,
				"rows_affected":  result.RowsAffected,
				"error":          errorString(result.Error),
			})
		}

		if userID != uuid.Nil {
			result := db.Delete(&User{}, userID)
			logStepAsJSON(t, "cleanup_user", map[string]any{
				"user_id":       userID,
				"rows_affected": result.RowsAffected,
				"error":         errorString(result.Error),
			})
		}
	}()

	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create integration user: %v", err)
	}
	userID = user.ID
	logStepAsJSON(t, "user_created", user)

	transaction := &Transaction{
		UserID:            user.ID,
		Amount:            2450,
		City:              "istanbul",
		TransactionStatus: TransactionStatusPending,
		FraudStatus:       false,
	}
	if err := db.Create(transaction).Error; err != nil {
		t.Fatalf("create integration transaction: %v", err)
	}
	transactionID = transaction.ID
	logStepAsJSON(t, "transaction_created", transaction)

	rules := []TransactionFraudRule{
		{TransactionID: transaction.ID, RuleName: FraudRuleAmount, Matched: false},
		{TransactionID: transaction.ID, RuleName: FraudRuleLocation, Matched: true},
		{TransactionID: transaction.ID, RuleName: FraudRuleVelocity, Matched: true},
	}
	if err := db.Create(&rules).Error; err != nil {
		t.Fatalf("create integration fraud rules: %v", err)
	}

	var persistedTransaction Transaction
	if err := db.Preload("FraudRules").First(&persistedTransaction, transaction.ID).Error; err != nil {
		t.Fatalf("load transaction with fraud rules: %v", err)
	}

	persistedTransaction.RefreshFraudStatus()
	if err := db.Model(&Transaction{}).
		Where("id = ?", persistedTransaction.ID).
		Update("fraud_status", persistedTransaction.FraudStatus).Error; err != nil {
		t.Fatalf("update transaction fraud status: %v", err)
	}

	if err := db.Preload("FraudRules").First(&persistedTransaction, transaction.ID).Error; err != nil {
		t.Fatalf("reload transaction with fraud rules: %v", err)
	}

	logStepAsJSON(t, "fraud_rules_set", map[string]any{
		"transaction":             persistedTransaction,
		"transaction_fraud_rules": persistedTransaction.FraudRules,
	})

	if !persistedTransaction.FraudStatus {
		t.Fatal("expected fraud status to be true when two fraud rules are matched")
	}
}

func ptr(value string) *string {
	return &value
}

func errorString(err error) any {
	if err == nil {
		return nil
	}

	return err.Error()
}
