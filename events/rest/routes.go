package rest

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.RouterGroup, h *Handler) {
	router.GET("events", h.list)
	router.POST("events", h.create)
	router.GET("events/:id", h.getByID)
	router.PUT("events/:id", h.update)
	router.DELETE("events/:id", h.delete)
}
