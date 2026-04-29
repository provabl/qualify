package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/provabl/qualify/internal/agent/lockfile"
	"github.com/provabl/qualify/internal/agent/store"
)

var (
	version   = "dev"
	commitSHA = "unknown"
	buildDate = "unknown"
)

const (
	defaultPort = "8737"
	defaultHost = "127.0.0.1" // localhost only for security
)

type server struct {
	store *store.Store
}

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("starting ark agent",
		"version", version,
		"commit", commitSHA,
		"buildDate", buildDate,
	)

	// Get data directory
	dataDir, err := getDataDir()
	if err != nil {
		slog.Error("failed to get data directory", "error", err)
		os.Exit(1)
	}

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		slog.Error("failed to create data directory", "error", err, "dir", dataDir)
		os.Exit(1)
	}

	// Acquire lock to ensure single instance
	lockPath := filepath.Join(dataDir, "agent.lock")
	lock := lockfile.New(lockPath)
	if err := lock.Acquire(); err != nil {
		slog.Error("failed to acquire lock", "error", err)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Another agent instance may already be running.\n")
		fmt.Fprintf(os.Stderr, "Use 'ark agent status' to check agent status.\n")
		os.Exit(1)
	}
	defer func() {
		if err := lock.Release(); err != nil {
			slog.Error("failed to release lock", "error", err)
		}
	}()

	slog.Info("lock acquired", "path", lockPath)

	// Open database
	dbPath := filepath.Join(dataDir, "agent.db")
	db, err := store.New(dbPath)
	if err != nil {
		slog.Error("failed to open database", "error", err, "path", dbPath)
		os.Exit(1)
	}
	defer db.Close()

	slog.Info("database opened", "path", dbPath)

	// Create server
	srv := &server{store: db}
	addr := fmt.Sprintf("%s:%s", defaultHost, getEnv("AGENT_PORT", defaultPort))
	httpSrv := &http.Server{
		Addr:         addr,
		Handler:      srv.setupRouter(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("agent listening", "addr", addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for interrupt signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	case sig := <-quit:
		slog.Info("shutdown signal received", "signal", sig.String())
	}

	// Graceful shutdown
	slog.Info("shutting down agent")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "error", err)
		os.Exit(1)
	}

	slog.Info("agent stopped")
}

func (s *server) setupRouter() http.Handler {
	r := chi.NewRouter()

	// Middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(loggerMiddleware)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/ping"))

	// CORS configuration (localhost only)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// API routes
	r.Route("/api", func(r chi.Router) {
		// System endpoints
		r.Route("/system", func(r chi.Router) {
			r.Get("/health", s.handleHealth)
			r.Get("/version", s.handleVersion)
		})

		// Credentials management
		r.Route("/credentials", func(r chi.Router) {
			r.Post("/", s.handleSetCredentials)
			r.Get("/", s.handleListCredentials)
			r.Delete("/{profile}", s.handleDeleteCredentials)
		})

		// S3 operations
		r.Route("/s3", func(r chi.Router) {
			r.Post("/buckets", s.handleCreateBucket)
		})

		// Agent configuration endpoints (future)
		// r.Route("/config", func(r chi.Router) {
		//     r.Get("/", s.handleGetConfig)
		//     r.Put("/", s.handleSetConfig)
		// })
	})

	return r
}

// loggerMiddleware logs HTTP requests with structured logging
func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		slog.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"bytes", ww.BytesWritten(),
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", middleware.GetReqID(r.Context()),
			"remote_addr", r.RemoteAddr,
		)
	})
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"status":  "healthy",
		"version": version,
		"time":    time.Now().UTC().Format(time.RFC3339),
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *server) handleVersion(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{
		"version":   version,
		"commit":    commitSHA,
		"buildDate": buildDate,
	}
	writeJSON(w, http.StatusOK, resp)
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode json response", "error", err)
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getDataDir returns the agent data directory
func getDataDir() (string, error) {
	// Check environment variable first
	if dir := os.Getenv("ARK_AGENT_DATA"); dir != "" {
		return dir, nil
	}

	// Use ~/.ark for data storage
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}

	return filepath.Join(home, ".ark"), nil
}
