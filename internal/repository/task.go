package repository

import (
	"context"
	"errors"
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

func (r *TaskRepository) CreateTask(ctx context.Context, userID int, title string, parentID *int, note *string, icon string) (*model.Task, error) {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO tasks (user_id, title, parent_id, note, icon) VALUES($1, $2, $3, $4, $5)
		RETURNING id, user_id, title, parent_id, status, note, icon, created_at, updated_at`,
		userID, title, parentID, note, icon,
	)

	var t model.Task
	err := row.Scan(&t.ID, &t.UserID, &t.Title, &t.ParentID, &t.Status, &t.Note, &t.Icon, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *TaskRepository) GetTask(ctx context.Context, userID int, id int) (*model.Task, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, title, parent_id, status, note, icon, created_at, updated_at
		FROM tasks WHERE user_id = $1 AND id = $2`,
		userID, id,
	)

	var t model.Task
	err := row.Scan(&t.ID, &t.UserID, &t.Title, &t.ParentID, &t.Status, &t.Note, &t.Icon, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *TaskRepository) GetChildren(ctx context.Context, userID int, parentID int) ([]model.Task, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, title, parent_id, status, note, icon, created_at, updated_at
		FROM tasks WHERE user_id = $1 AND parent_id = $2 ORDER BY title ASC`,
		userID, parentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.ParentID, &t.Status, &t.Note, &t.Icon, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (r *TaskRepository) UpdateTaskStatus(ctx context.Context, userID int, id int, status string) (*model.Task, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE tasks SET status = $1, updated_at = $2 WHERE user_id = $3 AND id = $4
		RETURNING id, user_id, title, parent_id, status, note, icon, created_at, updated_at`,
		status, time.Now(), userID, id,
	)

	var t model.Task
	err := row.Scan(&t.ID, &t.UserID, &t.Title, &t.ParentID, &t.Status, &t.Note, &t.Icon, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *TaskRepository) UpdateTask(ctx context.Context, userID int, id int, title string, status string, note *string, icon string) (*model.Task, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE tasks SET title = $1, status = $2, note = $3, icon = $4, updated_at = $5
		WHERE id = $6 AND user_id = $7
		RETURNING id, user_id, title, parent_id, status, note, icon, created_at, updated_at`,
		title, status, note, icon, time.Now(), id, userID,
	)

	var t model.Task
	err := row.Scan(&t.ID, &t.UserID, &t.Title, &t.ParentID, &t.Status, &t.Note, &t.Icon, &t.CreatedAt, &t.UpdatedAt)
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

	root, ok := nodes[rootID]
	if !ok || root == nil {
		return nil, errors.New("task not found")
	}
	return root, nil
}

func (r *TaskRepository) GetAncestors(ctx context.Context, userID int, id int) ([]model.Breadcrumb, error) {
	rows, err := r.pool.Query(ctx, `
		WITH RECURSIVE ancestors AS (
			SELECT id, title, parent_id
			FROM tasks WHERE id = $1 AND user_id = $2
			UNION ALL
			SELECT t.id, t.title, t.parent_id
			FROM tasks t
			INNER JOIN ancestors a ON t.id = a.parent_id
		)
		SELECT id, title FROM ancestors
	`, id, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var crumbs []model.Breadcrumb
	for rows.Next() {
		var b model.Breadcrumb
		if err := rows.Scan(&b.ID, &b.Title); err != nil {
			return nil, err
		}
		crumbs = append(crumbs, b)
	}

	for i, j := 0, len(crumbs)-1; i < j; i, j = i+1, j-1 {
		crumbs[i], crumbs[j] = crumbs[j], crumbs[i]
	}

	return crumbs, nil
}

func (r *TaskRepository) GetRootTasks(ctx context.Context, userID int) ([]model.Task, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, title, parent_id, status, note, icon, created_at, updated_at
		FROM tasks WHERE user_id = $1 AND parent_id IS NULL ORDER BY title ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.ParentID, &t.Status, &t.Note, &t.Icon, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (r *TaskRepository) UpdateTaskParent(ctx context.Context, userID int, id int, parentID *int) (*model.Task, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE tasks SET parent_id = $1, updated_at = $2 WHERE id = $3 AND user_id = $4
		RETURNING id, user_id, title, parent_id, status, note, icon, created_at, updated_at`,
		parentID, time.Now(), id, userID,
	)

	var t model.Task
	err := row.Scan(&t.ID, &t.UserID, &t.Title, &t.ParentID, &t.Status, &t.Note, &t.Icon, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &t, nil
}
