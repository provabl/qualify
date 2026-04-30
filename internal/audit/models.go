// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"time"
)

// LogEntry represents an audit log entry
type LogEntry struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id"`
	Status       string                 `json:"status"` // success, failure, blocked
	Details      map[string]interface{} `json:"details"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// QueryFilters represents filters for querying audit logs
type QueryFilters struct {
	UserID       string
	Action       string
	ResourceType string
	Status       string
	StartTime    time.Time
	EndTime      time.Time
	Limit        int
	Offset       int
}
