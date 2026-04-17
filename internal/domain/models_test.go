package domain

import (
	"reflect"
	"testing"
)

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

	transaction.RefreshFraudStatus()

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

	transaction.RefreshFraudStatus()

	if transaction.FraudStatus {
		t.Fatal("expected fraud status to be false when fewer than two rules are matched")
	}
}
