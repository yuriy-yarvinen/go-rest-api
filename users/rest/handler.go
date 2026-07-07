package rest

import (
	"errors"
	"net/http"
	"strconv"

	"go-rest-api/users"
	"go-rest-api/utils"

	"github.com/gin-gonic/gin"
)

// Handler exposes the users service over HTTP.
type Handler struct {
	service *users.Service
}

func NewHandler(service *users.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) login(context *gin.Context) {
	var user users.User
	if err := context.ShouldBindJSON(&user); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.Login(&user); err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": "Login successful"})
}

func (h *Handler) register(context *gin.Context) {
	var user users.User
	if err := context.ShouldBindJSON(&user); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := utils.ValidateEmailDomain(user.Email); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.Register(&user); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}
	user.Password = ""
	context.JSON(http.StatusCreated, user)
}

func (h *Handler) getByID(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		context.JSON(http.StatusNotFound, gin.H{"error": "Invalid user ID"})
		return
	}
	user, err := h.service.GetByID(id)
	if errors.Is(err, users.ErrUserNotFound) {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}
	context.JSON(http.StatusOK, user)
}

func (h *Handler) update(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		context.JSON(http.StatusNotFound, gin.H{"error": "Invalid user ID"})
		return
	}
	var user users.User
	if err := context.ShouldBindJSON(&user); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user.ID = id

	err := h.service.Update(&user)
	if errors.Is(err, users.ErrUserNotFound) {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}
	context.JSON(http.StatusOK, user)
}

func (h *Handler) delete(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		return
	}
	err := h.service.Delete(id)
	if errors.Is(err, users.ErrUserNotFound) {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// parseID reads and validates the :id path param. On failure it writes a 400
// and returns ok=false so the caller can just return.
func parseID(context *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return 0, false
	}
	return id, true
}
