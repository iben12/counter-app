package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/iben12/counter-app/internal/db"
	"github.com/iben12/counter-app/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// setupTestDB sets up a test database pool and runs migrations.
func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	// Use test database URL from environment or default
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/counter_test?sslmode=disable"
	}

	// Create pool
	pool, err := db.NewPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("failed to create test pool: %v", err)
	}

	// Run migrations
	if err := db.Migrate(ctx, databaseURL); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Cleanup function
	cleanup := func() {
		pool.Close()
	}

	return pool, cleanup
}

// TestHealthCheck tests the /health endpoint.
func TestHealthCheck(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	router := NewRouter(pool)
	req, _ := http.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", resp["status"])
	}
}

// TestCreateCounter tests creating a counter.
func TestCreateCounter(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	router := NewRouter(pool)

	body := []byte(`{"name":"test-counter","frequency":"1d"}`)
	req, _ := http.NewRequest("POST", "/counters", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var counter models.Counter
	if err := json.Unmarshal(rec.Body.Bytes(), &counter); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if counter.Name != "test-counter" {
		t.Errorf("expected name 'test-counter', got %q", counter.Name)
	}
	if counter.Frequency != "1d" {
		t.Errorf("expected frequency '1d', got %q", counter.Frequency)
	}
}

// TestCreateCounterDefaultFrequency tests creating a counter without frequency (should default to "1d").
func TestCreateCounterDefaultFrequency(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	router := NewRouter(pool)

	body := []byte(`{"name":"test-counter-default"}`)
	req, _ := http.NewRequest("POST", "/counters", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var counter models.Counter
	if err := json.Unmarshal(rec.Body.Bytes(), &counter); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if counter.Frequency != "1d" {
		t.Errorf("expected frequency '1d', got %q", counter.Frequency)
	}
}

// TestGetCounter tests retrieving a counter by ID.
func TestGetCounter(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a counter
	counter, err := models.CreateCounter(ctx, pool, "test-get", "2d", "UTC")
	if err != nil {
		t.Fatalf("failed to create counter: %v", err)
	}

	router := NewRouter(pool)
	req, _ := http.NewRequest("GET", fmt.Sprintf("/counters/%d", counter.ID), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var retrieved models.Counter
	if err := json.Unmarshal(rec.Body.Bytes(), &retrieved); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if retrieved.ID != counter.ID {
		t.Errorf("expected ID %d, got %d", counter.ID, retrieved.ID)
	}
	if retrieved.Frequency != "2d" {
		t.Errorf("expected frequency '2d', got %q", retrieved.Frequency)
	}
}

// TestListCounters tests retrieving all counters.
func TestListCounters(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a few counters
	models.CreateCounter(ctx, pool, "counter1", "1h", "UTC")
	models.CreateCounter(ctx, pool, "counter2", "1d", "UTC")

	req, _ := http.NewRequest("GET", "/counters", nil)
	router := NewRouter(pool)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var counters []models.Counter
	if err := json.Unmarshal(rec.Body.Bytes(), &counters); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(counters) < 2 {
		t.Errorf("expected at least 2 counters, got %d", len(counters))
	}
}

// TestUpdateCounterFrequency tests updating a counter's frequency.
func TestUpdateCounterFrequency(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a counter
	counter, err := models.CreateCounter(ctx, pool, "test-update", "1d", "UTC")
	if err != nil {
		t.Fatalf("failed to create counter: %v", err)
	}

	router := NewRouter(pool)
	body := []byte(`{"frequency":"3h"}`)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/counters/%d/frequency", counter.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var updated models.Counter
	if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if updated.Frequency != "3h" {
		t.Errorf("expected frequency '3h', got %q", updated.Frequency)
	}
}

// TestGetCurrentCount tests retrieving or creating the current count.
func TestGetCurrentCount(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a counter
	counter, err := models.CreateCounter(ctx, pool, "test-current-count", "1d", "UTC")
	if err != nil {
		t.Fatalf("failed to create counter: %v", err)
	}

	router := NewRouter(pool)
	req, _ := http.NewRequest("GET", fmt.Sprintf("/counters/%d/count", counter.ID), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var count models.Count
	if err := json.Unmarshal(rec.Body.Bytes(), &count); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if count.CounterID != counter.ID {
		t.Errorf("expected counter_id %d, got %d", counter.ID, count.CounterID)
	}
	if count.Value != 0 {
		t.Errorf("expected value 0, got %d", count.Value)
	}
	if count.Expiry == "" {
		t.Error("expected non-empty expiry")
	}
}

// TestIncrementCount tests incrementing a count.
func TestIncrementCount(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a counter and get initial count
	counter, err := models.CreateCounter(ctx, pool, "test-increment", "1d", "UTC")
	if err != nil {
		t.Fatalf("failed to create counter: %v", err)
	}

	models.GetOrCreateCurrentCount(ctx, pool, counter.ID)

	router := NewRouter(pool)

	// Increment by 5
	body := []byte(`{"delta":5}`)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/counters/%d/count/increment", counter.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var count models.Count
	if err := json.Unmarshal(rec.Body.Bytes(), &count); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if count.Value != 5 {
		t.Errorf("expected value 5, got %d", count.Value)
	}

	// Increment again by 3 (should be 8 total)
	body = []byte(`{"delta":3}`)
	req, _ = http.NewRequest("POST", fmt.Sprintf("/counters/%d/count/increment", counter.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if err := json.Unmarshal(rec.Body.Bytes(), &count); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if count.Value != 8 {
		t.Errorf("expected value 8, got %d", count.Value)
	}
}

// TestDecrementCount tests decrementing a count.
func TestDecrementCount(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a counter and get initial count
	counter, err := models.CreateCounter(ctx, pool, "test-decrement", "1d", "UTC")
	if err != nil {
		t.Fatalf("failed to create counter: %v", err)
	}

	models.GetOrCreateCurrentCount(ctx, pool, counter.ID)

	router := NewRouter(pool)

	// First increment to 10
	body := []byte(`{"delta":10}`)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/counters/%d/count/increment", counter.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Then decrement by 3 (should be 7)
	body = []byte(`{"delta":3}`)
	req, _ = http.NewRequest("POST", fmt.Sprintf("/counters/%d/count/decrement", counter.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var count models.Count
	if err := json.Unmarshal(rec.Body.Bytes(), &count); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if count.Value != 7 {
		t.Errorf("expected value 7, got %d", count.Value)
	}
}

// TestGetCountHistory tests retrieving count history.
func TestGetCountHistory(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a counter
	counter, err := models.CreateCounter(ctx, pool, "test-history", "1h", "UTC")
	if err != nil {
		t.Fatalf("failed to create counter: %v", err)
	}

	// Get and manipulate counts
	count1, _ := models.GetOrCreateCurrentCount(ctx, pool, counter.ID)
	models.IncrementCurrentCount(ctx, pool, counter.ID, 5)

	// Manually create a historical count (simulating expired count)
	//   This tests the history endpoint
	models.IncrementCurrentCount(ctx, pool, counter.ID, 2) // now at 7

	router := NewRouter(pool)
	req, _ := http.NewRequest("GET", fmt.Sprintf("/counters/%d/counts", counter.ID), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var counts []models.Count
	if err := json.Unmarshal(rec.Body.Bytes(), &counts); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(counts) < 1 {
		t.Errorf("expected at least 1 count in history, got %d", len(counts))
	}

	if counts[0].ID != count1.ID {
		t.Errorf("expected first count ID %d, got %d", count1.ID, counts[0].ID)
	}
}

// TestCountExpiryReset tests that an expired count is reset (new count created).
func TestCountExpiryReset(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a counter with 1h frequency
	counter, err := models.CreateCounter(ctx, pool, "test-expiry", "1h", "UTC")
	if err != nil {
		t.Fatalf("failed to create counter: %v", err)
	}

	// Get initial count
	count1, _ := models.GetOrCreateCurrentCount(ctx, pool, counter.ID)
	originalID := count1.ID

	// Manually set its expiry to the past
	updateExpirySQL := `UPDATE counts SET expiry = now() - interval '1 hour' WHERE id = $1`
	_, err = pool.Exec(ctx, updateExpirySQL, originalID)
	if err != nil {
		t.Fatalf("failed to update expiry: %v", err)
	}

	// Now get current count again - should be a new count
	time.Sleep(100 * time.Millisecond) // Small delay to ensure different timestamps
	count2, _ := models.GetOrCreateCurrentCount(ctx, pool, counter.ID)

	if count2.ID == originalID {
		t.Errorf("expected new count ID, but got same ID %d", originalID)
	}

	if count2.Value != 0 {
		t.Errorf("expected new count value 0, got %d", count2.Value)
	}
}

// TestIncrementDecrementSameRow tests that increment and decrement update the same current count row.
func TestIncrementDecrementSameRow(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a counter
	counter, err := models.CreateCounter(ctx, pool, "test-same-row", "1d", "UTC")
	if err != nil {
		t.Fatalf("failed to create counter: %v", err)
	}

	// Get initial count
	initialCount, _ := models.GetOrCreateCurrentCount(ctx, pool, counter.ID)
	initialID := initialCount.ID

	// Increment
	models.IncrementCurrentCount(ctx, pool, counter.ID, 10)

	// Decrement
	models.IncrementCurrentCount(ctx, pool, counter.ID, -5) // Use negative for decrement via API

	// Get history - should still have only 1 count row
	history, _ := models.GetCountHistory(ctx, pool, counter.ID)

	if len(history) != 1 {
		t.Errorf("expected 1 count in history, got %d", len(history))
	}

	if history[0].ID != initialID {
		t.Errorf("expected count ID %d, got %d", initialID, history[0].ID)
	}

	if history[0].Value != 5 {
		t.Errorf("expected count value 5 (10-5), got %d", history[0].Value)
	}
}
