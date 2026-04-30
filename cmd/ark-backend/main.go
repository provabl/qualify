// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/provabl/qualify/internal/audit"
	"github.com/provabl/qualify/internal/database"
	"github.com/provabl/qualify/internal/training"
)

var (
	version   = "dev"
	commitSHA = "unknown"
	buildDate = "unknown"
)

const (
	defaultPort = "8080"
	defaultHost = "0.0.0.0"
)

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("starting ark backend",
		"version", version,
		"commit", commitSHA,
		"buildDate", buildDate,
	)

	// Initialize database
	dbCfg := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "ark"),
		Password: getEnv("DB_PASSWORD", "ark_dev_password"),
		DBName:   getEnv("DB_NAME", "ark"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	db, err := database.New(dbCfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	slog.Info("database connected",
		"host", dbCfg.Host,
		"port", dbCfg.Port,
		"dbname", dbCfg.DBName,
	)

	// Run migrations
	migrationsPath := getEnv("MIGRATIONS_PATH", "./migrations")
	if err := db.RunMigrations(migrationsPath); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	version, dirty, err := db.MigrationVersion(migrationsPath)
	if err != nil {
		slog.Warn("could not get migration version", "error", err)
	} else {
		slog.Info("migrations completed",
			"version", version,
			"dirty", dirty,
		)
	}

	// Initialize services
	auditSvc := audit.NewService(db)
	trainingSvc := training.NewService(db)

	slog.Info("services initialized")

	// Create server
	addr := fmt.Sprintf("%s:%s", defaultHost, getEnv("PORT", defaultPort))
	srv := &http.Server{
		Addr:         addr,
		Handler:      setupRouter(auditSvc, trainingSvc),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("backend listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
	slog.Info("shutting down backend")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "error", err)
		os.Exit(1)
	}

	slog.Info("backend stopped")
}

func setupRouter(auditSvc *audit.Service, trainingSvc *training.Service) http.Handler {
	r := chi.NewRouter()

	// Middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(loggerMiddleware)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/ping"))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health and system endpoints
	r.Get("/health", handleHealth)

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/version", handleVersion)

		// System endpoints
		r.Route("/system", func(r chi.Router) {
			r.Get("/health", handleHealth)
			r.Get("/version", handleVersion)
		})

		// Audit endpoints
		r.Route("/audit", func(r chi.Router) {
			r.Post("/log", handleLogAudit(auditSvc))
			r.Get("/logs", handleQueryAudit(auditSvc))
		})

		// Policy and training endpoints
		r.Route("/policies", func(r chi.Router) {
			r.Post("/check", handleCheckPolicy(trainingSvc))
		})

		r.Route("/training", func(r chi.Router) {
			r.Get("/progress/{user_id}", handleGetUserProgress(trainingSvc))
			r.Get("/modules", handleListModules(trainingSvc))
			r.Get("/modules/{id}", handleGetModule(trainingSvc))
			r.Post("/modules/{id}/start", handleStartModule(trainingSvc))
			r.Post("/modules/{id}/complete", handleCompleteModule(trainingSvc))
			r.Post("/modules/{id}/quiz/submit", handleSubmitQuiz(trainingSvc))
			r.Get("/activity/{user_id}", handleGetUserActivity(trainingSvc))
		})

		r.Route("/dashboard", func(r chi.Router) {
			r.Get("/stats/{user_id}", handleGetDashboardStats(trainingSvc))
		})

		r.Route("/users", func(r chi.Router) {
			r.Get("/{user_id}/profile", handleGetUserProfile(trainingSvc))
			r.Put("/{user_id}/profile", handleUpdateUserProfile(trainingSvc))
		})

		// Future endpoints
		// r.Route("/auth", func(r chi.Router) { ... })
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

func handleHealth(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"status":  "healthy",
		"version": version,
		"time":    time.Now().UTC().Format(time.RFC3339),
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleVersion(w http.ResponseWriter, r *http.Request) {
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

// getEnvInt gets an integer environment variable or returns a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intVal int
		if _, err := fmt.Sscanf(value, "%d", &intVal); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// parseIntQueryParam parses an integer query parameter
func parseIntQueryParam(value string) (int, error) {
	var intVal int
	if _, err := fmt.Sscanf(value, "%d", &intVal); err != nil {
		return 0, err
	}
	return intVal, nil
}
