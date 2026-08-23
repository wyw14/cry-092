package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if len(os.Args) != 2 || (os.Args[1] != "up" && os.Args[1] != "seed") {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/migrate <up|seed>")
		os.Exit(2)
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://cry092:cry092@localhost:5432/cry092?sslmode=disable"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fail(err)
	}
	defer pool.Close()
	pattern := "migrations/*_init.up.sql"
	if os.Args[1] == "seed" {
		pattern = "migrations/*_seed.up.sql"
	}
	files, err := filepath.Glob(pattern)
	if err != nil {
		fail(err)
	}
	sort.Strings(files)
	for _, path := range files {
		body, err := os.ReadFile(path)
		if err != nil {
			fail(err)
		}
		if strings.TrimSpace(string(body)) == "" {
			continue
		}
		if _, err := pool.Exec(ctx, string(body)); err != nil {
			fail(fmt.Errorf("apply %s: %w", path, err))
		}
		fmt.Println("applied", path)
	}
}
func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
