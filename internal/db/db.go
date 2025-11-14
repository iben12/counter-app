package db

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5/pgxpool"
)

// NewPool creates a new pgxpool.Pool from a DATABASE_URL.
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
    pool, err := pgxpool.New(ctx, databaseURL)
    if err != nil {
        return nil, fmt.Errorf("pgxpool.New: %w", err)
    }
    return pool, nil
}

// Migrate runs small, idempotent SQL migration(s) needed for development.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
    sql := `
    CREATE TABLE IF NOT EXISTS counters (
        id SERIAL PRIMARY KEY,
        name TEXT NOT NULL UNIQUE,
        value BIGINT NOT NULL DEFAULT 0,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
    );
    `
    _, err := pool.Exec(ctx, sql)
    return err
}
