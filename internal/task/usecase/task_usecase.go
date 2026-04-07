package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Ramsi97/flowra-back-end/internal/task/domain"
	"github.com/Ramsi97/flowra-back-end/internal/task/repository/interfaces"
	"github.com/Ramsi97/flowra-back-end/pkg/ai"
)

type taskUseCase struct {
	repo   interfaces.TaskRepository
	gemini *ai.GeminiClient
}

func NewTaskUseCase(repo interfaces.TaskRepository, gemini *ai.GeminiClient) domain.TaskUseCase {
	return &taskUseCase{
		repo:   repo,
		gemini: gemini,
	}
}

func (u *taskUseCase) Create(userID string, task *domain.Task) error {
	if task.Title == "" {
		return errors.New("title is required")
	}
	if task.Status == "" {
		task.Status = "todo"
	}
	now := time.Now()
	task.UserID = userID
	task.CreatedAt = now
	task.UpdatedAt = now

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return u.repo.Create(ctx, task)
}

func (u *taskUseCase) GetByID(userID, id string) (*domain.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	task, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("task not found")
	}
	if task.UserID != userID {
		return nil, errors.New("task not found")
	}
	return task, nil
}

func (u *taskUseCase) ListByUser(userID string) ([]domain.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return u.repo.FindByUserID(ctx, userID)
}

func (u *taskUseCase) Update(userID, id string, input domain.UpdateTaskInput) (*domain.Task, error) {
	if _, err := u.GetByID(userID, id); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	updated, err := u.repo.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, errors.New("task not found")
	}
	return updated, nil
}

func (u *taskUseCase) Delete(userID, id string) error {
	if _, err := u.GetByID(userID, id); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return u.repo.Delete(ctx, id)
}

// AI Assistant methods

func (u *taskUseCase) SuggestDraftTasks(ctx context.Context, description string) ([]domain.Task, error) {
	if u.gemini == nil {
		return nil, errors.New("AI assistant not configured")
	}

	var drafts []domain.Task
	err := u.gemini.AnalyzeIntent(ctx, ai.SystemPromptTaskSuggest, description, &drafts)
	if err != nil {
		return nil, err
	}

	return drafts, nil
}

func (u *taskUseCase) RefineDraftTasks(ctx context.Context, drafts []domain.Task, instruction string) ([]domain.Task, error) {
	if u.gemini == nil {
		return nil, errors.New("AI assistant not configured")
	}

	draftsJSON, err := json.Marshal(drafts)
	if err != nil {
		return nil, err
	}

	prompt := ai.BuildChatPrompt(string(draftsJSON), instruction)

	var updatedDrafts []domain.Task
	// We use an empty system prompt here because the buildChatPrompt includes the system context
	err = u.gemini.AnalyzeIntent(ctx, "You are a Task Intelligence Assistant.", prompt, &updatedDrafts)
	if err != nil {
		return nil, err
	}

	return updatedDrafts, nil
}
