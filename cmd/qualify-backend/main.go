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
	"github.com/provabl/qualify/internal/auth"
	"github.com/provabl/qualify/internal/database"
	"github.com/provabl/qualify/internal/license"
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
	ctx := context.Background()

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
		User:     getEnv("DB_USER", "qualify"),
		Password: getEnv("DB_PASSWORD", "qualify_dev_password"),
		DBName:   getEnv("DB_NAME", "qualify"),
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

	// Load auth configuration
	authCfg := auth.Load()
	if authCfg.DevMode {
		slog.Warn("AUTH_DEV_MODE is enabled — DO NOT USE IN PRODUCTION")
	} else if !authCfg.ProductionReady() {
		slog.Warn("JWT_SECRET not configured — tokens will use a dev secret; set JWT_SECRET for production")
	}

	// Validate license (network-based, cached in system_config)
	licenseKey := getEnv("LICENSE_KEY", "")
	licenseEndpoint := getEnv("LICENSE_ENDPOINT", "https://licensing.provabl.co/api/v1/validate")
	cacheTTL := getEnvDuration("LICENSE_CACHE_TTL", 24*time.Hour)
	licenseValidator := license.NewValidator(licenseEndpoint, licenseKey, cacheTTL, db)
	licenseInfo, err := licenseValidator.ValidateAndCache(ctx)
	if err != nil {
		slog.Warn("license validation failed — running as community tier", "error", err)
		licenseInfo = license.CommunityLicense()
	}
	slog.Info("license active", "tier", licenseInfo.Tier, "expires", licenseInfo.ExpiresAt)

	// Initialize services
	auditSvc := audit.NewService(db)
	trainingSvc := training.NewService(db)

	slog.Info("services initialized")

	// Create server
	addr := fmt.Sprintf("%s:%s", defaultHost, getEnv("PORT", defaultPort))
	srv := &http.Server{
		Addr:         addr,
		Handler:      setupRouter(authCfg, auditSvc, trainingSvc),
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

func setupRouter(authCfg auth.Config, auditSvc *audit.Service, trainingSvc *training.Service) http.Handler {
	r := chi.NewRouter()

	// Middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(loggerMiddleware)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/ping"))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		// Restrict to known frontend ports — no wildcard port matching.
		// Note: Go HTTP clients (qualify agent on :8737) bypass CORS entirely —
		// CORS is enforced only by browsers. This list gates web dashboard access only.
		AllowedOrigins: []string{
			"http://localhost:5173", // Vite dev server
			"http://localhost:5174", // Playwright test server
			"http://127.0.0.1:5173",
			"http://127.0.0.1:5174",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public endpoints — no authentication required
	r.Get("/health", handleHealth)

	r.Route("/api", func(r chi.Router) {
		// Public: version, auth, public training listings
		r.Get("/version", handleVersion)
		r.Route("/system", func(r chi.Router) {
			r.Get("/health", handleHealth)
			r.Get("/version", handleVersion)
		})

		// Auth endpoints — public (issue tokens, return current user)
		r.Route("/auth", func(r chi.Router) {
			r.Get("/me", handleAuthMe(authCfg))
			if authCfg.DevMode {
				// Only available in dev mode — returns a signed JWT for the dev user
				r.Get("/dev-token", handleDevToken(authCfg))
			}
		})

		// Public training content (no personal data)
		r.Route("/training/modules", func(r chi.Router) {
			r.Get("/", handleListModules(trainingSvc))
			r.Get("/{id}", handleGetModule(trainingSvc))
		})

		// Protected endpoints — require valid JWT
		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware(authCfg))

			// Audit
			r.Route("/audit", func(r chi.Router) {
				r.Post("/log", handleLogAudit(auditSvc))
				r.Get("/logs", handleQueryAudit(auditSvc))
			})

			// Policy evaluation
			r.Route("/policies", func(r chi.Router) {
				r.Post("/check", handleCheckPolicy(trainingSvc))
			})

			// Training (personal data — user from context)
			r.Route("/training", func(r chi.Router) {
				r.Get("/progress", handleGetUserProgress(trainingSvc))
				r.Get("/activity", handleGetUserActivity(trainingSvc))
				r.Post("/modules/{id}/start", handleStartModule(trainingSvc))
				r.Post("/modules/{id}/complete", handleCompleteModule(trainingSvc))
				r.Post("/modules/{id}/quiz/submit", handleSubmitQuiz(trainingSvc))
			})

			// Dashboard
			r.Get("/dashboard/stats", handleGetDashboardStats(trainingSvc))

			// User profile
			r.Route("/users/me", func(r chi.Router) {
				r.Get("/", handleGetUserProfile(trainingSvc))
				r.Put("/", handleUpdateUserProfile(trainingSvc))
			})
		})
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

// handleAuthMe returns the authenticated user's claims from their token.
func handleAuthMe(cfg auth.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Apply middleware inline so /api/auth/me returns proper 401 when unauthenticated.
		var claims *auth.Claims
		if cfg.DevMode {
			// Dev mode: synthesize claims from config
			c := &auth.Claims{Email: cfg.DevEmail, Role: cfg.DevRole}
			c.Subject = cfg.DevUserID
			claims = c
		} else {
			token := r.Header.Get("Authorization")
			if len(token) > 7 {
				token = token[7:] // strip "Bearer "
			}
			var err error
			claims, err = auth.ValidateToken(cfg, token)
			if err != nil {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}
		}
		writeJSON(w, http.StatusOK, map[string]string{
			"user_id":     claims.UserID(),
			"email":       claims.Email,
			"institution": claims.Institution,
			"role":        claims.Role,
		})
	}
}

// handleDevToken issues a JWT for the dev user. Only registered when AUTH_DEV_MODE=true.
func handleDevToken(cfg auth.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok, err := auth.IssueToken(cfg, cfg.DevUserID, cfg.DevEmail, "", cfg.DevRole)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{
			"token":   tok,
			"user_id": cfg.DevUserID,
			"email":   cfg.DevEmail,
			"role":    cfg.DevRole,
		})
	}
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

// getEnvDuration gets a duration environment variable or returns a default value
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
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
