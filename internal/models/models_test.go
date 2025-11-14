package models

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

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

	// Truncate tables before test to ensure clean state
	pool.Exec(ctx, "TRUNCATE TABLE counts RESTART IDENTITY CASCADE")
	pool.Exec(ctx, "TRUNCATE TABLE counters RESTART IDENTITY CASCADE")

	cleanup := func() {
		// Truncate tables after test to clean up
		pool.Exec(ctx, "TRUNCATE TABLE counts RESTART IDENTITY CASCADE")
		pool.Exec(ctx, "TRUNCATE TABLE counters RESTART IDENTITY CASCADE")
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
	_, err := CreateCounter(ctx, pool, "duplicate-test", "1d", "UTC")
	if err != nil {
		t.Fatalf("failed to create first counter: %v", err)
	}

	// Try to create second counter with same name
	_, err = CreateCounter(ctx, pool, "duplicate-test", "2h", "UTC")
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

	counter, err := CreateCounter(ctx, pool, "default-freq-test", "", "UTC")
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
		counter, err := CreateCounter(ctx, pool, name, freq, "UTC")
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
	created, err := CreateCounter(ctx, pool, "get-test", "2d", "UTC")
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
	counter, err := CreateCounter(ctx, pool, "update-test", "1h", "UTC")
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
	_, err := CreateCounter(ctx, pool, "all-test-1", "1h", "UTC")
	if err != nil {
		t.Fatalf("failed to create counter 1: %v", err)
	}
	_, err = CreateCounter(ctx, pool, "all-test-2", "1d", "UTC")
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
	_, err := CreateCounter(ctx, pool, "case-test", "1d", "UTC")
	if err != nil {
		t.Fatalf("failed to create counter: %v", err)
	}

	// Try to create with different case (should succeed because names are case-sensitive)
	_, err = CreateCounter(ctx, pool, "Case-Test", "1d", "UTC")
	if err != nil {
		t.Errorf("expected success for case-different name, got error: %v", err)
	}

	// But exact duplicate should fail
	_, err = CreateCounter(ctx, pool, "case-test", "1d", "UTC")
	if err == nil {
		t.Error("expected error for exact duplicate name, but got none")
	}
}

// TestCounterCreatedAtNotZero tests that created_at timestamp is populated.
func TestCounterCreatedAtNotZero(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	counter, err := CreateCounter(ctx, pool, "timestamp-test", "1d", "UTC")
	if err != nil {
		t.Fatalf("failed to create counter: %v", err)
	}

	if counter.CreatedAt == "" {
		t.Error("expected non-empty created_at timestamp")
	}
}

// ===== Count Model Edge Case Tests =====

// TestIncrementWithZeroDelta tests that IncrementCurrentCount rejects zero delta.
func TestIncrementWithZeroDelta(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a counter and get initial count
	counter, err := CreateCounter(ctx, pool, "zero-delta-test", "1d", "UTC")
	if err != nil {
		t.Fatalf("failed to create counter: %v", err)
	}

	GetOrCreateCurrentCount(ctx, pool, counter.ID)

	// Try to increment with zero delta
	_, err = IncrementCurrentCount(ctx, pool, counter.ID, 0)
	if err == nil {
		t.Error("expected error when incrementing with zero delta, but got none")
	}
	if err != nil && err.Error() != "delta must be non-zero" {
		t.Errorf("expected 'delta must be non-zero' error, got: %v", err)
	}
}

// TestIncrementWithLargePositiveDelta tests that IncrementCurrentCount accepts large positive deltas.
func TestIncrementWithLargePositiveDelta(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a counter
	counter, err := CreateCounter(ctx, pool, "large-delta-test", "1d", "UTC")
	if err != nil {
		t.Fatalf("failed to create counter: %v", err)
	}

	GetOrCreateCurrentCount(ctx, pool, counter.ID)

	// Increment with large positive delta
	count, err := IncrementCurrentCount(ctx, pool, counter.ID, 1000000)
	if err != nil {
		t.Fatalf("failed to increment with large delta: %v", err)
	}

	if count.Value != 1000000 {
		t.Errorf("expected value 1000000, got %d", count.Value)
	}
}

// TestDecrementWithLargeNegativeDelta tests that IncrementCurrentCount handles large negative deltas (decrement).
func TestDecrementWithLargeNegativeDelta(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a counter
	counter, err := CreateCounter(ctx, pool, "large-negative-delta-test", "1d", "UTC")
	if err != nil {
		t.Fatalf("failed to create counter: %v", err)
	}

	GetOrCreateCurrentCount(ctx, pool, counter.ID)

	// Increment to 100
	IncrementCurrentCount(ctx, pool, counter.ID, 100)

	// Decrement with large negative delta (should clamp to 0)
	count, err := IncrementCurrentCount(ctx, pool, counter.ID, -1000000)
	if err != nil {
		t.Fatalf("failed to decrement with large negative delta: %v", err)
	}

	if count.Value != 0 {
		t.Errorf("expected clamped value 0 after large negative delta, got %d", count.Value)
	}
}

// TestCountValueUnderflowClamp tests that count value is clamped to zero (no negative values).
func TestCountValueUnderflowClamp(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a counter with initial count
	counter, err := CreateCounter(ctx, pool, "underflow-test", "1d", "UTC")
	if err != nil {
		t.Fatalf("failed to create counter: %v", err)
	}

	count, err := GetOrCreateCurrentCount(ctx, pool, counter.ID)
	if err != nil {
		t.Fatalf("failed to get current count: %v", err)
	}

	// Increment to 5
	count, err = IncrementCurrentCount(ctx, pool, counter.ID, 5)
	if err != nil {
		t.Fatalf("failed to increment: %v", err)
	}
	if count.Value != 5 {
		t.Errorf("expected value 5, got %d", count.Value)
	}

	// Now decrement by 10 (should clamp to 0, not go negative)
	count, err = IncrementCurrentCount(ctx, pool, counter.ID, -10)
	if err != nil {
		t.Fatalf("failed to decrement: %v", err)
	}

	if count.Value != 0 {
		t.Errorf("expected clamped value 0 after underflow, got %d", count.Value)
	}
}

// TestCountUnderflowMultipleDecrements tests that multiple decrments clamp to zero.
func TestCountUnderflowMultipleDecrements(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a counter
	counter, err := CreateCounter(ctx, pool, "multi-underflow-test", "1d", "UTC")
	if err != nil {
		t.Fatalf("failed to create counter: %v", err)
	}

	GetOrCreateCurrentCount(ctx, pool, counter.ID)

	// Increment to 3
	count, err := IncrementCurrentCount(ctx, pool, counter.ID, 3)
	if err != nil {
		t.Fatalf("failed to increment: %v", err)
	}
	if count.Value != 3 {
		t.Errorf("expected value 3, got %d", count.Value)
	}

	// Decrement by 2 (should be 1)
	count, err = IncrementCurrentCount(ctx, pool, counter.ID, -2)
	if err != nil {
		t.Fatalf("failed to decrement: %v", err)
	}
	if count.Value != 1 {
		t.Errorf("expected value 1, got %d", count.Value)
	}

	// Decrement by 5 (should clamp to 0, not -4)
	count, err = IncrementCurrentCount(ctx, pool, counter.ID, -5)
	if err != nil {
		t.Fatalf("failed to decrement: %v", err)
	}
	if count.Value != 0 {
		t.Errorf("expected clamped value 0, got %d", count.Value)
	}

	// Decrement again when already at 0 (should stay 0)
	count, err = IncrementCurrentCount(ctx, pool, counter.ID, -1)
	if err != nil {
		t.Fatalf("failed to decrement: %v", err)
	}
	if count.Value != 0 {
		t.Errorf("expected value to remain 0, got %d", count.Value)
	}
}

// TestGetCountHistoryNonExistentCounter tests that GetCountHistory returns error for non-existent counter.
func TestGetCountHistoryNonExistentCounter(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Try to get history for a counter that doesn't exist
	history, err := GetCountHistory(ctx, pool, 999999)
	if err != nil {
		t.Fatalf("expected empty history or no error, got: %v", err)
	}

	// History should be empty (no counts for non-existent counter)
	if len(history) != 0 {
		t.Errorf("expected empty history for non-existent counter, got %d counts", len(history))
	}
}

// TestGetCountHistoryMultipleCounts tests that GetCountHistory returns all counts ordered by creation time.
func TestGetCountHistoryMultipleCounts(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a counter with 1h frequency
	counter, err := CreateCounter(ctx, pool, "history-test", "1h", "UTC")
	if err != nil {
		t.Fatalf("failed to create counter: %v", err)
	}

	// Create first count and increment
	count1, err := GetOrCreateCurrentCount(ctx, pool, counter.ID)
	if err != nil {
		t.Fatalf("failed to get initial count: %v", err)
	}
	IncrementCurrentCount(ctx, pool, counter.ID, 5)

	// Manually create a "historical" count by setting its expiry to the past
	updateExpirySQL := `UPDATE counts SET expiry = now() - interval '2 hours' WHERE id = $1`
	_, err = pool.Exec(ctx, updateExpirySQL, count1.ID)
	if err != nil {
		t.Fatalf("failed to update expiry: %v", err)
	}

	// Get or create a new count (should create a new one since previous is expired)
	count2, err := GetOrCreateCurrentCount(ctx, pool, counter.ID)
	if err != nil {
		t.Fatalf("failed to get new count after expiry: %v", err)
	}

	// Now we should have 2 counts in history
	history, err := GetCountHistory(ctx, pool, counter.ID)
	if err != nil {
		t.Fatalf("failed to get history: %v", err)
	}

	if len(history) < 2 {
		t.Errorf("expected at least 2 counts in history, got %d", len(history))
	}

	// Verify newer count comes first (ordered by creation_at DESC)
	if history[0].ID != count2.ID {
		t.Errorf("expected newest count (ID %d) first, got ID %d", count2.ID, history[0].ID)
	}
}

// TestCountExpiryTime tests that Count expiry time is correct and not empty.
func TestCountExpiryTime(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a counter
	counter, err := CreateCounter(ctx, pool, "expiry-test", "1d", "UTC")
	if err != nil {
		t.Fatalf("failed to create counter: %v", err)
	}

	count, err := GetOrCreateCurrentCount(ctx, pool, counter.ID)
	if err != nil {
		t.Fatalf("failed to get current count: %v", err)
	}

	if count.Expiry == "" {
		t.Error("expected non-empty expiry timestamp")
	}

	// Parse the expiry to verify it's a valid RFC3339 timestamp
	_, err = time.Parse(time.RFC3339, count.Expiry)
	if err != nil {
		t.Errorf("expected valid RFC3339 timestamp, got: %v", err)
	}
}
