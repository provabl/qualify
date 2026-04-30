// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/provabl/qualify/internal/training"
)

// handleGetUserProfile retrieves a user profile
func handleGetUserProfile(trainingSvc *training.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "user_id")
		if userID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "user_id parameter required",
			})
			return
		}

		// TODO: Implement actual user profile retrieval from database
		// For now, return a mock profile with default preferences
		profile := training.UserProfile{
			UserID:      userID,
			Email:       "user@example.com",
			Name:        "Research User",
			Institution: "University",
			Role:        "researcher",
			Preferences: training.UserPreferences{
				HasCompletedOnboarding: false,
				ShowTrainingReminders:  true,
				DefaultAWSRegion:       "us-east-1",
			},
			CreatedAt: "2025-01-01T00:00:00Z",
		}

		slog.Info("user profile retrieved",
			"user_id", userID,
		)

		writeJSON(w, http.StatusOK, profile)
	}
}

// handleUpdateUserProfile updates a user profile
func handleUpdateUserProfile(trainingSvc *training.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "user_id")
		if userID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "user_id parameter required",
			})
			return
		}

		var updates struct {
			Email       *string                   `json:"email,omitempty"`
			Name        *string                   `json:"name,omitempty"`
			Institution *string                   `json:"institution,omitempty"`
			Preferences *training.UserPreferences `json:"preferences,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			slog.Error("failed to decode profile update request", "error", err)
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid request body",
			})
			return
		}

		// TODO: Implement actual user profile update in database
		// For now, return the updated profile
		profile := training.UserProfile{
			UserID:      userID,
			Email:       "user@example.com",
			Name:        "Research User",
			Institution: "University",
			Role:        "researcher",
			Preferences: training.UserPreferences{
				HasCompletedOnboarding: false,
				ShowTrainingReminders:  true,
				DefaultAWSRegion:       "us-east-1",
			},
			CreatedAt: "2025-01-01T00:00:00Z",
		}

		// Apply updates if provided
		if updates.Email != nil {
			profile.Email = *updates.Email
		}
		if updates.Name != nil {
			profile.Name = *updates.Name
		}
		if updates.Institution != nil {
			profile.Institution = *updates.Institution
		}
		if updates.Preferences != nil {
			profile.Preferences = *updates.Preferences
		}

		slog.Info("user profile updated",
			"user_id", userID,
		)

		writeJSON(w, http.StatusOK, profile)
	}
}
