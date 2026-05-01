// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func devConfig() Config {
	return Config{
		DevMode:     true,
		DevUserID:   "test-user-uuid",
		DevEmail:    "dev@example.edu",
		DevRole:     "researcher",
		TokenExpiry: time.Hour,
	}
}

func prodConfig() Config {
	return Config{
		DevMode:     false,
		JWTSecret:   []byte("test-secret-that-is-long-enough-32b"),
		TokenExpiry: time.Hour,
	}
}

// ── Config ─────────────────────────────────────────────────────────────────

func TestConfig_ProductionReady(t *testing.T) {
	if devConfig().ProductionReady() {
		t.Error("dev config should not be production-ready")
	}
	if prodConfig().ProductionReady() == false {
		t.Error("prod config with 32-byte secret should be production-ready")
	}
	short := Config{JWTSecret: []byte("too-short")}
	if short.ProductionReady() {
		t.Error("config with short secret should not be production-ready")
	}
}

// ── JWT ────────────────────────────────────────────────────────────────────

func TestIssueAndValidateToken(t *testing.T) {
	cfg := prodConfig()
	tok, err := IssueToken(cfg, "user-1", "alice@mru.edu", "MRU", "researcher")
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}
	if tok == "" {
		t.Fatal("empty token")
	}

	claims, err := ValidateToken(cfg, tok)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.UserID() != "user-1" {
		t.Errorf("UserID: got %q, want user-1", claims.UserID())
	}
	if claims.Email != "alice@mru.edu" {
		t.Errorf("Email: got %q", claims.Email)
	}
	if claims.Role != "researcher" {
		t.Errorf("Role: got %q", claims.Role)
	}
}

func TestValidateToken_Expired(t *testing.T) {
	cfg := Config{JWTSecret: []byte("test-secret-that-is-long-enough-32b"), TokenExpiry: -time.Minute}
	tok, _ := IssueToken(cfg, "u", "e@x.com", "", "researcher")
	_, err := ValidateToken(cfg, tok)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	cfg1 := prodConfig()
	cfg2 := Config{JWTSecret: []byte("different-secret-that-is-also-32b"), TokenExpiry: time.Hour}
	tok, _ := IssueToken(cfg1, "u", "e@x.com", "", "researcher")
	_, err := ValidateToken(cfg2, tok)
	if err == nil {
		t.Error("expected error for wrong secret")
	}
}

func TestValidateToken_Garbage(t *testing.T) {
	_, err := ValidateToken(prodConfig(), "not.a.jwt")
	if err == nil {
		t.Error("expected error for garbage token")
	}
}

// ── Context ────────────────────────────────────────────────────────────────

func TestContext_RoundTrip(t *testing.T) {
	cfg := prodConfig()
	tok, _ := IssueToken(cfg, "ctx-user", "ctx@x.com", "Inst", "admin")
	claims, _ := ValidateToken(cfg, tok)

	req := httptest.NewRequest("GET", "/", nil)
	ctx := WithClaims(req.Context(), claims)

	if GetUserID(ctx) != "ctx-user" {
		t.Errorf("GetUserID: got %q", GetUserID(ctx))
	}
	if GetRole(ctx) != "admin" {
		t.Errorf("GetRole: got %q", GetRole(ctx))
	}
}

func TestContext_Empty(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	if GetUserID(req.Context()) != "" {
		t.Error("expected empty UserID from unauthenticated context")
	}
}

// ── Middleware ─────────────────────────────────────────────────────────────

func TestMiddleware_DevMode(t *testing.T) {
	handler := Middleware(devConfig())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := GetUserID(r.Context())
		if uid != "test-user-uuid" {
			t.Errorf("dev mode injected wrong user: %q", uid)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestMiddleware_ValidToken(t *testing.T) {
	cfg := prodConfig()
	tok, _ := IssueToken(cfg, "prod-user", "prod@x.com", "", "researcher")

	handler := Middleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if GetUserID(r.Context()) != "prod-user" {
			t.Errorf("wrong user in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestMiddleware_MissingToken(t *testing.T) {
	handler := Middleware(prodConfig())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called without token")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestMiddleware_InvalidToken(t *testing.T) {
	handler := Middleware(prodConfig())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with invalid token")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer garbage.token.here")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestRequireRole(t *testing.T) {
	cfg := prodConfig()
	tok, _ := IssueToken(cfg, "u", "e@x.com", "", "researcher")

	// researcher calling admin-only route
	handler := Middleware(cfg)(RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler")
	})))

	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}
