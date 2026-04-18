package repository

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ug-ldg/elist/internal/model"
)

type StatsRepository struct {
	pool *pgxpool.Pool
}

func NewStatsRepository(pool *pgxpool.Pool) *StatsRepository {
	return &StatsRepository{pool: pool}
}

func (r *StatsRepository) GetStats(ctx context.Context) (*model.Stats, error) {
	var stats model.Stats
	var mu sync.Mutex
	var wg sync.WaitGroup
	var firstErr error

	queries := []struct {
		sql    string
		target *int
	}{
		{"SELECT COUNT(*) FROM tasks WHERE status = 'pending'", &stats.Pending},
		{"SELECT COUNT(*) FROM tasks WHERE status = 'done'", &stats.Done},
		{"SELECT COUNT(*) FROM tasks WHERE parent_id IS NULL", &stats.RootTasks},
	}

	for _, q := range queries {
		wg.Add(1)
		go func(sql string, target *int) {
			defer wg.Done()
			var count int

			err := r.pool.QueryRow(ctx, sql).Scan(&count)

			mu.Lock()
			defer mu.Unlock()
			if err != nil && firstErr == nil {
				firstErr = err
				return
			}

			*target = count
		}(q.sql, q.target)
	}

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	stats.TotalTasks = stats.Pending + stats.Done
	stats.SubTasks = stats.TotalTasks - stats.RootTasks

	return &stats, nil
}
