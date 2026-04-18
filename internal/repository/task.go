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

func (r *TaskRepository) CreateTask(ctx context.Context, userID int, title string, parentID *int) (*model.Task, error) {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO tasks (user_id, title, parent_id) VALUES($1, $2, $3)
		RETURNING id, user_id, title, parent_id, status, created_at, updated_at`,
		userID, title, parentID,
	)

	var t model.Task
	err := row.Scan(&t.ID, &t.UserID, &t.Title, &t.ParentID, &t.Status, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *TaskRepository) GetTask(ctx context.Context, userID int, id int) (*model.Task, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, title, parent_id, status, created_at, updated_at FROM tasks WHERE user_id = $1 AND id = $2`,
		userID, id,
	)

	var t model.Task
	err := row.Scan(&t.ID, &t.UserID, &t.Title, &t.ParentID, &t.Status, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *TaskRepository) GetChildren(ctx context.Context, userID int, parentID int) ([]model.Task, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, title, parent_id, status, created_at, updated_at FROM tasks WHERE user_id = $1 AND parent_id = $2`,
		userID, parentID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.ParentID, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (r *TaskRepository) UpdateTaskStatus(ctx context.Context, userID int, id int, status string) (*model.Task, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE tasks SET status = $1, updated_at = $2 WHERE user_id = $3 AND id = $4
		RETURNING id, user_id, title, parent_id, status, created_at, updated_at`,
		status, time.Now(), userID, id,
	)

	var t model.Task
	err := row.Scan(&t.ID, &t.UserID, &t.Title, &t.ParentID, &t.Status, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *TaskRepository) DeleteTask(ctx context.Context, userID int, id int) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM tasks WHERE user_id = $1 AND id = $2`,
		userID, id,
	)

	return err
}

func (r *TaskRepository) GetTree(ctx context.Context, userID int, id int) (*model.TaskNode, error) {
	rows, err := r.pool.Query(ctx, `
			WITH RECURSIVE tree AS (
				SELECT id, title, status, parent_id, created_at, updated_at
				FROM tasks
				WHERE user_id = $1 AND id = $2
				UNION ALL
				SELECT t.id, t.title, t.status, t.parent_id, t.created_at, t.updated_at
				FROM tasks t
				INNER JOIN tree ON t.parent_id = tree.id
			)
			SELECT id, title, status, parent_id, created_at, updated_at FROM tree
		 `, userID, id)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	nodes := make(map[int]*model.TaskNode)
	parents := make(map[int]*int)

	var rootID int

	for rows.Next() {
		var node model.TaskNode
		var parentID *int
		err := rows.Scan(&node.ID, &node.Title, &node.Status, &parentID, &node.CreatedAt, &node.UpdatedAt)
		if err != nil {
			return nil, err
		}
		node.Children = []*model.TaskNode{}
		nodes[node.ID] = &node
		parents[node.ID] = parentID
		if parentID == nil {
			rootID = node.ID
		}
	}

	for _, node := range nodes {
		pid := parents[node.ID]
		if pid != nil {
			nodes[*pid].Children = append(nodes[*pid].Children, node)
		}
	}

	return nodes[rootID], nil
}
