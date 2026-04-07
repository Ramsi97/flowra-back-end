package http

import "github.com/gin-gonic/gin"

// SetupRoutes registers all focus routes. All routes are JWT-protected.
func SetupRoutes(router *gin.RouterGroup, h *FocusHandler) {
	focus := router.Group("/focus")
	{
		focus.GET("/status", h.GetStatus)
		focus.PUT("/config", h.UpdateConfig)
	}
}
