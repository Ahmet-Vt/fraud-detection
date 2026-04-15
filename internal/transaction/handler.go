package transaction

import (
	"errors"
	"log"
	"net/http"

	"github.com/Ahmet-Vt/fraud-detection/internal/domain"
	"github.com/Ahmet-Vt/fraud-detection/internal/secureerr"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

type createTransactionRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required" format:"uuid" example:"4ca5e74c-3a07-4010-9f07-14e90a35868b" extensions:"x-order=0"`
	Amount int64     `json:"amount" binding:"required,min=1" example:"5000" extensions:"x-order=1"`
	City   string    `json:"city" binding:"required,max=100" example:"Istanbul" extensions:"x-order=2"`
}

type CreateTransactionRequest = createTransactionRequest

type transactionResponse struct {
	ID     uuid.UUID `json:"id" format:"uuid" example:"1dc90ef0-76a0-4cb2-891e-5193c60aa82f" extensions:"x-order=0"`
	UserID uuid.UUID `json:"user_id" format:"uuid" example:"4ca5e74c-3a07-4010-9f07-14e90a35868b" extensions:"x-order=1"`
	Amount int64     `json:"amount" example:"5000" extensions:"x-order=2"`
	City   string    `json:"city" example:"Istanbul" extensions:"x-order=3"`
}

type TransactionResponse = transactionResponse

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router gin.IRouter) {
	router.POST("/transactions", h.CreateTransaction)
	router.GET("/transactions/:id", h.GetTransactionByID)
	router.GET("/transactions", h.GetAllTransactions)
	router.GET("/transactions/user/:id", h.GetTransactionsByUserID)
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
		UserID: req.UserID,
		Amount: req.Amount,
		City:   req.City,
	}

	if err := h.service.CreateTransaction(transaction); err != nil {
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
// @Description Returns transactions for a specific user UUID.
// @Tags transactions
// @Produce json
// @Param id path string true "User ID"
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

	transactions, err := h.service.GetTransactionsByUserID(userID)
	if err != nil {
		h.writeError(c, h.mapError(err))
		return
	}

	c.JSON(http.StatusOK, toTransactionResponses(transactions))
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

func (h *Handler) mapError(err error) *secureerr.Error {
	if errors.Is(err, ErrTransactionNotFound) {
		return secureerr.NotFound("TRANSACTION_NOT_FOUND", "transaction not found", err, nil)
	}
	if errors.Is(err, ErrTransactionAlreadyExists) {
		return secureerr.Conflict("TRANSACTION_ALREADY_EXISTS", "transaction already exists", err, nil)
	}
	if errors.Is(err, ErrUserNotExists) {
		return secureerr.BadRequest("USER_NOT_EXISTS", "referenced user does not exist", err, nil)
	}
	return secureerr.Internal("INTERNAL_ERROR", "an unexpected error occurred", err, nil)
}

func (h *Handler) writeError(c *gin.Context, appErr *secureerr.Error) {
	log.Printf("transaction handler error: %s", appErr.LogString())

	c.JSON(appErr.StatusCode(), gin.H{
		"code":  appErr.Code,
		"error": appErr.Error(),
	})
}

func toTransactionResponse(transaction *domain.Transaction) transactionResponse {
	return transactionResponse{
		ID:     transaction.ID,
		UserID: transaction.UserID,
		Amount: transaction.Amount,
		City:   transaction.City,
	}
}

func toTransactionResponses(transactions []*domain.Transaction) []transactionResponse {
	responses := make([]transactionResponse, 0, len(transactions))
	for _, transaction := range transactions {
		responses = append(responses, toTransactionResponse(transaction))
	}

	return responses
}
