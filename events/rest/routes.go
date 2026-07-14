package rest

import "github.com/gin-gonic/gin"

// authRequired guards routes that mutate data; it's provided by the caller
// so this package doesn't need to depend on the users package or JWT.
func RegisterRoutes(router *gin.RouterGroup, h *Handler, authRequired gin.HandlerFunc) {
	router.GET("events", h.list)
	router.GET("events/:id", h.getByID)
	router.POST("events", authRequired, h.create)
	router.PUT("events/:id", authRequired, h.update)
	router.DELETE("events/:id", authRequired, h.delete)
}
