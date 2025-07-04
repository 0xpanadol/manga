package handler

import (
	"net/http"

	"github.com/0xpanadol/manga/internal/service"
	"github.com/0xpanadol/manga/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get(middleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	user, err := h.userService.GetProfile(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not retrieve user profile"})
		return
	}

	// Use the same userResponse struct from auth_handler
	c.JSON(http.StatusOK, userResponse{
		ID:       user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
	})
}
