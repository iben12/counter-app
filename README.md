# counter-app (Go)

Lightweight Go API server with a Postgres backend for storing simple `Counter` records.

Quick start (macOS / zsh)

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

Notes for contributors
- The project uses `pgx` (pgxpool) for DB access and `gorilla/mux` for routing. Dependencies are in `go.mod`.
- Migrations are intentionally small and run at startup via `internal/db.Migrate` — suitable for development. For production, replace with a real migration tool.
