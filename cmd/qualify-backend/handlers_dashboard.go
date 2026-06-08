// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log/slog"
	"net/http"

	"github.com/provabl/qualify/internal/auth"
	"github.com/provabl/qualify/internal/training"
)

// handleGetDashboardStats retrieves dashboard statistics for the authenticated user.
func handleGetDashboardStats(trainingSvc *training.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.GetUserID(r.Context())
		if userID == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
			return
		}

		stats, err := trainingSvc.GetDashboardStats(r.Context(), userID)
		if err != nil {
			slog.Error("failed to get dashboard stats", "error", err, "user_id", userID)
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Failed to retrieve dashboard statistics",
			})
			return
		}

		writeJSON(w, http.StatusOK, stats)
	}
}
