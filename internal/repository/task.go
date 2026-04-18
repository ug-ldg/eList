package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ug-ldg/elist/internal/model"
)

type TaskRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{pool: pool}
}

func (r *TaskRepository) CreateTask(ctx context.Context, title string, parentID *int) (*model.Task, error) {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO tasks (title, parent_id) VALUES($1, $2)
		RETURNING id, title, parent_id, status, created_at, updated_at`,
		title, parentID,
	)

	var t model.Task
	err := row.Scan(&t.ID, &t.Title, &t.ParentID, &t.Status, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *TaskRepository) GetTask(ctx context.Context, id int) (*model.Task, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, title, parent_id, status, created_at, updated_at FROM tasks WHERE id = $1`,
		id,
	)

	var t model.Task
	err := row.Scan(&t.ID, &t.Title, &t.ParentID, &t.Status, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *TaskRepository) GetChildren(ctx context.Context, parentID int) ([]model.Task, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, title, parent_id, status, created_at, updated_at FROM tasks WHERE parent_id = $1`,
		parentID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.ParentID, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (r *TaskRepository) UpdateTaskStatus(ctx context.Context, id int, status string) (*model.Task, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE tasks SET status = $1, updated_at = $2 WHERE id = $3
		RETURNING id, title, parent_id, status, created_at, updated_at`,
		status, time.Now(), id,
	)

	var t model.Task
	err := row.Scan(&t.ID, &t.Title, &t.ParentID, &t.Status, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *TaskRepository) DeleteTask(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM tasks WHERE id = $1`,
		id,
	)

	return err
}
