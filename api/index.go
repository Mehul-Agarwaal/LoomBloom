package api

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"loombloom/internal/config"
	"loombloom/internal/httpapi"
	"loombloom/internal/store"
)

var (
	appHandler http.Handler
	db         *sql.DB
)

func init() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		return
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel(),
	}))
	slog.SetDefault(logger)

	// Connect to database
	db, err = sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		logger.Error("open database", "error", err)
		return
	}

	// Important for serverless: keep max connections low so you don't exhaust the database
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(1)

	// Initialize the application router
	api := httpapi.New(store.New(db), logger)
	appHandler = api.Routes()
}

// Handler is the Vercel serverless entrypoint
func Handler(w http.ResponseWriter, r *http.Request) {
	if appHandler == nil {
		http.Error(w, "Application failed to initialize", http.StatusInternalServerError)
		return
	}
	appHandler.ServeHTTP(w, r)
}
