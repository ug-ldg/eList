package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ug-ldg/elist/internal/model"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Upsert(ctx context.Context, provider, providerID, email, name string) (*model.User, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO users (provider, provider_id, email, name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (provider, provider_id) DO UPDATE
			SET email = EXCLUDED.email,
			    name  = EXCLUDED.name
		RETURNING id, provider, provider_id, email, name, created_at
	`, provider, providerID, email, name)

	var u model.User
	err := row.Scan(&u.ID, &u.Provider, &u.ProviderID, &u.Email, &u.Name, &u.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &u, nil
}
