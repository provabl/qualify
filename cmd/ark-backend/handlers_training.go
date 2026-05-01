// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/provabl/qualify/internal/auth"
	"github.com/provabl/qualify/internal/training"
)

// handleCheckPolicy evaluates training gate policies for an action
func handleCheckPolicy(trainingSvc *training.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			UserID          string                 `json:"user_id"`
			Action          string                 `json:"action"`
			ResourceType    string                 `json:"resource_type"`
			ResourceDetails map[string]interface{} `json:"resource_details"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("failed to decode policy check request", "error", err)
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid request body",
			})
			return
		}

		// Always use the authenticated user from context — never trust body user_id.
		userID := auth.GetUserID(r.Context())
		if userID == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
			return
		}
		if req.Action == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "action is required"})
			return
		}

		decision, err := trainingSvc.CheckTrainingGate(r.Context(), userID, req.Action)
		if err != nil {
			slog.Error("failed to check training gate", "error", err, "user_id", userID, "action", req.Action)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to evaluate policy"})
			return
		}

		slog.Info("policy evaluated", "user_id", userID, "action", req.Action, "decision", decision.Action)

		writeJSON(w, http.StatusOK, decision)
	}
}

// handleGetUserProgress retrieves training progress for a user
func handleGetUserProgress(trainingSvc *training.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Route is /api/training/progress — user comes from JWT context
		userID := auth.GetUserID(r.Context())
		if userID == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
			return
		}

		progress, err := trainingSvc.GetUserProgress(r.Context(), userID)
		if err != nil {
			slog.Error("failed to get user progress",
				"error", err,
				"user_id", userID,
			)
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Failed to retrieve training progress",
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"user_id":  userID,
			"progress": progress,
		})
	}
}

// handleListModules returns all active training modules
func handleListModules(trainingSvc *training.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		modules, err := trainingSvc.ListModules(r.Context())
		if err != nil {
			slog.Error("failed to list modules", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Failed to retrieve training modules",
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"modules": modules,
		})
	}
}

// handleGetModule returns a specific module with full content
func handleGetModule(trainingSvc *training.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		moduleID := chi.URLParam(r, "id")
		if moduleID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "module id parameter required",
			})
			return
		}

		module, err := trainingSvc.GetModule(r.Context(), moduleID)
		if err != nil {
			slog.Error("failed to get module",
				"error", err,
				"module_id", moduleID,
			)
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "Module not found",
			})
			return
		}

		writeJSON(w, http.StatusOK, module)
	}
}

// handleStartModule marks a module as started for a user
func handleStartModule(trainingSvc *training.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		moduleID := chi.URLParam(r, "id")
		if moduleID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "module id parameter required",
			})
			return
		}

		// User comes from JWT context — body user_id is ignored for security
		userID := auth.GetUserID(r.Context())
		if userID == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
			return
		}

		err := trainingSvc.StartModule(r.Context(), userID, moduleID)
		if err != nil {
			slog.Error("failed to start module",
				"error", err,
				"module_id", moduleID,
			)
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Failed to start module",
			})
			return
		}

		slog.Info("module started", "user_id", userID, "module_id", moduleID)

		writeJSON(w, http.StatusOK, map[string]interface{}{"status": "started", "user_id": userID, "module_id": moduleID})
	}
}

// handleCompleteModule marks a module as completed for a user
func handleCompleteModule(trainingSvc *training.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		moduleID := chi.URLParam(r, "id")
		if moduleID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "module id parameter required",
			})
			return
		}

		var req struct {
			UserID    string `json:"user_id"`
			Score     int    `json:"score"`
			TimeSpent int    `json:"time_spent_seconds"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("failed to decode complete module request", "error", err)
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid request body",
			})
			return
		}

		userID := auth.GetUserID(r.Context())
		if userID == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
			return
		}

		err := trainingSvc.CompleteModule(r.Context(), userID, moduleID, req.Score)
		if err != nil {
			slog.Error("failed to complete module",
				"error", err,
				"module_id", moduleID,
			)
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Failed to complete module",
			})
			return
		}

		slog.Info("module completed", "user_id", userID, "module_id", moduleID, "score", req.Score)

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status":    "completed",
			"message":   "Module completed successfully",
			"user_id": userID, "module_id": moduleID})
	}
}

// handleSubmitQuiz evaluates quiz answers and returns results
func handleSubmitQuiz(trainingSvc *training.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		moduleID := chi.URLParam(r, "id")
		if moduleID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "module id parameter required",
			})
			return
		}

		var submission training.QuizSubmission
		if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
			slog.Error("failed to decode quiz submission", "error", err)
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid request body",
			})
			return
		}

		// Override body user_id with authenticated user from context.
		submission.UserID = auth.GetUserID(r.Context())
		if submission.UserID == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
			return
		}

		if len(submission.Answers) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "answers are required",
			})
			return
		}

		result, err := trainingSvc.SubmitQuiz(r.Context(), moduleID, submission)
		if err != nil {
			slog.Error("failed to submit quiz",
				"error", err,
				"user_id", submission.UserID,
				"module_id", moduleID,
			)
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Failed to evaluate quiz",
			})
			return
		}

		slog.Info("quiz submitted",
			"user_id", submission.UserID,
			"module_id", moduleID,
			"score", result.Score,
		)

		writeJSON(w, http.StatusOK, result)
	}
}

// handleGetUserActivity retrieves user training activity log
func handleGetUserActivity(trainingSvc *training.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.GetUserID(r.Context())
		if userID == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
			return
		}

		// Parse query parameters for pagination
		limit := 10
		offset := 0
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if n, err := parseIntQueryParam(limitStr); err == nil && n > 0 {
				limit = n
			}
		}
		if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
			if n, err := parseIntQueryParam(offsetStr); err == nil && n >= 0 {
				offset = n
			}
		}

		activities, err := trainingSvc.GetUserActivity(r.Context(), userID, limit, offset)
		if err != nil {
			slog.Error("failed to get user activity",
				"error", err,
				"user_id", userID,
			)
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Failed to retrieve activity",
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"user_id":    userID,
			"activities": activities,
		})
	}
}
