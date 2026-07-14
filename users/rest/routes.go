package rest

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.RouterGroup, h *Handler, authRequired gin.HandlerFunc) {
	router.POST("login", h.login)
	router.POST("register", h.register)

	authenticated := router.Group("/")
	authenticated.Use(authRequired)
	authenticated.GET("users/:id", h.getByID)
	authenticated.PUT("users/:id", h.update)
	authenticated.DELETE("users/:id", h.delete)
}
