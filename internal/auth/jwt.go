// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims holds the payload of a qualify JWT.
// sub = user UUID, email from institution directory, role from qualify.
type Claims struct {
	Email       string `json:"email"`
	Institution string `json:"institution,omitempty"`
	Role        string `json:"role"` // researcher | admin | instructor
	jwt.RegisteredClaims
}

// UserID returns the subject claim (user UUID).
func (c *Claims) UserID() string {
	return c.Subject
}

// IssueToken creates and signs a JWT for the given user.
func IssueToken(cfg Config, userID, email, institution, role string) (string, error) {
	if len(cfg.JWTSecret) == 0 && !cfg.DevMode {
		return "", fmt.Errorf("JWT_SECRET is not configured")
	}
	now := time.Now()
	claims := &Claims{
		Email:       email,
		Institution: institution,
		Role:        role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.TokenExpiry)),
			Issuer:    "qualify",
			Audience:  jwt.ClaimStrings{"qualify-api"},
		},
	}
	secret := cfg.JWTSecret
	if len(secret) == 0 {
		// Dev mode: use a deterministic dev secret so tokens survive restarts.
		secret = []byte("qualify-dev-secret-not-for-production")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ValidateToken parses and validates a Bearer token string.
func ValidateToken(cfg Config, tokenStr string) (*Claims, error) {
	secret := cfg.JWTSecret
	if len(secret) == 0 {
		secret = []byte("qualify-dev-secret-not-for-production")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuer("qualify"),
		jwt.WithAudience("qualify-api"),
		jwt.WithExpirationRequired())

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	return claims, nil
}
