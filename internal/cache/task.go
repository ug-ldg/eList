package cache

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/ug-ldg/elist/internal/model"
)

const taskTTL = 5 * time.Minute

type TaskCache struct {
	client *redis.Client
}

func NewTaskCache(addr string) *TaskCache {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &TaskCache{client: client}
}

func (c *TaskCache) Get(ctx context.Context, userID int, id int) (*model.Task, error) {
	val, err := c.client.Get(ctx, taskKey(userID, id)).Result()
	if err != nil {
		return nil, err
	}

	var task model.Task
	if err := json.Unmarshal([]byte(val), &task); err != nil {
		return nil, err
	}

	return &task, nil
}

func (c *TaskCache) Set(ctx context.Context, task *model.Task) error {
	bytes, err := json.Marshal(task)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, taskKey(task.UserID, task.ID), bytes, taskTTL).Err()
}

func (c *TaskCache) Delete(ctx context.Context, userID int, id int) error {
	return c.client.Del(ctx, taskKey(userID, id)).Err()
}

func taskKey(userID int, id int) string {
	return "task:" + strconv.Itoa(userID) + ":" + strconv.Itoa(id)
}
