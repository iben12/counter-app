package models

import (
    "context"
    "errors"

    "github.com/jackc/pgx/v5/pgxpool"
)

type Counter struct {
    ID        int64  `json:"id"`
    Name      string `json:"name"`
    Value     int64  `json:"value"`
    CreatedAt string `json:"created_at"`
}

func CreateCounter(ctx context.Context, pool *pgxpool.Pool, name string) (*Counter, error) {
    var c Counter
    err := pool.QueryRow(ctx, "INSERT INTO counters (name, value) VALUES ($1, 0) RETURNING id, name, value, created_at", name).Scan(&c.ID, &c.Name, &c.Value, &c.CreatedAt)
    if err != nil {
        return nil, err
    }
    return &c, nil
}

func GetAllCounters(ctx context.Context, pool *pgxpool.Pool) ([]Counter, error) {
    rows, err := pool.Query(ctx, "SELECT id, name, value, created_at FROM counters ORDER BY id")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []Counter
    for rows.Next() {
        var c Counter
        if err := rows.Scan(&c.ID, &c.Name, &c.Value, &c.CreatedAt); err != nil {
            return nil, err
        }
        out = append(out, c)
    }
    return out, nil
}

func GetCounterByID(ctx context.Context, pool *pgxpool.Pool, id int64) (*Counter, error) {
    var c Counter
    err := pool.QueryRow(ctx, "SELECT id, name, value, created_at FROM counters WHERE id=$1", id).Scan(&c.ID, &c.Name, &c.Value, &c.CreatedAt)
    if err != nil {
        return nil, err
    }
    return &c, nil
}

func IncrementCounter(ctx context.Context, pool *pgxpool.Pool, id int64, delta int64) (*Counter, error) {
    // Update and return new row
    var c Counter
    tag, err := pool.Exec(ctx, "UPDATE counters SET value = value + $1 WHERE id = $2", delta, id)
    if err != nil {
        return nil, err
    }
    if tag.RowsAffected() == 0 {
        return nil, errors.New("not found")
    }
    return GetCounterByID(ctx, pool, id)
}
