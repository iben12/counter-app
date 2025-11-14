package models

import (
    "context"
    "errors"

    "github.com/jackc/pgx/v5/pgxpool"
)

type Counter struct {
    ID        int64  `json:"id"`
    Name      string `json:"name"`
    Frequency string `json:"frequency"`
    CreatedAt string `json:"created_at"`
}

func CreateCounter(ctx context.Context, pool *pgxpool.Pool, name string, frequency string) (*Counter, error) {
    var c Counter
    if frequency == "" {
        frequency = "1d"
    }
    err := pool.QueryRow(ctx, "INSERT INTO counters (name, frequency) VALUES ($1, $2) RETURNING id, name, frequency, created_at::TEXT", name, frequency).Scan(&c.ID, &c.Name, &c.Frequency, &c.CreatedAt)
    if err != nil {
        return nil, err
    }
    return &c, nil
}

func GetAllCounters(ctx context.Context, pool *pgxpool.Pool) ([]Counter, error) {
    rows, err := pool.Query(ctx, "SELECT id, name, frequency, created_at::TEXT FROM counters ORDER BY id")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []Counter
    for rows.Next() {
        var c Counter
        if err := rows.Scan(&c.ID, &c.Name, &c.Frequency, &c.CreatedAt); err != nil {
            return nil, err
        }
        out = append(out, c)
    }
    return out, nil
}

func GetCounterByID(ctx context.Context, pool *pgxpool.Pool, id int64) (*Counter, error) {
    var c Counter
    err := pool.QueryRow(ctx, "SELECT id, name, frequency, created_at::TEXT FROM counters WHERE id=$1", id).Scan(&c.ID, &c.Name, &c.Frequency, &c.CreatedAt)
    if err != nil {
        return nil, err
    }
    return &c, nil
}

func UpdateCounterFrequency(ctx context.Context, pool *pgxpool.Pool, id int64, frequency string) (*Counter, error) {
    tag, err := pool.Exec(ctx, "UPDATE counters SET frequency = $1 WHERE id = $2", frequency, id)
    if err != nil {
        return nil, err
    }
    if tag.RowsAffected() == 0 {
        return nil, errors.New("not found")
    }
    return GetCounterByID(ctx, pool, id)
}
