package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"loombloom/internal/config"
)

func main() {
	reset := flag.Bool("reset", false, "Drop the public schema before running migrations")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel(),
	}))
	slog.SetDefault(logger)

	logger.Info("connecting to database", "url", cfg.DatabaseURL)
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		logger.Error("ping database failed", "error", err)
		os.Exit(1)
	}

	if *reset {
		logger.Info("resetting database schema")
		_, err := db.ExecContext(ctx, "DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
		if err != nil {
			logger.Error("failed to reset database schema", "error", err)
			os.Exit(1)
		}
		logger.Info("successfully reset database schema")
	}

	// Dynamically scan the migrations directory
	files, err := os.ReadDir("migrations")
	if err != nil {
		logger.Error("failed to read migrations directory", "error", err)
		os.Exit(1)
	}

	var migrationFiles []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		// Only run files ending with .sql and not ending with .down.sql
		if strings.HasSuffix(name, ".sql") && !strings.HasSuffix(name, ".down.sql") {
			migrationFiles = append(migrationFiles, filepath.Join("migrations", name))
		}
	}

	// Sort alphabetically so that they run in sequence (e.g. 001, 002, 003)
	sort.Strings(migrationFiles)

	logger.Info("found migration files", "files", migrationFiles)

	for _, file := range migrationFiles {
		logger.Info("running migration file", "file", file)
		content, err := os.ReadFile(file)
		if err != nil {
			logger.Error("failed to read migration file", "file", file, "error", err)
			os.Exit(1)
		}

		_, err = db.ExecContext(ctx, string(content))
		if err != nil {
			logger.Error("failed to execute migration", "file", file, "error", err)
			os.Exit(1)
		}
		logger.Info("successfully applied migration", "file", file)
	}

	fmt.Println("Database successfully migrated and seeded!")
}
