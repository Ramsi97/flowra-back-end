package http

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes registers auth routes on the provided router.
// jwtMiddleware is applied only to protected endpoints.
func SetupRoutes(router *gin.Engine, h *AuthHandler, jwtMiddleware gin.HandlerFunc) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)

		// Protected routes
		protected := auth.Group("")
		protected.Use(jwtMiddleware)
		{
			protected.POST("/logout", h.Logout)
		}
	}
}
