package main

import (
    "context"
    "log"
    "net/http"
    "os"

    "github.com/iben12/counter-app/internal/db"
    "github.com/iben12/counter-app/internal/handlers"
)

func main() {
    ctx := context.Background()

    databaseURL := os.Getenv("DATABASE_URL")
    if databaseURL == "" {
        databaseURL = "postgres://postgres:postgres@localhost:5432/counter?sslmode=disable"
    }

    pool, err := db.NewPool(ctx, databaseURL)
    if err != nil {
        log.Fatalf("failed to connect to db: %v", err)
    }
    defer pool.Close()

    // Run simple migration (create table if not exists)
    if err := db.Migrate(ctx, pool); err != nil {
        log.Fatalf("migration failed: %v", err)
    }

    r := handlers.NewRouter(pool)

    addr := os.Getenv("ADDR")
    if addr == "" {
        addr = ":8080"
    }

    log.Printf("starting server on %s", addr)
    if err := http.ListenAndServe(addr, r); err != nil {
        log.Fatalf("server error: %v", err)
    }
}
