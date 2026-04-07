package http

import (
	"net/http"
	"time"

	"github.com/Ramsi97/flowra-back-end/internal/focus/domain"
	"github.com/gin-gonic/gin"
)

type FocusHandler struct {
	usecase domain.FocusUseCase
}

func NewFocusHandler(uc domain.FocusUseCase) *FocusHandler {
	return &FocusHandler{usecase: uc}
}

// GET /focus/status
func (h *FocusHandler) GetStatus(c *gin.Context) {
	userID := c.GetString("userID")
	now := time.Now().UTC()

	status, err := h.usecase.GetStatus(c.Request.Context(), userID, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// PUT /focus/config
func (h *FocusHandler) UpdateConfig(c *gin.Context) {
	userID := c.GetString("userID")
	var input domain.UpdateConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.usecase.UpdateConfig(c.Request.Context(), userID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "focus configuration updated"})
}
