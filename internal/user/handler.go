package user

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Ahmet-Vt/fraud-detection/internal/domain"
	"github.com/Ahmet-Vt/fraud-detection/internal/secureerr"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

type createUserRequest struct {
	Name    string  `json:"name" binding:"required,max=100"`
	Surname string  `json:"surname" binding:"required,max=100"`
	Email   *string `json:"email" binding:"omitempty,email"`
	Status  string  `json:"status" binding:"required,oneof=active suspended deleted"`
}

type userResponse struct {
	ID        uuid.UUID         `json:"id"`
	Name      string            `json:"name"`
	Surname   string            `json:"surname"`
	Email     *string           `json:"email,omitempty"`
	Status    domain.UserStatus `json:"status"`
	CreatedAt time.Time         `json:"created_at"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router gin.IRouter) {
	router.POST("/users", h.CreateUser)
	router.GET("/users/:id", h.GetUserByID)
	router.GET("/users", h.GetAllUsers)
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, secureerr.BadRequest(
			"INVALID_REQUEST_BODY",
			"invalid request body",
			err,
			nil,
		))
		return
	}

	user := &domain.User{
		Name:       req.Name,
		Surname:    req.Surname,
		Email:      req.Email,
		UserStatus: domain.UserStatus(req.Status),
	}

	if err := h.service.CreateUser(user); err != nil {
		h.writeError(c, h.mapError(err))
		return
	}

	c.JSON(http.StatusCreated, toUserResponse(user))
}

func (h *Handler) GetUserByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.writeError(c, secureerr.BadRequest(
			"INVALID_USER_ID",
			"invalid user id",
			err,
			map[string]string{"user_id": c.Param("id")},
		))
		return
	}

	user, err := h.service.GetUserByID(id)
	if err != nil {
		h.writeError(c, h.mapError(err))
		return
	}

	c.JSON(http.StatusOK, toUserResponse(user))
}

func (h *Handler) GetAllUsers(c *gin.Context) {
	users, err := h.service.GetAllUsers()
	if err != nil {
		h.writeError(c, h.mapError(err))
		return
	}

	c.JSON(http.StatusOK, toUserResponses(users))
}

func (h *Handler) mapError(err error) *secureerr.Error {
	if errors.Is(err, ErrUserNotFound) {
		return secureerr.NotFound("USER_NOT_FOUND", "user not found", err, nil)
	}
	if errors.Is(err, ErrUserAlreadyExists) {
		return secureerr.Conflict("USER_ALREADY_EXISTS", "user already exists", err, nil)
	}
	if errors.Is(err, ErrInvalidUserID) {
		return secureerr.BadRequest("INVALID_USER_ID", "invalid user id", err, nil)
	}
	return secureerr.Internal("INTERNAL_ERROR", "an unexpected error occurred", err, nil)
}

func (h *Handler) writeError(c *gin.Context, appErr *secureerr.Error) {
	log.Printf("user handler error: %s", appErr.LogString())

	c.JSON(appErr.StatusCode(), gin.H{
		"code":  appErr.Code,
		"error": appErr.Error(),
	})
}

func toUserResponse(u *domain.User) userResponse {
	return userResponse{
		ID:        u.ID,
		Name:      u.Name,
		Surname:   u.Surname,
		Email:     u.Email,
		Status:    u.UserStatus,
		CreatedAt: u.CreatedAt,
	}
}

func toUserResponses(users []*domain.User) []userResponse {
	responses := make([]userResponse, 0, len(users))
	for _, u := range users {
		responses = append(responses, toUserResponse(u))
	}
	return responses
}
