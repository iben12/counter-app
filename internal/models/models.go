package models

import (
	"context"
	"errors"
	"time"

	"github.com/iben12/counter-app/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Counter struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Frequency string `json:"frequency"`
	Timezone  string `json:"timezone"`
	CreatedAt string `json:"created_at"`
}

func CreateCounter(ctx context.Context, pool *pgxpool.Pool, name string, frequency string, timezone string) (*Counter, error) {
	var c Counter
	if frequency == "" {
		frequency = "1d"
	}
	if timezone == "" {
		timezone = "UTC"
	}
	err := pool.QueryRow(ctx, "INSERT INTO counters (name, frequency, timezone) VALUES ($1, $2, $3) RETURNING id, name, frequency, timezone, created_at::TEXT", name, frequency, timezone).Scan(&c.ID, &c.Name, &c.Frequency, &c.Timezone, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func GetAllCounters(ctx context.Context, pool *pgxpool.Pool) ([]Counter, error) {
	rows, err := pool.Query(ctx, "SELECT id, name, frequency, timezone, created_at::TEXT FROM counters ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Counter
	for rows.Next() {
		var c Counter
		if err := rows.Scan(&c.ID, &c.Name, &c.Frequency, &c.Timezone, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func GetCounterByID(ctx context.Context, pool *pgxpool.Pool, id int64) (*Counter, error) {
	var c Counter
	err := pool.QueryRow(ctx, "SELECT id, name, frequency, timezone, created_at::TEXT FROM counters WHERE id=$1", id).Scan(&c.ID, &c.Name, &c.Frequency, &c.Timezone, &c.CreatedAt)
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

// Count represents a count record with value and expiry aligned to calendar boundaries.
type Count struct {
	ID        int64  `json:"id"`
	CounterID int64  `json:"counter_id"`
	Value     int64  `json:"value"`
	Expiry    string `json:"expiry"`
	CreatedAt string `json:"created_at"`
}

// GetOrCreateCurrentCount retrieves the current (non-expired) count for a counter,
// creating a new one if the existing one has expired.
func GetOrCreateCurrentCount(ctx context.Context, pool *pgxpool.Pool, counterID int64) (*Count, error) {
	counter, err := GetCounterByID(ctx, pool, counterID)
	if err != nil {
		return nil, err
	}

	var c Count
	// Try to get the current (latest) count
	var expiryTime time.Time
	var createdAt time.Time
	err = pool.QueryRow(ctx,
		"SELECT id, counter_id, value, expiry, created_at FROM counts WHERE counter_id = $1 ORDER BY id DESC LIMIT 1",
		counterID).Scan(&c.ID, &c.CounterID, &c.Value, &expiryTime, &createdAt)
	if err != nil {
		// No count exists yet, create one
		return createNewCount(ctx, pool, counterID, counter.Frequency, counter.Timezone)
	}

	c.Expiry = expiryTime.UTC().Format(time.RFC3339)
	c.CreatedAt = createdAt.UTC().Format(time.RFC3339)

	// Check if current count is expired
	if time.Now().UTC().After(expiryTime) {
		// Expired, create new count
		return createNewCount(ctx, pool, counterID, counter.Frequency, counter.Timezone)
	}

	return &c, nil
}

// createNewCount creates a new count record with value 0 and expiry based on the counter's frequency and timezone.
func createNewCount(ctx context.Context, pool *pgxpool.Pool, counterID int64, frequency string, timezone string) (*Count, error) {
	expiry, err := db.NextExpiryTime(frequency, time.Now().UTC(), timezone)
	if err != nil {
		return nil, err
	}

	var c Count
	var expiryTime time.Time
	var createdAt time.Time
	err = pool.QueryRow(ctx,
		"INSERT INTO counts (counter_id, value, expiry) VALUES ($1, 0, $2) RETURNING id, counter_id, value, expiry, created_at",
		counterID, expiry).Scan(&c.ID, &c.CounterID, &c.Value, &expiryTime, &createdAt)
	if err == nil {
		c.Expiry = expiryTime.UTC().Format(time.RFC3339)
		c.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// IncrementCurrentCount increments the current count by delta. If expired, creates a new one first.
// Delta must be non-zero and positive. The count value is clamped to zero (no negative values).
func IncrementCurrentCount(ctx context.Context, pool *pgxpool.Pool, counterID int64, delta int64) (*Count, error) {
	// Validate delta: must be non-zero
	if delta == 0 {
		return nil, errors.New("delta must be non-zero")
	}

	current, err := GetOrCreateCurrentCount(ctx, pool, counterID)
	if err != nil {
		return nil, err
	}

	// Calculate new value and clamp to zero
	newValue := current.Value + delta
	if newValue < 0 {
		newValue = 0
	}

	// Update the value
	var c Count
	var expiryTime time.Time
	var createdAt time.Time
	err = pool.QueryRow(ctx,
		"UPDATE counts SET value = $1 WHERE id = $2 RETURNING id, counter_id, value, expiry, created_at",
		newValue, current.ID).Scan(&c.ID, &c.CounterID, &c.Value, &expiryTime, &createdAt)
	if err == nil {
		c.Expiry = expiryTime.UTC().Format(time.RFC3339)
		c.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// GetCountHistory retrieves all count records for a counter, ordered by creation time descending.
func GetCountHistory(ctx context.Context, pool *pgxpool.Pool, counterID int64) ([]Count, error) {
	rows, err := pool.Query(ctx,
		"SELECT id, counter_id, value, expiry, created_at FROM counts WHERE counter_id = $1 ORDER BY created_at DESC",
		counterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counts []Count
	for rows.Next() {
		var c Count
		var expiryTime time.Time
		var createdAt time.Time
		if err := rows.Scan(&c.ID, &c.CounterID, &c.Value, &expiryTime, &createdAt); err != nil {
			return nil, err
		}
		c.Expiry = expiryTime.UTC().Format(time.RFC3339)
		c.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		counts = append(counts, c)
	}
	return counts, nil
}
