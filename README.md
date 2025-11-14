
# counter-app (Go)

Lightweight Go API server with a Postgres backend for storing simple `Counter` records.

## Quick start (macOS / zsh)

1. Start Postgres locally with Docker Compose:

```bash
	docker-compose up -d
	```

2. (Optional) copy `.env.sample` to `.env` and edit if needed.

3. Run the server:

	```bash
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/counter?sslmode=disable
	go run ./cmd/server
	```

The server listens on :8080 by default. Endpoints:
# counter-app (Go)

Lightweight Go API server with a Postgres backend for storing simple `Counter` records.

## Quick start (macOS / zsh)

1. Start Postgres locally with Docker Compose:

```bash
docker-compose up -d
```

2. (Optional) copy `.env.sample` to `.env` and edit if needed.

3. Run the server:

```bash
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/counter?sslmode=disable
go run ./cmd/server
```

The server listens on :8080 by default. Endpoints:

- GET /health
- GET /counters
- POST /counters    {"name":"example"}
- GET /counters/{id}
- POST /counters/{id}/increment  {"delta": 1}

## Notes for contributors

- The project uses `pgx` (pgxpool) for DB access and `gorilla/mux` for routing. Dependencies are in `go.mod`.
- Migrations are intentionally small and run at startup via `internal/db.Migrate` — suitable for development. For production, replace with a real migration tool.

## Development vs production compose

This repo includes two compose files:

- `docker-compose.yml` — developer-friendly: mounts source folders into the app container so you can edit code and re-run locally.
- `docker-compose.prod.yml` — image-first: uses the built image and does not mount source files (closer to what you'd run in production).

To run the production-style stack locally (builds the app image and runs the container):

```bash
docker compose -f docker-compose.prod.yml up --build -d
```

## Running tests

Unit tests (fast, no postgres required for frequency logic):

```bash
go test ./internal/db -v
```

Integration tests (use a test Postgres DB). The test suite uses `TEST_DATABASE_URL` environment variable. Example — create a local test DB and run the handler integration tests:

```bash
PGPASSWORD=postgres psql -U postgres -h localhost -c "CREATE DATABASE counter_test;" || true
TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/counter_test?sslmode=disable" go test ./internal/handlers -v
```

## CI

There is a GitHub Actions workflow at `.github/workflows/ci.yml` which runs unit tests first, then runs integration tests in a job that starts a Postgres service and runs handler integration tests against it.

## Notes

- Because migrations are embedded in the binary, the app runs migrations at startup and does not require migration files to be present at runtime.
- For development, `docker-compose.yml` still mounts source dirs (cmd/internal) for quick edit/run cycles. For production runs, use `docker-compose.prod.yml`.
