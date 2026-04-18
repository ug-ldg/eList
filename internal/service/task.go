package service

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"github.com/ug-ldg/elist/internal/cache"
	"github.com/ug-ldg/elist/internal/model"
	"github.com/ug-ldg/elist/internal/repository"
)

var validStatuses = map[string]bool{
	"pending":     true,
	"in_progress": true,
	"done":        true,
}

type TaskService struct {
	repo  *repository.TaskRepository
	cache *cache.TaskCache
}

func NewTaskService(repo *repository.TaskRepository, cache *cache.TaskCache) *TaskService {
	return &TaskService{repo: repo, cache: cache}
}

func (s *TaskService) CreateTask(ctx context.Context, userID int, title string, parentID *int) (*model.Task, error) {
	if title == "" {
		return nil, errors.New("title cannot be empty")
	}

	return s.repo.CreateTask(ctx, userID, title, parentID)
}

func (s *TaskService) GetTask(ctx context.Context, userID int, id int) (*model.Task, bool, error) {
	task, err := s.cache.Get(ctx, userID, id)
	if err == nil {
		return task, true, nil
	}

	if err != redis.Nil {
		return nil, false, err
	}

	task, err = s.repo.GetTask(ctx, userID, id)
	if err != nil {
		return nil, false, err
	}

	_ = s.cache.Set(ctx, task)
	return task, false, nil
}

func (s *TaskService) GetChildren(ctx context.Context, userID int, parentID int) ([]model.Task, error) {
	return s.repo.GetChildren(ctx, userID, parentID)
}

func (s *TaskService) UpdateTaskStatus(ctx context.Context, userID int, id int, status string) (*model.Task, error) {
	if !validStatuses[status] {
		return nil, errors.New("invalid status: must be pending, in_progress or done")
	}

	task, err := s.repo.UpdateTaskStatus(ctx, userID, id, status)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Delete(ctx, userID, id)
	return task, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, userID int, id int) error {
	return s.repo.DeleteTask(ctx, userID, id)
}

func (s *TaskService) GetTree(ctx context.Context, userID int, id int) (*model.TaskNode, error) {
	task, err := s.repo.GetTree(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	return task, nil
}
