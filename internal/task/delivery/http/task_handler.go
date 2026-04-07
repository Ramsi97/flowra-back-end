package http

import (
	"net/http"

	"github.com/Ramsi97/flowra-back-end/internal/task/domain"
	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	usecase domain.TaskUseCase
}

func NewTaskHandler(uc domain.TaskUseCase) *TaskHandler {
	return &TaskHandler{usecase: uc}
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	userID := c.GetString("userID")
	var task domain.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.usecase.Create(userID, &task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, task)
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")
	task, err := h.usecase.GetByID(userID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) ListTasks(c *gin.Context) {
	userID := c.GetString("userID")
	tasks, err := h.usecase.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")
	var input domain.UpdateTaskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := h.usecase.Update(userID, id, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")
	if err := h.usecase.Delete(userID, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "task deleted"})
}

// AI Assistant Handlers

type suggestRequest struct {
	Description string `json:"description" binding:"required"`
}

func (h *TaskHandler) SuggestTasks(c *gin.Context) {
	var req suggestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	drafts, err := h.usecase.SuggestDraftTasks(c.Request.Context(), req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, drafts)
}

type refineRequest struct {
	Drafts      []domain.Task `json:"drafts" binding:"required"`
	Instruction string        `json:"instruction" binding:"required"`
}

func (h *TaskHandler) RefineTasks(c *gin.Context) {
	var req refineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.usecase.RefineDraftTasks(c.Request.Context(), req.Drafts, req.Instruction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}
