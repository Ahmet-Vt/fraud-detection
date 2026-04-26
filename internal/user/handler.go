package user

import (
	"errors"
	"log"
	"net/http"
	"time"

	"fraud-detection/internal/domain"
	"fraud-detection/internal/secureerr"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

type createUserRequest struct {
	Name    string  `json:"name" binding:"required,max=100" example:"Ahmet" extensions:"x-order=0"`
	Surname string  `json:"surname" binding:"required,max=100" example:"Vt" extensions:"x-order=1"`
	Email   *string `json:"email" binding:"omitempty,email" example:"ahmet@example.com" extensions:"x-order=2"`
	Status  string  `json:"status" binding:"required,oneof=active suspended deleted" enums:"active,suspended,deleted" example:"active" extensions:"x-order=3"`
}

type CreateUserRequest = createUserRequest

type userResponse struct {
	ID        uuid.UUID         `json:"id" format:"uuid" example:"4ca5e74c-3a07-4010-9f07-14e90a35868b" extensions:"x-order=0"`
	Name      string            `json:"name" example:"Ahmet" extensions:"x-order=1"`
	Surname   string            `json:"surname" example:"Vt" extensions:"x-order=2"`
	Email     *string           `json:"email,omitempty" example:"ahmet@example.com" extensions:"x-order=3"`
	Status    domain.UserStatus `json:"status" enums:"active,suspended,deleted" example:"active" extensions:"x-order=4"`
	CreatedAt time.Time         `json:"created_at" swaggertype:"string" format:"date-time" example:"2026-04-15T19:00:00Z" extensions:"x-order=5"`
}

type UserResponse = userResponse

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router gin.IRouter) {
	router.POST("/users", h.CreateUser)
	router.GET("/users/:id", h.GetUserByID)
	router.GET("/users", h.GetAllUsers)
}

// CreateUser godoc
// @Summary Create user
// @Description Creates a new user.
// @Tags users
// @Accept json
// @Produce json
// @Param payload body CreateUserRequest true "Create user payload"
// @Success 201 {object} UserResponse
// @Failure 400 {object} apidoc.InvalidRequestBodyError
// @Failure 409 {object} apidoc.UserAlreadyExistsError
// @Failure 500 {object} apidoc.InternalServerError
// @Router /users [post]
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

// GetUserByID godoc
// @Summary Get user by ID
// @Description Returns a single user by UUID.
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} UserResponse
// @Failure 400 {object} apidoc.InvalidUserIDError
// @Failure 404 {object} apidoc.UserNotFoundError
// @Failure 500 {object} apidoc.InternalServerError
// @Router /users/{id} [get]
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

// GetAllUsers godoc
// @Summary List users
// @Description Returns all users in deterministic order.
// @Tags users
// @Produce json
// @Success 200 {array} UserResponse
// @Failure 500 {object} apidoc.InternalServerError
// @Router /users [get]
func (h *Handler) GetAllUsers(c *gin.Context) {
	users, err := h.service.GetAllUsers()
	if err != nil {
		h.writeError(c, h.mapError(err))
		return
	}

	c.JSON(http.StatusOK, toUserResponses(users))
}

func (h *Handler) mapError(err error) error {
	if errors.Is(err, ErrUserNotFound) {
		return secureerr.NotFound("USER_NOT_FOUND", "user not found", err, nil)
	}
	if errors.Is(err, ErrUserAlreadyExists) {
		return secureerr.Conflict("USER_ALREADY_EXISTS", "user already exists", err, nil)
	}
	if errors.Is(err, ErrInvalidUserID) {
		return secureerr.BadRequest("INVALID_USER_ID", "invalid user id", err, nil)
	}
	return secureerr.FromError(err)
}

func (h *Handler) writeError(c *gin.Context, err error) {
	appErr := secureerr.FromError(err)
	if appErr == nil {
		return
	}

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
