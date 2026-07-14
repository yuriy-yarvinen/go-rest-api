package rest

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.RouterGroup, h *Handler, authRequired gin.HandlerFunc) {
	router.POST("login", h.login)
	router.POST("register", h.register)
	router.GET("users/:id", h.getByID)
	router.PUT("users/:id", authRequired, h.update)
	router.DELETE("users/:id", authRequired, h.delete)
}
