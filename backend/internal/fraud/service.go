package fraud

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"fraud-detection/internal/domain"
	"fraud-detection/internal/transaction/distance"

	"github.com/google/uuid"
)

const (
	defaultRuleTimeout         = 2 * time.Second
	requestCountFraudThreshold = 5
)

type transactionReaderWriter interface {
	GetTransactionByID(id uuid.UUID) (*domain.Transaction, error)
	GetPreviousTransactionByUserID(userID uuid.UUID, before time.Time) (*domain.Transaction, bool, error)
	GetTripleAmountAverageLast24HoursByUserID(userID uuid.UUID, excludeTxID *uuid.UUID) (int64, error)
	UpdateFraudRules(transactionID uuid.UUID, rules []domain.TransactionFraudRule) error
	UpdateTransactionStatus(transactionID uuid.UUID, status domain.Status) error
	PublishTransactionEvaluated(ctx context.Context, transaction *domain.Transaction, evaluatedAt time.Time) error
}

type transactionStateCache interface {
	GetLastLocationByUserID(ctx context.Context, userID uuid.UUID) (string, bool, error)
	SetLastLocationByUserID(ctx context.Context, userID uuid.UUID, location string) error
	GetLastTransactionCreatedAtByUserID(ctx context.Context, userID uuid.UUID) (time.Time, bool, error)
	SetLastTransactionCreatedAtByUserID(ctx context.Context, userID uuid.UUID, createdAt time.Time) error
	GetRequestCountByUserID(ctx context.Context, userID uuid.UUID, referenceTime time.Time) (int, error)
	GetLast24HoursAverageSpendByUserID(ctx context.Context, userID uuid.UUID) (int64, bool, error)
	SetLast24HoursAverageSpendByUserID(ctx context.Context, userID uuid.UUID, average int64) error
}

type Service struct {
	transactionService transactionReaderWriter
	cache              transactionStateCache
}

type cachedRuleInputs struct {
	lastLocation     string
	lastCreatedAt    time.Time
	hasLastLocation  bool
	hasLastCreatedAt bool
	requestCount     int
	hasRequestCount  bool
}

func NewService(transactionService transactionReaderWriter, cache transactionStateCache) *Service {
	return &Service{
		transactionService: transactionService,
		cache:              cache,
	}
}

func (s *Service) EvaluateTransaction(ctx context.Context, transactionID uuid.UUID, userID uuid.UUID) ([]domain.TransactionFraudRule, error) {
	currentTransaction, err := s.transactionService.GetTransactionByID(transactionID)
	if err != nil {
		return nil, fmt.Errorf("evaluate transaction: %w", errors.Join(ErrTransactionLookupFailed, err))
	}

	previousTransaction, hasPrevious, err := s.transactionService.GetPreviousTransactionByUserID(userID, currentTransaction.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("evaluate transaction: %w", errors.Join(ErrUserTransactionsFailed, err))
	}

	cacheInputs, err := s.loadCachedRuleInputs(ctx, currentTransaction)
	if err != nil {
		return nil, fmt.Errorf("evaluate transaction: %w", errors.Join(ErrFraudRuleEvaluation, err))
	}

	rules := []domain.TransactionFraudRule{
		{RuleName: domain.FraudRuleAmount},
		{RuleName: domain.FraudRuleLocation},
		{RuleName: domain.FraudRuleVelocity},
	}

	groupCtx, cancelGroup := context.WithCancel(ctx)
	defer cancelGroup()

	errCh := make(chan error, len(rules))
	var wg sync.WaitGroup
	wg.Add(len(rules))

	go s.runAmountRule(groupCtx, &wg, errCh, &rules[0], currentTransaction)
	go s.runLocationRule(groupCtx, &wg, errCh, &rules[1], currentTransaction, previousTransaction, hasPrevious, cacheInputs)
	go s.runVelocityRule(groupCtx, &wg, errCh, &rules[2], cacheInputs)

	go func() {
		wg.Wait()
		close(errCh)
	}()

	var firstErr error
	for ruleErr := range errCh {
		if ruleErr != nil && firstErr == nil {
			firstErr = ruleErr
			cancelGroup()
		}
	}
	if firstErr != nil {
		return nil, fmt.Errorf("evaluate transaction: %w", errors.Join(ErrFraudRuleEvaluation, firstErr))
	}

	if err := s.transactionService.UpdateFraudRules(transactionID, rules); err != nil {
		return nil, fmt.Errorf("evaluate transaction: %w", errors.Join(ErrUpdateFraudRulesFailed, err))
	}

	currentTransaction.FraudRules = rules
	currentTransaction.RefreshFraudStatus()

	if err := s.transactionService.UpdateTransactionStatus(transactionID, domain.TransactionStatusCompleted); err != nil {
		return nil, fmt.Errorf("evaluate transaction: %w", errors.Join(ErrUpdateFraudRulesFailed, err))
	}
	currentTransaction.TransactionStatus = domain.TransactionStatusCompleted

	if err := s.transactionService.PublishTransactionEvaluated(ctx, currentTransaction, time.Now()); err != nil {
		return nil, fmt.Errorf("evaluate transaction: %w", errors.Join(ErrFraudRuleEvaluation, err))
	}

	return rules, nil
}

func (s *Service) runAmountRule(
	ctx context.Context,
	wg *sync.WaitGroup,
	errCh chan<- error,
	rule *domain.TransactionFraudRule,
	currentTransaction *domain.Transaction,
) {
	defer wg.Done()

	jobCtx, cancel := context.WithTimeout(ctx, defaultRuleTimeout)
	defer cancel()

	matched, err := s.evaluateAmountRule(jobCtx, currentTransaction)
	if err != nil {
		errCh <- fmt.Errorf("amount rule failed: %w", errors.Join(ErrAmountRuleEvaluation, err))
		return
	}
	rule.Matched = matched
}

func (s *Service) runLocationRule(
	ctx context.Context,
	wg *sync.WaitGroup,
	errCh chan<- error,
	rule *domain.TransactionFraudRule,
	currentTransaction *domain.Transaction,
	previousTransaction *domain.Transaction,
	hasPrevious bool,
	cacheInputs cachedRuleInputs,
) {
	defer wg.Done()

	jobCtx, cancel := context.WithTimeout(ctx, defaultRuleTimeout)
	defer cancel()

	matched, err := s.evaluateLocationRule(jobCtx, currentTransaction, previousTransaction, hasPrevious, cacheInputs)
	if err != nil {
		errCh <- fmt.Errorf("location rule failed: %w", errors.Join(ErrLocationRuleEvaluation, err))
		return
	}
	rule.Matched = matched
}

func (s *Service) runVelocityRule(
	ctx context.Context,
	wg *sync.WaitGroup,
	errCh chan<- error,
	rule *domain.TransactionFraudRule,
	cacheInputs cachedRuleInputs,
) {
	defer wg.Done()

	jobCtx, cancel := context.WithTimeout(ctx, defaultRuleTimeout)
	defer cancel()

	matched, err := s.evaluateVelocityRule(jobCtx, cacheInputs)
	if err != nil {
		errCh <- fmt.Errorf("velocity rule failed: %w", errors.Join(ErrVelocityRuleEvaluation, err))
		return
	}
	rule.Matched = matched
}

func (s *Service) evaluateAmountRule(ctx context.Context, currentTransaction *domain.Transaction) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	var (
		tripleAverage int64
		err           error
		cacheHit      bool
	)

	if s.cache != nil {
		cachedAverage, found, cacheErr := s.cache.GetLast24HoursAverageSpendByUserID(ctx, currentTransaction.UserID)
		if cacheErr != nil {
			return false, errors.Join(ErrAmountAverageCacheFailed, cacheErr)
		}
		if found {
			tripleAverage = cachedAverage
			cacheHit = true
		}
	}

	if !cacheHit {
		tripleAverage, err = s.transactionService.GetTripleAmountAverageLast24HoursByUserID(
			currentTransaction.UserID,
			&currentTransaction.ID,
		)
		if err != nil {
			return false, err
		}
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}

	matched := tripleAverage != 0 && currentTransaction.Amount > tripleAverage

	if s.cache != nil && !matched {
		latestAverage, err := s.transactionService.GetTripleAmountAverageLast24HoursByUserID(currentTransaction.UserID, nil)
		if err != nil {
			return false, err
		}
		if err := s.cache.SetLast24HoursAverageSpendByUserID(ctx, currentTransaction.UserID, latestAverage); err != nil {
			return false, errors.Join(ErrAmountAverageCacheFailed, err)
		}
	}

	return matched, nil
}

func (s *Service) evaluateLocationRule(
	ctx context.Context,
	currentTransaction *domain.Transaction,
	previousTransaction *domain.Transaction,
	hasPrevious bool,
	cacheInputs cachedRuleInputs,
) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	matched := computeLocationImpossibleTravelMatch(currentTransaction, previousTransaction, hasPrevious, cacheInputs)

	if s.cache != nil && !matched {
		if err := s.cache.SetLastLocationByUserID(ctx, currentTransaction.UserID, currentTransaction.Location); err != nil {
			return false, errors.Join(ErrLastLocationCacheFailed, err)
		}
		if err := s.cache.SetLastTransactionCreatedAtByUserID(ctx, currentTransaction.UserID, currentTransaction.CreatedAt); err != nil {
			return false, errors.Join(ErrLastCreatedAtCacheFailed, err)
		}
	}

	return matched, nil
}

func computeLocationImpossibleTravelMatch(
	currentTransaction *domain.Transaction,
	previousTransaction *domain.Transaction,
	hasPrevious bool,
	cacheInputs cachedRuleInputs,
) bool {
	currentCity := normalizeCity(currentTransaction.Location)
	if currentCity == "" {
		return false
	}
	previousCity := ""
	previousCreatedAt := time.Time{}
	hasCachePrevious := false

	if cacheInputs.hasLastLocation {
		previousCity = normalizeCity(cacheInputs.lastLocation)
	}
	if cacheInputs.hasLastCreatedAt {
		previousCreatedAt = cacheInputs.lastCreatedAt
	}

	hasCachePrevious = previousCity != "" && !previousCreatedAt.IsZero()
	if hasCachePrevious && !previousCreatedAt.Before(currentTransaction.CreatedAt) {
		hasCachePrevious = false
	}

	if !hasCachePrevious {
		if !hasPrevious || previousTransaction == nil {
			return false
		}
		previousCity = normalizeCity(previousTransaction.Location)
		previousCreatedAt = previousTransaction.CreatedAt
	}
	if previousCity == "" || previousCreatedAt.IsZero() {
		return false
	}

	if currentCity == previousCity {
		return false
	}

	currentCoordinates, ok := distance.GetCityCoordinates(currentCity)
	if !ok {
		return false
	}
	previousCoordinates, ok := distance.GetCityCoordinates(previousCity)
	if !ok {
		return false
	}

	kmDistance := distance.HaversineDistance(
		previousCoordinates.Lat,
		previousCoordinates.Lon,
		currentCoordinates.Lat,
		currentCoordinates.Lon,
	)
	elapsed := currentTransaction.CreatedAt.Sub(previousCreatedAt)
	if elapsed < 0 {
		return false
	}

	reachable := distance.IsReachableWithPlane(kmDistance, elapsed)

	return !reachable
}

func (s *Service) evaluateVelocityRule(
	ctx context.Context,
	cacheInputs cachedRuleInputs,
) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if !cacheInputs.hasRequestCount {
		return false, nil
	}

	return cacheInputs.requestCount > requestCountFraudThreshold, nil
}

func (s *Service) loadCachedRuleInputs(ctx context.Context, currentTransaction *domain.Transaction) (cachedRuleInputs, error) {
	result := cachedRuleInputs{}
	if s.cache == nil {
		return result, nil
	}

	lastLocation, found, err := s.cache.GetLastLocationByUserID(ctx, currentTransaction.UserID)
	if err != nil {
		return result, errors.Join(ErrLastLocationCacheFailed, err)
	}
	if found {
		result.lastLocation = lastLocation
		result.hasLastLocation = true
	}

	lastCreatedAt, found, err := s.cache.GetLastTransactionCreatedAtByUserID(ctx, currentTransaction.UserID)
	if err != nil {
		return result, errors.Join(ErrLastCreatedAtCacheFailed, err)
	}
	if found {
		result.lastCreatedAt = lastCreatedAt
		result.hasLastCreatedAt = true
	}

	requestCount, err := s.cache.GetRequestCountByUserID(
		ctx,
		currentTransaction.UserID,
		currentTransaction.CreatedAt,
	)
	if err != nil {
		return result, errors.Join(ErrRequestCountCacheFailed, err)
	}
	result.requestCount = requestCount
	result.hasRequestCount = true

	return result, nil
}

func normalizeCity(city string) string {
	return strings.ToLower(strings.TrimSpace(city))
}
