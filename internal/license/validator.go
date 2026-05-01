// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package license

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/provabl/qualify/internal/database"
)

const cacheKey = "license_info"

// Validator validates a license key against the Provabl licensing server
// and caches the result in the system_config table.
type Validator struct {
	endpoint string
	key      string
	cacheTTL time.Duration
	db       *database.DB
}

// NewValidator creates a license validator.
// endpoint: HTTPS URL of the Provabl licensing API
// key: customer license key (empty = community tier)
// cacheTTL: how long to cache a valid license response
func NewValidator(endpoint, key string, cacheTTL time.Duration, db *database.DB) *Validator {
	return &Validator{endpoint: endpoint, key: key, cacheTTL: cacheTTL, db: db}
}

// ValidateAndCache validates the license, using the DB cache when possible.
// Returns CommunityLicense if no key is configured or if validation fails with no cache.
func (v *Validator) ValidateAndCache(ctx context.Context) (*Info, error) {
	// No key = community tier, no validation needed.
	if v.key == "" {
		return CommunityLicense(), nil
	}

	// Try DB cache first.
	if cached := v.loadCache(ctx); cached != nil {
		return cached, nil
	}

	// Call licensing server.
	info, err := v.callServer(ctx)
	if err != nil {
		// Validation failed; no usable cache. Warn and fall back to community.
		return nil, fmt.Errorf("licensing server unreachable: %w", err)
	}

	v.saveCache(ctx, info)
	return info, nil
}

func (v *Validator) loadCache(ctx context.Context) *Info {
	var raw []byte
	var expiresAt time.Time
	err := v.db.QueryRowContext(ctx,
		`SELECT value, expires_at FROM system_config WHERE key = $1 AND expires_at > NOW()`,
		cacheKey,
	).Scan(&raw, &expiresAt)
	if err != nil {
		return nil
	}
	var info Info
	if err := json.Unmarshal(raw, &info); err != nil {
		return nil
	}
	return &info
}

func (v *Validator) saveCache(ctx context.Context, info *Info) {
	raw, err := json.Marshal(info)
	if err != nil {
		return
	}
	expiresAt := time.Now().Add(v.cacheTTL)
	_, _ = v.db.ExecContext(ctx,
		`INSERT INTO system_config (key, value, expires_at, updated_at)
		 VALUES ($1, $2, $3, NOW())
		 ON CONFLICT (key) DO UPDATE SET value = $2, expires_at = $3, updated_at = NOW()`,
		cacheKey, raw, expiresAt,
	)
}

type validateRequest struct {
	LicenseKey     string `json:"license_key"`
	DeploymentType string `json:"deployment_type"`
}

type validateResponse struct {
	Valid        bool            `json:"valid"`
	Tier         string          `json:"tier"`
	ExpiresAt    string          `json:"expires_at"`
	Company      string          `json:"company"`
	ContentPacks []string        `json:"content_packs"`
	Features     map[string]bool `json:"features"`
	Message      string          `json:"message,omitempty"`
}

func (v *Validator) callServer(ctx context.Context) (*Info, error) {
	body, _ := json.Marshal(validateRequest{
		LicenseKey:     v.key,
		DeploymentType: "self-hosted",
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "qualify-backend/1.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("licensing server returned %d", resp.StatusCode)
	}

	var vr validateResponse
	if err := json.NewDecoder(resp.Body).Decode(&vr); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if !vr.Valid {
		return nil, fmt.Errorf("license invalid: %s", vr.Message)
	}

	info := &Info{
		Valid:        vr.Valid,
		Tier:         vr.Tier,
		Company:      vr.Company,
		ContentPacks: vr.ContentPacks,
		Features:     vr.Features,
	}
	if t, err := time.Parse(time.RFC3339, vr.ExpiresAt); err == nil {
		info.ExpiresAt = t
	}
	return info, nil
}
