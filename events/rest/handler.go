package rest

import (
	"errors"
	"net/http"
	"strconv"

	"go-rest-api/authctx"
	"go-rest-api/events"

	"github.com/gin-gonic/gin"
)

// Handler exposes the events service over HTTP.
type Handler struct {
	service *events.Service
}

func NewHandler(service *events.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) list(context *gin.Context) {
	list, err := h.service.List()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve events"})
		return
	}
	context.JSON(http.StatusOK, gin.H{"events": list})
}

func (h *Handler) getByID(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		context.JSON(http.StatusNotFound, gin.H{"error": "Invalid event ID"})
		return
	}
	event, err := h.service.GetByID(id)
	if errors.Is(err, events.ErrEventNotFound) {
		context.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event"})
		return
	}
	context.JSON(http.StatusOK, event)
}

func (h *Handler) create(context *gin.Context) {
	var event events.Event
	if err := context.ShouldBindJSON(&event); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, ok := authctx.UserID(context)
	if !ok {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Missing authenticated user"})
		return
	}
	// The owner is always the authenticated caller, never whatever the
	// client put in the request body.
	event.UserID = userID

	if err := h.service.Create(&event); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save event"})
		return
	}
	context.JSON(http.StatusCreated, event)
}

func (h *Handler) update(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		context.JSON(http.StatusNotFound, gin.H{"error": "Invalid event ID"})
		return
	}

	userID, ok := authctx.UserID(context)
	if !ok {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Missing authenticated user"})
		return
	}

	existing, err := h.service.GetByID(id)
	if errors.Is(err, events.ErrEventNotFound) {
		context.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event"})
		return
	}
	if existing.UserID != userID {
		context.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own events"})
		return
	}

	var event events.Event
	if err := context.ShouldBindJSON(&event); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	event.ID = id
	// Ownership isn't transferable through a plain update.
	event.UserID = existing.UserID

	if err := h.service.Update(&event); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event"})
		return
	}
	context.JSON(http.StatusOK, event)
}

func (h *Handler) delete(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		return
	}

	userID, ok := authctx.UserID(context)
	if !ok {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Missing authenticated user"})
		return
	}

	existing, err := h.service.GetByID(id)
	if errors.Is(err, events.ErrEventNotFound) {
		context.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event"})
		return
	}
	if existing.UserID != userID {
		context.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own events"})
		return
	}

	if err := h.service.Delete(id); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event"})
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}

// parseID reads and validates the :id path param. On failure it writes a 400
// and returns ok=false so the caller can just return.
func parseID(context *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return 0, false
	}
	return id, true
}
