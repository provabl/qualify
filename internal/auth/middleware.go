// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"net/http"
	"strings"
)

// Middleware returns a chi-compatible middleware that:
//   - In dev mode: injects a dev user into the context without validating any token.
//   - In production mode: requires a valid Bearer JWT in the Authorization header.
//
// Public paths (health, version, module listing, auth endpoints) bypass this
// middleware; they must be registered outside the protected route group.
func Middleware(cfg Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.DevMode {
				claims := &Claims{
					Email: cfg.DevEmail,
					Role:  cfg.DevRole,
				}
				claims.Subject = cfg.DevUserID
				next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
				return
			}

			// Extract Bearer token.
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, `{"error":"missing Authorization header"}`, http.StatusUnauthorized)
				return
			}
			tokenStr := strings.TrimPrefix(header, "Bearer ")

			claims, err := ValidateToken(cfg, tokenStr)
			if err != nil {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
		})
	}
}

// RequireRole returns a middleware that enforces the caller has one of the
// allowed roles. Must be used inside the auth Middleware group.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := GetRole(r.Context())
			if !allowed[role] {
				http.Error(w, `{"error":"insufficient permissions"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
