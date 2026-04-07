package http

import "github.com/gin-gonic/gin"

// SetupRoutes registers all schedule routes. All routes are expected to be in a protected group.
func SetupRoutes(router *gin.RouterGroup, h *ScheduleHandler) {
	sched := router.Group("/schedule")
	{
		sched.POST("/generate", h.Generate)
		sched.POST("/regenerate", h.Regenerate)
		sched.PUT("/item/:id", h.UpdateItem)
		sched.DELETE("/item/:id", h.DeleteItem)
		sched.DELETE("/day", h.ClearDay)
		sched.DELETE("/week", h.ClearWeek)
		sched.DELETE("/month", h.ClearMonth)
		sched.POST("/fix", h.Fix)
		sched.DELETE("/task/:id", h.RemoveTask)

		// AI Assistant routes
		sched.POST("/ai/generate", h.AISchedule)
	}
}
