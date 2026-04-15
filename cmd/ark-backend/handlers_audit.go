package main

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/provabl/ark/internal/audit"
)

// handleLogAudit receives and stores audit log entries from the agent
func handleLogAudit(auditSvc *audit.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var entry audit.LogEntry

		if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
			slog.Error("failed to decode audit log", "error", err)
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid request body",
			})
			return
		}

		// Validate required fields
		if entry.Action == "" || entry.ResourceType == "" || entry.Status == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "action, resource_type, and status are required",
			})
			return
		}

		// Store audit log
		if err := auditSvc.Log(r.Context(), entry); err != nil {
			slog.Error("failed to store audit log", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Failed to store audit log",
			})
			return
		}

		slog.Info("audit log stored",
			"user_id", entry.UserID,
			"action", entry.Action,
			"resource", entry.ResourceID,
			"status", entry.Status,
		)

		writeJSON(w, http.StatusOK, map[string]string{
			"status": "success",
			"log_id": entry.ID,
		})
	}
}

// handleQueryAudit retrieves audit logs based on query parameters
func handleQueryAudit(auditSvc *audit.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		query := r.URL.Query()
		filters := audit.QueryFilters{
			UserID:       query.Get("user_id"),
			Action:       query.Get("action"),
			ResourceType: query.Get("resource_type"),
			Status:       query.Get("status"),
			Limit:        100, // Default limit
		}

		// Query audit logs
		entries, err := auditSvc.Query(r.Context(), filters)
		if err != nil {
			slog.Error("failed to query audit logs", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Failed to query audit logs",
			})
			return
		}

		writeJSON(w, http.StatusOK, entries)
	}
}
