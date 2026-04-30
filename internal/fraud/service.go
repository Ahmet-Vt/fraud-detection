package fraud

import (
	"context"
	"errors"
	"fmt"
	"sort"
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
	GetTransactionsByUserID(userID uuid.UUID) ([]*domain.Transaction, error)
	GetTripleAmountAverageLast24HoursByUserID(userID uuid.UUID) (int64, error)
	UpdateFraudRules(transactionID uuid.UUID, rules []domain.TransactionFraudRule) error
}

type transactionStateCache interface {
	GetLastLocationByUserID(ctx context.Context, userID uuid.UUID) (string, bool, error)
	GetLastTransactionCreatedAtByUserID(ctx context.Context, userID uuid.UUID) (time.Time, bool, error)
	GetRequestCountByUserID(ctx context.Context, userID uuid.UUID, referenceTime time.Time) (int, error)
	GetLast24HoursAverageSpendByUserID(ctx context.Context, userID uuid.UUID) (int64, bool, error)
	SetLast24HoursAverageSpendByUserID(ctx context.Context, userID uuid.UUID, average int64) error
	IsVelocityHotByUserID(ctx context.Context, userID uuid.UUID) (bool, error)
	MarkVelocityHotByUserID(ctx context.Context, userID uuid.UUID) error
}

type Service struct {
	transactionService transactionReaderWriter
	cache              transactionStateCache
}

type cachedRuleInputs struct {
	lastLocation             string
	lastCreatedAt            time.Time
	hasLastLocation          bool
	hasLastCreatedAt         bool
	requestCount             int
	hasRequestCount          bool
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

	transactions, err := s.transactionService.GetTransactionsByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("evaluate transaction: %w", errors.Join(ErrUserTransactionsFailed, err))
	}

	previousTransaction, hasPrevious := findPreviousTransaction(transactions, transactionID)
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
		tripleAverage, err = s.transactionService.GetTripleAmountAverageLast24HoursByUserID(currentTransaction.UserID)
		if err != nil {
			return false, err
		}
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}

	matched := tripleAverage != 0 && currentTransaction.Amount > tripleAverage

	if s.cache != nil {
		latestAverage, err := s.transactionService.GetTripleAmountAverageLast24HoursByUserID(currentTransaction.UserID)
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
	currentCity := normalizeCity(currentTransaction.Location)
	if currentCity == "" {
		return false, nil
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
			return false, nil
		}
		previousCity = normalizeCity(previousTransaction.Location)
		previousCreatedAt = previousTransaction.CreatedAt
	}
	if previousCity == "" || previousCreatedAt.IsZero() {
		return false, nil
	}

	if currentCity == previousCity {
		return false, nil
	}

	currentCoordinates, ok := distance.GetCityCoordinates(currentCity)
	if !ok {
		return false, nil
	}
	previousCoordinates, ok := distance.GetCityCoordinates(previousCity)
	if !ok {
		return false, nil
	}

	kmDistance := distance.HaversineDistance(
		previousCoordinates.Lat,
		previousCoordinates.Lon,
		currentCoordinates.Lat,
		currentCoordinates.Lon,
	)
	reachable := distance.IsReachableWithPlane(kmDistance, previousCreatedAt)

	return !reachable, nil
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
	if err := ctx.Err(); err != nil {
		return false, err
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

	hot, err := s.cache.IsVelocityHotByUserID(ctx, currentTransaction.UserID)
	if err != nil {
		return result, errors.Join(ErrRequestCountCacheFailed, err)
	}
	if hot {
		result.requestCount = requestCountFraudThreshold + 1
		result.hasRequestCount = true
		return result, nil
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
	if requestCount > requestCountFraudThreshold {
		if err := s.cache.MarkVelocityHotByUserID(ctx, currentTransaction.UserID); err != nil {
			return result, errors.Join(ErrRequestCountCacheFailed, err)
		}
	}

	return result, nil
}

func findPreviousTransaction(transactions []*domain.Transaction, currentTransactionID uuid.UUID) (*domain.Transaction, bool) {
	if len(transactions) == 0 {
		return nil, false
	}

	sorted := make([]*domain.Transaction, 0, len(transactions))
	for _, tx := range transactions {
		if tx == nil {
			continue
		}
		sorted = append(sorted, tx)
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CreatedAt.After(sorted[j].CreatedAt)
	})

	for _, tx := range sorted {
		if tx.ID == currentTransactionID {
			continue
		}
		return tx, true
	}

	return nil, false
}

func normalizeCity(city string) string {
	return strings.ToLower(strings.TrimSpace(city))
}
