package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/provabl/ark/internal/agent/aws"
)

// handleCreateBucket handles S3 bucket creation requests
func (s *server) handleCreateBucket(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BucketName string `json:"bucket_name"`
		Region     string `json:"region"`
		Encryption struct {
			Type     string `json:"type"`
			KMSKeyID string `json:"kms_key_id,omitempty"`
		} `json:"encryption"`
		VersioningEnabled bool   `json:"versioning_enabled"`
		Profile           string `json:"profile"`
	}

	// Parse request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("failed to decode request", "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	// Validate required fields
	if req.BucketName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "bucket_name is required",
		})
		return
	}

	// Default profile
	if req.Profile == "" {
		req.Profile = "default"
	}

	// Get credentials from store
	creds, err := s.store.GetCredential(req.Profile)
	if err != nil {
		slog.Error("failed to get credentials", "error", err, "profile", req.Profile)
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "Credentials not found for profile: " + req.Profile,
		})
		return
	}

	// Check training gate with backend
	allowed, requiredModules, err := s.checkTrainingGate("s3:CreateBucket", map[string]interface{}{
		"bucket_name": req.BucketName,
		"region":      req.Region,
	})
	if err != nil {
		slog.Error("failed to check training gate", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to check training requirements",
		})
		return
	}

	if !allowed {
		slog.Info("operation blocked by training gate",
			"user", getCurrentUser(),
			"action", "s3:CreateBucket",
			"bucket", req.BucketName,
		)

		// Send audit log for blocked operation
		go s.sendAuditLog(r.Context(), map[string]interface{}{
			"action":        "s3:CreateBucket",
			"resource_type": "s3:bucket",
			"resource_id":   req.BucketName,
			"status":        "blocked",
			"details": map[string]interface{}{
				"region":           req.Region,
				"required_modules": requiredModules,
			},
		})

		writeJSON(w, http.StatusForbidden, map[string]interface{}{
			"status":           "blocked",
			"reason":           "training_required",
			"required_modules": requiredModules,
		})
		return
	}

	// Create AWS client
	client, err := aws.NewClientFromCredentials(r.Context(), creds, req.Region)
	if err != nil {
		slog.Error("failed to create AWS client", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to initialize AWS client",
		})
		return
	}

	// Default encryption type
	if req.Encryption.Type == "" {
		req.Encryption.Type = "AES256"
	}

	// Execute S3 CreateBucket
	slog.Info("creating S3 bucket",
		"bucket", req.BucketName,
		"region", req.Region,
		"encryption", req.Encryption.Type,
		"versioning", req.VersioningEnabled,
	)

	output, err := aws.CreateBucket(r.Context(), client, aws.CreateBucketInput{
		BucketName:        req.BucketName,
		Region:            req.Region,
		EncryptionType:    req.Encryption.Type,
		KMSKeyID:          req.Encryption.KMSKeyID,
		VersioningEnabled: req.VersioningEnabled,
	})

	if err != nil {
		slog.Error("failed to create bucket",
			"error", err,
			"bucket", req.BucketName,
		)

		// Send audit log for failed operation
		go s.sendAuditLog(r.Context(), map[string]interface{}{
			"action":        "s3:CreateBucket",
			"resource_type": "s3:bucket",
			"resource_id":   req.BucketName,
			"status":        "failure",
			"details": map[string]interface{}{
				"region":     req.Region,
				"encryption": req.Encryption.Type,
				"versioning": req.VersioningEnabled,
				"error":      err.Error(),
			},
		})

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	slog.Info("bucket created successfully",
		"bucket", output.BucketName,
		"region", output.Region,
		"location", output.Location,
	)

	// Send audit log to backend (non-blocking)
	go s.sendAuditLog(r.Context(), map[string]interface{}{
		"action":        "s3:CreateBucket",
		"resource_type": "s3:bucket",
		"resource_id":   output.BucketName,
		"status":        "success",
		"details": map[string]interface{}{
			"region":     output.Region,
			"encryption": req.Encryption.Type,
			"versioning": req.VersioningEnabled,
		},
	})

	writeJSON(w, http.StatusCreated, output)
}

// getBackendURL returns the backend URL from environment or default
func getBackendURL() string {
	if url := os.Getenv("ARK_BACKEND_URL"); url != "" {
		return url
	}
	return "http://localhost:8080"
}

// getCurrentUser returns the current system user as a placeholder for user_id
// TODO: Replace with proper authentication when auth is implemented
func getCurrentUser() string {
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	if user := os.Getenv("USERNAME"); user != "" {
		return user
	}
	return "unknown"
}

// checkTrainingGate checks with backend if user is allowed to perform action
func (s *server) checkTrainingGate(action string, resourceDetails map[string]interface{}) (bool, []map[string]interface{}, error) {
	backendURL := getBackendURL()
	url := backendURL + "/api/policies/check"

	reqBody := map[string]interface{}{
		"user_id":          getCurrentUser(),
		"action":           action,
		"resource_type":    "s3:bucket",
		"resource_details": resourceDetails,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return false, nil, err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		// If backend is unavailable, log warning and allow (graceful degradation)
		slog.Warn("backend unavailable for policy check, allowing operation", "error", err)
		return true, nil, nil
	}
	defer resp.Body.Close()

	var result struct {
		Action          string                   `json:"action"`
		Reason          string                   `json:"reason,omitempty"`
		RequiredModules []map[string]interface{} `json:"required_modules,omitempty"`
		Message         string                   `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.Warn("failed to decode policy response", "error", err)
		return true, nil, nil
	}

	if result.Action == "block" {
		return false, result.RequiredModules, nil
	}

	return true, nil, nil
}

// sendAuditLog sends an audit log entry to the backend (non-blocking)
func (s *server) sendAuditLog(ctx interface{}, logData map[string]interface{}) {
	backendURL := getBackendURL()
	url := backendURL + "/api/audit/log"

	// Add user_id if not present
	if _, ok := logData["user_id"]; !ok {
		logData["user_id"] = getCurrentUser()
	}

	bodyBytes, err := json.Marshal(logData)
	if err != nil {
		slog.Warn("failed to marshal audit log", "error", err)
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		slog.Warn("failed to send audit log to backend", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("backend returned non-200 for audit log", "status", resp.StatusCode)
		return
	}

	slog.Debug("audit log sent to backend successfully")
}
