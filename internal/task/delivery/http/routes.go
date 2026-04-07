package http

import "github.com/gin-gonic/gin"

// SetupRoutes registers all task routes. All routes are JWT-protected.
func SetupRoutes(router *gin.RouterGroup, h *TaskHandler) {
	tasks := router.Group("/tasks")
	{
		tasks.POST("", h.CreateTask)
		tasks.GET("", h.ListTasks)
		tasks.GET("/:id", h.GetTask)
		tasks.PUT("/:id", h.UpdateTask)
		tasks.DELETE("/:id", h.DeleteTask)

		// AI Assistant routes
		tasks.POST("/ai/suggest", h.SuggestTasks)
		tasks.POST("/ai/chat", h.RefineTasks)
	}
}
