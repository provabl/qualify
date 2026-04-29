package main

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/provabl/qualify/internal/training"
)

// handleGetDashboardStats retrieves dashboard statistics for a user
func handleGetDashboardStats(trainingSvc *training.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "user_id")
		if userID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "user_id parameter required",
			})
			return
		}

		stats, err := trainingSvc.GetDashboardStats(r.Context(), userID)
		if err != nil {
			slog.Error("failed to get dashboard stats",
				"error", err,
				"user_id", userID,
			)
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Failed to retrieve dashboard statistics",
			})
			return
		}

		writeJSON(w, http.StatusOK, stats)
	}
}
