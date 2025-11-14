package models

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/iben12/counter-app/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

// setupTestDB creates a test database connection and runs migrations.
func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/counter_test?sslmode=disable"
	}

	pool, err := db.NewPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("failed to create test pool: %v", err)
	}

	if err := db.Migrate(ctx, databaseURL); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	cleanup := func() {
		pool.Close()
	}

	return pool, cleanup
}

// TestCreateCounterDuplicateName tests that creating a counter with a duplicate name fails.
func TestCreateCounterDuplicateName(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create first counter
	_, err := CreateCounter(ctx, pool, "duplicate-test", "1d")
	if err != nil {
		t.Fatalf("failed to create first counter: %v", err)
	}

	// Try to create second counter with same name
	_, err = CreateCounter(ctx, pool, "duplicate-test", "2h")
	if err == nil {
		t.Error("expected error when creating duplicate counter, but got none")
	}
	if err != nil && err.Error() == "" {
		t.Error("expected non-empty error message for duplicate name")
	}
}

// TestCreateCounterDefaultFrequency tests that CreateCounter defaults frequency to "1d" if empty.
func TestCreateCounterDefaultFrequency(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	counter, err := CreateCounter(ctx, pool, "default-freq-test", "")
	if err != nil {
		t.Fatalf("failed to create counter with empty frequency: %v", err)
	}

	if counter.Frequency != "1d" {
		t.Errorf("expected frequency '1d', got %q", counter.Frequency)
	}
}

// TestCreateCounterWithValidFrequencies tests that counters can be created with valid frequency values.
func TestCreateCounterWithValidFrequencies(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	validFrequencies := []string{"1h", "2d", "3w"}
	for i, freq := range validFrequencies {
		name := fmt.Sprintf("valid-freq-test-%d", i)
		counter, err := CreateCounter(ctx, pool, name, freq)
		if err != nil {
			t.Errorf("failed to create counter with frequency %q: %v", freq, err)
		}
		if counter.Frequency != freq {
			t.Errorf("expected frequency %q, got %q", freq, counter.Frequency)
		}
	}
}

// TestGetCounterByIDNotFound tests that GetCounterByID returns an error for non-existent counter.
func TestGetCounterByIDNotFound(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Try to get a counter with an ID that doesn't exist (e.g., 999999)
	counter, err := GetCounterByID(ctx, pool, 999999)
	if err == nil {
		t.Error("expected error when getting non-existent counter, but got none")
	}
	if counter != nil {
		t.Errorf("expected nil counter for non-existent ID, got %+v", counter)
	}
}

// TestGetCounterByIDSuccess tests that GetCounterByID returns the correct counter.
func TestGetCounterByIDSuccess(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a counter
	created, err := CreateCounter(ctx, pool, "get-test", "2d")
	if err != nil {
		t.Fatalf("failed to create counter: %v", err)
	}

	// Retrieve it
	retrieved, err := GetCounterByID(ctx, pool, created.ID)
	if err != nil {
		t.Fatalf("failed to get counter: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, retrieved.ID)
	}
	if retrieved.Name != "get-test" {
		t.Errorf("expected name 'get-test', got %q", retrieved.Name)
	}
	if retrieved.Frequency != "2d" {
		t.Errorf("expected frequency '2d', got %q", retrieved.Frequency)
	}
}

// TestUpdateCounterFrequencySuccess tests that UpdateCounterFrequency updates correctly.
func TestUpdateCounterFrequencySuccess(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a counter with initial frequency
	counter, err := CreateCounter(ctx, pool, "update-test", "1h")
	if err != nil {
		t.Fatalf("failed to create counter: %v", err)
	}

	// Update frequency
	updated, err := UpdateCounterFrequency(ctx, pool, counter.ID, "3d")
	if err != nil {
		t.Fatalf("failed to update counter frequency: %v", err)
	}

	if updated.Frequency != "3d" {
		t.Errorf("expected frequency '3d', got %q", updated.Frequency)
	}
	if updated.ID != counter.ID {
		t.Errorf("expected ID %d, got %d", counter.ID, updated.ID)
	}
}

// TestUpdateCounterFrequencyNotFound tests that UpdateCounterFrequency fails for non-existent counter.
func TestUpdateCounterFrequencyNotFound(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Try to update a counter that doesn't exist
	_, err := UpdateCounterFrequency(ctx, pool, 999999, "2h")
	if err == nil {
		t.Error("expected error when updating non-existent counter, but got none")
	}
}

// TestGetAllCounters tests that GetAllCounters returns all created counters.
func TestGetAllCounters(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a few counters
	_, err := CreateCounter(ctx, pool, "all-test-1", "1h")
	if err != nil {
		t.Fatalf("failed to create counter 1: %v", err)
	}
	_, err = CreateCounter(ctx, pool, "all-test-2", "1d")
	if err != nil {
		t.Fatalf("failed to create counter 2: %v", err)
	}

	// Get all counters
	counters, err := GetAllCounters(ctx, pool)
	if err != nil {
		t.Fatalf("failed to get all counters: %v", err)
	}

	if len(counters) < 2 {
		t.Errorf("expected at least 2 counters, got %d", len(counters))
	}

	// Verify our test counters are in the list
	names := make(map[string]bool)
	for _, c := range counters {
		names[c.Name] = true
	}
	if !names["all-test-1"] {
		t.Error("counter 'all-test-1' not found in results")
	}
	if !names["all-test-2"] {
		t.Error("counter 'all-test-2' not found in results")
	}
}

// TestCounterNameConstraint tests that counter names are case-sensitive and respect uniqueness.
func TestCounterNameConstraint(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a counter
	_, err := CreateCounter(ctx, pool, "case-test", "1d")
	if err != nil {
		t.Fatalf("failed to create counter: %v", err)
	}

	// Try to create with different case (should succeed because names are case-sensitive)
	_, err = CreateCounter(ctx, pool, "Case-Test", "1d")
	if err != nil {
		t.Errorf("expected success for case-different name, got error: %v", err)
	}

	// But exact duplicate should fail
	_, err = CreateCounter(ctx, pool, "case-test", "1d")
	if err == nil {
		t.Error("expected error for exact duplicate name, but got none")
	}
}

// TestCounterCreatedAtNotZero tests that created_at timestamp is populated.
func TestCounterCreatedAtNotZero(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	counter, err := CreateCounter(ctx, pool, "timestamp-test", "1d")
	if err != nil {
		t.Fatalf("failed to create counter: %v", err)
	}

	if counter.CreatedAt == "" {
		t.Error("expected non-empty created_at timestamp")
	}
}
