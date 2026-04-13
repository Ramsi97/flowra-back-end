package http

import (
	"net/http"

	"github.com/Ramsi97/flowra-back-end/internal/schedule/domain"
	"github.com/gin-gonic/gin"
)

type ScheduleHandler struct {
	usecase domain.ScheduleUseCase
}

func NewScheduleHandler(uc domain.ScheduleUseCase) *ScheduleHandler {
	return &ScheduleHandler{usecase: uc}
}

// POST /schedule/generate
func (h *ScheduleHandler) Generate(c *gin.Context) {
	userID := c.GetString("userID")
	var req domain.GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	items, err := h.usecase.Generate(userID, req.Date)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// POST /schedule/regenerate
func (h *ScheduleHandler) Regenerate(c *gin.Context) {
	userID := c.GetString("userID")
	var req domain.GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	items, err := h.usecase.Regenerate(userID, req.Date)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// PUT /schedule/item/:id
func (h *ScheduleHandler) UpdateItem(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")
	var input domain.UpdateItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	items, err := h.usecase.UpdateItem(userID, id, input)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// DELETE /schedule/item/:id
func (h *ScheduleHandler) DeleteItem(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")
	if err := h.usecase.DeleteItem(userID, id); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "schedule item deleted"})
}

// DELETE /schedule/day?date=YYYY-MM-DD
func (h *ScheduleHandler) ClearDay(c *gin.Context) {
	userID := c.GetString("userID")
	date := c.Query("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date query param required (YYYY-MM-DD)"})
		return
	}
	if err := h.usecase.ClearDay(userID, date); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "day cleared"})
}

// DELETE /schedule/week?start=YYYY-MM-DD
func (h *ScheduleHandler) ClearWeek(c *gin.Context) {
	userID := c.GetString("userID")
	start := c.Query("start")
	if start == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start query param required (YYYY-MM-DD)"})
		return
	}
	if err := h.usecase.ClearWeek(userID, start); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "week cleared"})
}

// DELETE /schedule/month?month=YYYY-MM
func (h *ScheduleHandler) ClearMonth(c *gin.Context) {
	userID := c.GetString("userID")
	month := c.Query("month")
	if month == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "month query param required (YYYY-MM)"})
		return
	}
	if err := h.usecase.ClearMonth(userID, month); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "month cleared"})
}

// POST /schedule/fix
func (h *ScheduleHandler) Fix(c *gin.Context) {
	userID := c.GetString("userID")
	var req domain.FixRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	items, err := h.usecase.Fix(userID, req.CurrentTime)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// POST /schedule/item/:id/remove
func (h *ScheduleHandler) RemoveTask(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")
	if err := h.usecase.RemoveTask(userID, id); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "item removed from schedule, task reset to todo"})
}

// AI Assistant Handlers

type aiScheduleRequest struct {
	Date   string `json:"date" binding:"required"`
	Prompt string `json:"prompt" binding:"required"`
}

func (h *ScheduleHandler) AISchedule(c *gin.Context) {
	userID := c.GetString("userID")
	var req aiScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items, err := h.usecase.AISchedule(c.Request.Context(), userID, req.Date, req.Prompt)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, items)
}
