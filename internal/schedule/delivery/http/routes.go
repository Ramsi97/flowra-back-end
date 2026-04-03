package http

import "github.com/gin-gonic/gin"

// SetupRoutes registers all schedule routes. All routes are JWT-protected.
func SetupRoutes(router *gin.Engine, h *ScheduleHandler, jwtMiddleware gin.HandlerFunc) {
	sched := router.Group("/schedule")
	sched.Use(jwtMiddleware)
	{
		sched.POST("/generate", h.Generate)
		sched.POST("/regenerate", h.Regenerate)
		sched.POST("/fix", h.Fix)

		sched.PUT("/item/:id", h.UpdateItem)
		sched.DELETE("/item/:id", h.DeleteItem)
		sched.POST("/item/:id/remove", h.RemoveTask)

		sched.DELETE("/day", h.ClearDay)
		sched.DELETE("/week", h.ClearWeek)
		sched.DELETE("/month", h.ClearMonth)
	}
}
