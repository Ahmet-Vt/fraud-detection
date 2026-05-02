package transaction

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"fraud-detection/internal/domain"
	"fraud-detection/internal/secureerr"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

type createTransactionRequest struct {
	UserID   uuid.UUID `json:"user_id" binding:"required" format:"uuid" example:"4ca5e74c-3a07-4010-9f07-14e90a35868b" extensions:"x-order=0"`
	Amount   int64     `json:"amount" binding:"required,min=1" example:"5000" extensions:"x-order=1"`
	Location string    `json:"location" binding:"required,max=100" example:"Istanbul" extensions:"x-order=2"`
}

type CreateTransactionRequest = createTransactionRequest

type transactionResponse struct {
	ID                uuid.UUID     `json:"id" format:"uuid" example:"1dc90ef0-76a0-4cb2-891e-5193c60aa82f" extensions:"x-order=0"`
	UserID            uuid.UUID     `json:"user_id" format:"uuid" example:"4ca5e74c-3a07-4010-9f07-14e90a35868b" extensions:"x-order=1"`
	Amount            int64         `json:"amount" example:"5000" extensions:"x-order=2"`
	Location          string        `json:"location" example:"Istanbul" extensions:"x-order=3"`
	TransactionStatus domain.Status `json:"transaction_status" enums:"pending,completed,failed" example:"pending" extensions:"x-order=4"`
	FraudStatus       bool          `json:"fraud_status" example:"false" extensions:"x-order=5"`
	CreatedAt         time.Time     `json:"created_at" swaggertype:"string" format:"date-time" example:"2026-05-02T12:00:00Z" extensions:"x-order=6"`
}

type TransactionResponse = transactionResponse

type TimeRangeQuery struct {
	From *time.Time `form:"from" time_format:"2006-01-02T15:04:05Z07:00"`
	To   *time.Time `form:"to" time_format:"2006-01-02T15:04:05Z07:00"`
}

type userTransactionTimeBucketQuery struct {
	TimeRangeQuery
	Bucket    string `form:"bucket"`
	FraudOnly bool   `form:"fraud_only"`
}

func validateBoundTimeRange(fromPtr, toPtr *time.Time) *secureerr.Error {
	if fromPtr == nil && toPtr == nil {
		return nil
	}
	if fromPtr == nil || toPtr == nil {
		return secureerr.BadRequest(
			"INVALID_QUERY",
			"from and to query parameters must both be provided together",
			nil,
			nil,
		)
	}
	if toPtr.Before(*fromPtr) {
		return secureerr.BadRequest(
			"INVALID_QUERY",
			"to must not be before from",
			nil,
			nil,
		)
	}
	return nil
}

func bindTimeRangeQuery(c *gin.Context) (fromPtr, toPtr *time.Time, secErr *secureerr.Error) {
	var q TimeRangeQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		return nil, nil, secureerr.BadRequest(
			"INVALID_QUERY",
			"invalid query parameters",
			err,
			nil,
		)
	}
	if sec := validateBoundTimeRange(q.From, q.To); sec != nil {
		return nil, nil, sec
	}
	return q.From, q.To, nil
}

func bindUserTransactionTimeBucketQuery(c *gin.Context) (fromPtr, toPtr *time.Time, bucket string, fraudOnly bool, secErr *secureerr.Error) {
	var q userTransactionTimeBucketQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		return nil, nil, "", false, secureerr.BadRequest(
			"INVALID_QUERY",
			"invalid query parameters",
			err,
			nil,
		)
	}
	if sec := validateBoundTimeRange(q.From, q.To); sec != nil {
		return nil, nil, "", false, sec
	}
	return q.From, q.To, strings.TrimSpace(q.Bucket), q.FraudOnly, nil
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router gin.IRouter) {
	router.POST("/transactions", h.CreateTransaction)
	router.GET("/transactions/:id", h.GetTransactionByID)
	router.GET("/transactions", h.GetAllTransactions)
	router.GET("/transactions/user/:id", h.GetTransactionsByUserID)
	router.GET("/transactions/user/:id/detail", h.GetUserTransactionsDetail)
	router.GET("/transactions/user/:id/summary", h.GetUserTransactionSummary)
	router.DELETE("/transactions/:id", h.DeleteTransaction)
}

// CreateTransaction godoc
// @Summary Create transaction
// @Description Creates a new transaction for a user.
// @Tags transactions
// @Accept json
// @Produce json
// @Param payload body CreateTransactionRequest true "Create transaction payload"
// @Success 201 {object} TransactionResponse
// @Failure 400 {object} apidoc.ReferencedUserNotExistsError
// @Failure 409 {object} apidoc.TransactionAlreadyExistsError
// @Failure 500 {object} apidoc.InternalServerError
// @Router /transactions [post]
func (h *Handler) CreateTransaction(c *gin.Context) {
	var req createTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, secureerr.BadRequest(
			"INVALID_REQUEST_BODY",
			"invalid request body",
			err,
			nil,
		))
		return
	}

	transaction := &domain.Transaction{
		UserID:   req.UserID,
		Amount:   req.Amount,
		Location: req.Location,
	}

	if err := h.service.CreateTransaction(c.Request.Context(), transaction); err != nil {
		h.writeError(c, h.mapError(err))
		return
	}

	c.JSON(http.StatusCreated, toTransactionResponse(transaction))
}

// GetTransactionByID godoc
// @Summary Get transaction by ID
// @Description Returns a single transaction by UUID.
// @Tags transactions
// @Produce json
// @Param id path string true "Transaction ID"
// @Success 200 {object} TransactionResponse
// @Failure 400 {object} apidoc.InvalidTransactionIDError
// @Failure 404 {object} apidoc.TransactionNotFoundError
// @Failure 500 {object} apidoc.InternalServerError
// @Router /transactions/{id} [get]
func (h *Handler) GetTransactionByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.writeError(c, secureerr.BadRequest(
			"INVALID_TRANSACTION_ID",
			"invalid transaction id",
			err,
			map[string]string{"transaction_id": c.Param("id")},
		))
		return
	}

	transaction, err := h.service.GetTransactionByID(id)
	if err != nil {
		h.writeError(c, h.mapError(err))
		return
	}

	c.JSON(http.StatusOK, toTransactionResponse(transaction))
}

// GetAllTransactions godoc
// @Summary List transactions
// @Description Returns all transactions in deterministic order.
// @Tags transactions
// @Produce json
// @Success 200 {array} TransactionResponse
// @Failure 500 {object} apidoc.InternalServerError
// @Router /transactions [get]
func (h *Handler) GetAllTransactions(c *gin.Context) {
	transactions, err := h.service.GetAllTransactions()
	if err != nil {
		h.writeError(c, h.mapError(err))
		return
	}

	c.JSON(http.StatusOK, toTransactionResponses(transactions))
}

// GetTransactionsByUserID godoc
// @Summary List transactions by user
// @Description Returns transactions for a specific user UUID. Optional from/to (RFC3339) filter on created_at; both must be sent together.
// @Tags transactions
// @Produce json
// @Param id path string true "User ID"
// @Param from query string false "Range start RFC3339 (omit with to for all-time list)"
// @Param to query string false "Range end RFC3339 (omit with from for all-time list)"
// @Success 200 {array} TransactionResponse
// @Failure 400 {object} apidoc.InvalidUserIDError
// @Failure 500 {object} apidoc.InternalServerError
// @Router /transactions/user/{id} [get]
func (h *Handler) GetTransactionsByUserID(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.writeError(c, secureerr.BadRequest(
			"INVALID_USER_ID",
			"invalid user id",
			err,
			map[string]string{"user_id": c.Param("id")},
		))
		return
	}

	fromPtr, toPtr, secErr := bindTimeRangeQuery(c)
	if secErr != nil {
		h.writeError(c, secErr)
		return
	}

	transactions, err := h.service.GetTransactionsByUserID(userID, fromPtr, toPtr, false)
	if err != nil {
		h.writeError(c, h.mapError(err))
		return
	}

	c.JSON(http.StatusOK, toTransactionResponses(transactions))
}

type transactionSummaryResponse struct {
	BucketStart  time.Time `json:"bucket_start"`
	RequestCount int64     `json:"request_count"`
	AnomalyCount int64     `json:"anomaly_count"`
	AmountSum    int64     `json:"amount_sum"`
	AmountAvg    float64   `json:"amount_avg"`
}

type userTransactionsDetailResponse struct {
	Transactions []transactionResponse        `json:"transactions"`
	Summary      []transactionSummaryResponse `json:"summary"`
}

func bucketSummariesToResponses(summary []UserTransactionBucketSummary) []transactionSummaryResponse {
	response := make([]transactionSummaryResponse, 0, len(summary))
	for _, item := range summary {
		response = append(response, transactionSummaryResponse{
			BucketStart:  item.BucketStart,
			RequestCount: item.RequestCount,
			AnomalyCount: item.AnomalyCount,
			AmountSum:    item.AmountSum,
			AmountAvg:    item.AmountAvg,
		})
	}
	return response
}

// GetUserTransactionsDetail godoc
// @Summary List user transactions with summary
// @Description Returns transactions and bucket summary in one response (same filters as separate list/summary calls). Default bucket is day when omitted.
// @Tags transactions
// @Produce json
// @Param id path string true "User ID"
// @Param from query string false "Range start RFC3339 (omit with to for all-time)"
// @Param to query string false "Range end RFC3339 (omit with from for all-time)"
// @Param bucket query string false "Bucket unit: minute|hour|day (default day)"
// @Param fraud_only query bool false "When true, transactions list and summary use only fraud_status=true rows"
// @Success 200 {object} transaction.userTransactionsDetailResponse
// @Failure 400 {object} apidoc.InvalidUserIDError
// @Failure 500 {object} apidoc.InternalServerError
// @Router /transactions/user/{id}/detail [get]
func (h *Handler) GetUserTransactionsDetail(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.writeError(c, secureerr.BadRequest(
			"INVALID_USER_ID",
			"invalid user id",
			err,
			map[string]string{"user_id": c.Param("id")},
		))
		return
	}

	fromPtr, toPtr, bucket, fraudOnly, secErr := bindUserTransactionTimeBucketQuery(c)
	if secErr != nil {
		h.writeError(c, secErr)
		return
	}

	if bucket == "" {
		bucket = "day"
	}

	transactions, err := h.service.GetTransactionsByUserID(userID, fromPtr, toPtr, fraudOnly)
	if err != nil {
		h.writeError(c, h.mapError(err))
		return
	}

	summary, err := h.service.GetUserTransactionSummary(userID, fromPtr, toPtr, bucket, fraudOnly)
	if err != nil {
		h.writeError(c, h.mapError(err))
		return
	}

	c.JSON(http.StatusOK, userTransactionsDetailResponse{
		Transactions: toTransactionResponses(transactions),
		Summary:      bucketSummariesToResponses(summary),
	})
}

// GetUserTransactionSummary godoc
// @Summary Get user transaction summary
// @Description Returns bucket-based transaction aggregates for a user. With fraud_only=true, only fraud_status=true transactions are included in counts and amounts.
// @Tags transactions
// @Produce json
// @Param id path string true "User ID"
// @Param from query string false "Range start in RFC3339 (omit with to for all-time summary)"
// @Param to query string false "Range end in RFC3339 (omit with from for all-time summary)"
// @Param bucket query string false "Bucket unit: minute|hour|day (empty: day for all-time, minute when from/to set)"
// @Param fraud_only query bool false "When true, aggregate only fraud_status=true transactions"
// @Success 200 {array} transaction.transactionSummaryResponse
// @Failure 400 {object} apidoc.InvalidUserIDError
// @Failure 500 {object} apidoc.InternalServerError
// @Router /transactions/user/{id}/summary [get]
func (h *Handler) GetUserTransactionSummary(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.writeError(c, secureerr.BadRequest(
			"INVALID_USER_ID",
			"invalid user id",
			err,
			map[string]string{"user_id": c.Param("id")},
		))
		return
	}

	fromPtr, toPtr, bucket, fraudOnly, secErr := bindUserTransactionTimeBucketQuery(c)
	if secErr != nil {
		h.writeError(c, secErr)
		return
	}

	summary, err := h.service.GetUserTransactionSummary(userID, fromPtr, toPtr, bucket, fraudOnly)
	if err != nil {
		h.writeError(c, h.mapError(err))
		return
	}

	c.JSON(http.StatusOK, bucketSummariesToResponses(summary))
}

// DeleteTransaction godoc
// @Summary Delete transaction
// @Description Deletes a transaction by UUID.
// @Tags transactions
// @Produce json
// @Param id path string true "Transaction ID"
// @Success 204
// @Failure 400 {object} apidoc.InvalidTransactionIDError
// @Failure 404 {object} apidoc.TransactionNotFoundError
// @Failure 500 {object} apidoc.InternalServerError
// @Router /transactions/{id} [delete]
func (h *Handler) DeleteTransaction(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.writeError(c, secureerr.BadRequest(
			"INVALID_TRANSACTION_ID",
			"invalid transaction id",
			err,
			map[string]string{"transaction_id": c.Param("id")},
		))
		return
	}

	if err := h.service.DeleteTransaction(id); err != nil {
		h.writeError(c, h.mapError(err))
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) mapError(err error) error {
	if errors.Is(err, ErrTransactionNotFound) {
		return secureerr.NotFound("TRANSACTION_NOT_FOUND", "transaction not found", err, nil)
	}
	if errors.Is(err, ErrTransactionAlreadyExists) {
		return secureerr.Conflict("TRANSACTION_ALREADY_EXISTS", "transaction already exists", err, nil)
	}
	if errors.Is(err, ErrUserNotExists) {
		return secureerr.BadRequest("USER_NOT_EXISTS", "referenced user does not exist", err, nil)
	}
	if errors.Is(err, ErrInvalidTimeRange) {
		return secureerr.BadRequest("INVALID_QUERY", "invalid time range", err, nil)
	}
	if errors.Is(err, ErrInvalidBucket) {
		return secureerr.BadRequest("INVALID_QUERY", "invalid bucket", err, nil)
	}
	return secureerr.FromError(err)
}

func (h *Handler) writeError(c *gin.Context, err error) {
	appErr := secureerr.FromError(err)
	if appErr == nil {
		return
	}

	log.Printf("transaction handler error: %s", appErr.LogString())

	c.JSON(appErr.StatusCode(), gin.H{
		"code":  appErr.Code,
		"error": appErr.Error(),
	})
}

func toTransactionResponse(transaction *domain.Transaction) transactionResponse {
	return transactionResponse{
		ID:                transaction.ID,
		UserID:            transaction.UserID,
		Amount:            transaction.Amount,
		Location:          transaction.Location,
		TransactionStatus: transaction.TransactionStatus,
		FraudStatus:       transaction.FraudStatus,
		CreatedAt:         transaction.CreatedAt,
	}
}

func toTransactionResponses(transactions []*domain.Transaction) []transactionResponse {
	responses := make([]transactionResponse, 0, len(transactions))
	for _, transaction := range transactions {
		responses = append(responses, toTransactionResponse(transaction))
	}

	return responses
}
