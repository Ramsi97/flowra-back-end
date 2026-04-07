package http

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes registers authentication routes.
func SetupRoutes(router *gin.Engine, h *AuthHandler, jwtMiddleware gin.HandlerFunc) {
	auth := router.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/register", h.Register)
		auth.POST("/logout", jwtMiddleware, h.Logout)
	}
}
