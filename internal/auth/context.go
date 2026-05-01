// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package auth

import "context"

type contextKey string

const claimsKey contextKey = "auth_claims"

// WithClaims returns a new context with the given claims attached.
func WithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

// GetClaims retrieves auth claims from the context.
// Returns nil if no claims are present (unauthenticated context).
func GetClaims(ctx context.Context) *Claims {
	v, _ := ctx.Value(claimsKey).(*Claims)
	return v
}

// GetUserID retrieves the authenticated user's UUID from the context.
// Returns "" if the context is unauthenticated.
func GetUserID(ctx context.Context) string {
	if c := GetClaims(ctx); c != nil {
		return c.UserID()
	}
	return ""
}

// GetRole retrieves the authenticated user's role from the context.
func GetRole(ctx context.Context) string {
	if c := GetClaims(ctx); c != nil {
		return c.Role
	}
	return ""
}
