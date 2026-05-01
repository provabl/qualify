// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/provabl/qualify/internal/auth"
	"github.com/provabl/qualify/internal/training"
)

// handleGetUserProfile returns the authenticated user's profile.
// Route is /api/users/me — user identity comes from the JWT, not a URL param.
func handleGetUserProfile(trainingSvc *training.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.GetClaims(r.Context())
		if claims == nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
			return
		}
		userID := claims.UserID()

		profile, err := trainingSvc.GetUserProfile(r.Context(), userID)
		if err != nil {
			// Profile not found — build one from JWT claims and return it
			profile = &training.UserProfile{
				UserID:      userID,
				Email:       claims.Email,
				Institution: claims.Institution,
				Role:        claims.Role,
				Preferences: training.UserPreferences{
					HasCompletedOnboarding: false,
					ShowTrainingReminders:  true,
					DefaultAWSRegion:       "us-east-1",
				},
				CreatedAt: time.Now().UTC().Format(time.RFC3339),
			}
		}

		slog.Info("user profile retrieved", "user_id", userID)
		writeJSON(w, http.StatusOK, profile)
	}
}

// handleUpdateUserProfile updates the authenticated user's profile.
func handleUpdateUserProfile(trainingSvc *training.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.GetClaims(r.Context())
		if claims == nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
			return
		}
		userID := claims.UserID()

		var updates struct {
			Email       *string                   `json:"email,omitempty"`
			Name        *string                   `json:"name,omitempty"`
			Institution *string                   `json:"institution,omitempty"`
			Preferences *training.UserPreferences `json:"preferences,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}

		if err := trainingSvc.UpdateUserProfile(r.Context(), userID, updates.Email, updates.Name, updates.Institution, updates.Preferences); err != nil {
			slog.Error("failed to update user profile", "user_id", userID, "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update profile"})
			return
		}

		// Return updated profile
		profile, err := trainingSvc.GetUserProfile(r.Context(), userID)
		if err != nil {
			profile = &training.UserProfile{UserID: userID, Email: claims.Email, Role: claims.Role}
		}

		slog.Info("user profile updated", "user_id", userID)
		writeJSON(w, http.StatusOK, profile)
	}
}
