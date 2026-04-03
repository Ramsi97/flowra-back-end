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

// POST /tasks
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

// GET /tasks
func (h *TaskHandler) ListTasks(c *gin.Context) {
	userID := c.GetString("userID")
	tasks, err := h.usecase.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tasks == nil {
		tasks = []domain.Task{}
	}
	c.JSON(http.StatusOK, tasks)
}

// GET /tasks/:id
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

// PUT /tasks/:id
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

// DELETE /tasks/:id
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")
	if err := h.usecase.Delete(userID, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "task deleted"})
}
