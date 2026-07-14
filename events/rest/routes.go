package rest

import "github.com/gin-gonic/gin"

// authRequired guards routes that mutate data; it's provided by the caller
// so this package doesn't need to depend on the users package or JWT.
func RegisterRoutes(router *gin.RouterGroup, h *Handler, authRequired gin.HandlerFunc) {
	router.GET("events", h.list)
	router.GET("events/:id", h.getByID)

	authenticated := router.Group("/")
	authenticated.Use(authRequired)
	authenticated.POST("events", h.create)
	authenticated.PUT("events/:id", h.update)
	authenticated.DELETE("events/:id", h.delete)

}
