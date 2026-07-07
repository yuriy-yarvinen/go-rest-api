package rest

import (
	"errors"
	"net/http"
	"strconv"

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
	if err := h.service.Create(&event); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save event"})
		return
	}
	context.JSON(http.StatusCreated, event)
}

func (h *Handler) update(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		return
	}
	var event events.Event
	if err := context.ShouldBindJSON(&event); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	event.ID = id

	err := h.service.Update(&event)
	if errors.Is(err, events.ErrEventNotFound) {
		context.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}
	if err != nil {
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
	err := h.service.Delete(id)
	if errors.Is(err, events.ErrEventNotFound) {
		context.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}
	if err != nil {
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
