package user

import (
	"log"
	"net/http"

	"github.com/Ahmet-Vt/fraud-detection/internal/domain"
	"github.com/Ahmet-Vt/fraud-detection/internal/secureerr"
	"github.com/gin-gonic/gin"
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
		h.writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *Handler) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	user, err := h.service.GetUserByID(id)
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *Handler) GetAllUsers(c *gin.Context) {
	users, err := h.service.GetAllUsers()
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *Handler) writeError(c *gin.Context, err error) {
	appErr := secureerr.FromError(err)
	log.Printf("user handler error: %s", appErr.LogString())

	c.JSON(appErr.StatusCode(), gin.H{
		"code":  appErr.Code,
		"error": appErr.Error(),
	})
}
