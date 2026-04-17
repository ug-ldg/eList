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

func (s *TaskService) CreateTask(ctx context.Context, title string, parentID *int) (*model.Task, error) {
	if title == "" {
		return nil, errors.New("title cannot be empty")
	}

	return s.repo.CreateTask(ctx, title, parentID)
}

func (s *TaskService) GetTask(ctx context.Context, id int) (*model.Task, bool, error) {
	task, err := s.cache.Get(ctx, id)
	if err == nil {
		return task, true, nil
	}

	if err != redis.Nil {
		return nil, false, err
	}

	task, err = s.repo.GetTask(ctx, id)
	if err != nil {
		return nil, false, err
	}

	_ = s.cache.Set(ctx, task)
	return task, false, nil
}

func (s *TaskService) GetChildren(ctx context.Context, parentID int) ([]model.Task, error) {
	return s.repo.GetChildren(ctx, parentID)
}

func (s *TaskService) UpdateTaskStatus(ctx context.Context, id int, status string) (*model.Task, error) {
	if !validStatuses[status] {
		return nil, errors.New("invalid status: must be pending, in_progress or done")
	}

	task, err := s.repo.UpdateTaskStatus(ctx, id, status)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Delete(ctx, id)
	return task, nil
}
