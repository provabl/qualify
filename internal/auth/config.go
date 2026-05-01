// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"os"
	"time"
)

// Config holds authentication configuration for the qualify backend.
// In dev mode the backend issues tokens without requiring an external IdP —
// suitable for local development. In production JWT_SECRET must be set.
//
// Environment variables:
//
//	AUTH_DEV_MODE=true         bypass JWT validation; inject DEV_USER_ID into context
//	AUTH_DEV_USER_ID=<uuid>    user ID injected when dev mode is active
//	AUTH_DEV_EMAIL=<email>     email injected when dev mode is active
//	AUTH_DEV_ROLE=researcher   role injected when dev mode is active (default: researcher)
//	JWT_SECRET=<32+ bytes>     HMAC-SHA256 signing secret; required in production
//	JWT_EXPIRY=24h             token lifetime (default: 24h)
type Config struct {
	DevMode     bool
	DevUserID   string
	DevEmail    string
	DevRole     string
	JWTSecret   []byte
	TokenExpiry time.Duration
}

// Load reads auth configuration from environment variables.
func Load() Config {
	devMode := os.Getenv("AUTH_DEV_MODE") == "true"
	devUserID := getEnv("AUTH_DEV_USER_ID", "00000000-0000-0000-0000-000000000001")
	devEmail := getEnv("AUTH_DEV_EMAIL", "dev@example.edu")
	devRole := getEnv("AUTH_DEV_ROLE", "researcher")
	secret := []byte(os.Getenv("JWT_SECRET"))
	expiry := parseDuration(os.Getenv("JWT_EXPIRY"), 24*time.Hour)

	return Config{
		DevMode:     devMode,
		DevUserID:   devUserID,
		DevEmail:    devEmail,
		DevRole:     devRole,
		JWTSecret:   secret,
		TokenExpiry: expiry,
	}
}

// ProductionReady returns true when the config is suitable for non-dev deployments.
// Requires JWT_SECRET to be set to at least 32 bytes.
func (c Config) ProductionReady() bool {
	return !c.DevMode && len(c.JWTSecret) >= 32
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseDuration(s string, fallback time.Duration) time.Duration {
	if s == "" {
		return fallback
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return fallback
	}
	return d
}
