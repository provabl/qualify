package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/provabl/qualify/internal/agent/store"
)

// handleSetCredentials stores AWS credentials for a profile
func (s *server) handleSetCredentials(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Profile         string `json:"profile"`
		AccessKeyID     string `json:"access_key_id"`
		SecretAccessKey string `json:"secret_access_key"`
		SessionToken    string `json:"session_token,omitempty"`
		Region          string `json:"region"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	// Validate required fields
	if req.Profile == "" || req.AccessKeyID == "" || req.SecretAccessKey == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "profile, access_key_id, and secret_access_key are required",
		})
		return
	}

	// Default region
	if req.Region == "" {
		req.Region = "us-east-1"
	}

	// Create credentials object
	creds := store.Credentials{
		AccessKeyID:     req.AccessKeyID,
		SecretAccessKey: req.SecretAccessKey,
		SessionToken:    req.SessionToken,
		Region:          req.Region,
		Expiration:      time.Time{}, // No expiration for long-term credentials
	}

	// Store in BoltDB
	if err := s.store.SetCredential(req.Profile, creds); err != nil {
		slog.Error("failed to store credentials", "error", err, "profile", req.Profile)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to store credentials",
		})
		return
	}

	slog.Info("credentials stored", "profile", req.Profile, "region", req.Region)

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"profile": req.Profile,
	})
}

// handleListCredentials lists all stored credential profiles
func (s *server) handleListCredentials(w http.ResponseWriter, r *http.Request) {
	profiles, err := s.store.ListCredentials()
	if err != nil {
		slog.Error("failed to list credentials", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to list credentials",
		})
		return
	}

	// Convert to response format
	type profileInfo struct {
		Profile string `json:"profile"`
		Region  string `json:"region"`
	}

	var result []profileInfo
	for profile, creds := range profiles {
		result = append(result, profileInfo{
			Profile: profile,
			Region:  creds.Region,
		})
	}

	writeJSON(w, http.StatusOK, result)
}

// handleDeleteCredentials deletes a credential profile
func (s *server) handleDeleteCredentials(w http.ResponseWriter, r *http.Request) {
	profile := chi.URLParam(r, "profile")
	if profile == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "profile parameter required",
		})
		return
	}

	// Check if exists
	_, err := s.store.GetCredential(profile)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "Profile not found",
		})
		return
	}

	// Delete from store
	if err := s.store.DeleteCredential(profile); err != nil {
		slog.Error("failed to delete credentials", "error", err, "profile", profile)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete credentials",
		})
		return
	}

	slog.Info("credentials deleted", "profile", profile)

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "success",
	})
}
