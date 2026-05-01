package fraud

import "errors"

var (
	ErrInvalidTransactionID     = errors.New("invalid transaction id")
	ErrInvalidUserID            = errors.New("invalid user id")
	ErrTransactionLookupFailed  = errors.New("transaction lookup failed")
	ErrUserTransactionsFailed   = errors.New("user transactions lookup failed")
	ErrAmountRuleEvaluation     = errors.New("amount rule evaluation failed")
	ErrLocationRuleEvaluation   = errors.New("location rule evaluation failed")
	ErrVelocityRuleEvaluation   = errors.New("velocity rule evaluation failed")
	ErrFraudRuleEvaluation      = errors.New("fraud rule evaluation failed")
	ErrUpdateFraudRulesFailed   = errors.New("update fraud rules failed")
	ErrLastLocationCacheFailed  = errors.New("last location cache failed")
	ErrLastCreatedAtCacheFailed = errors.New("last transaction time cache failed")
	ErrRequestCountCacheFailed  = errors.New("request count cache failed")
	ErrAmountAverageCacheFailed = errors.New("amount average cache failed")
)
