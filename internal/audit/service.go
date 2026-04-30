// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/provabl/qualify/internal/database"
)

// Service provides audit logging functionality
type Service struct {
	db *database.DB
}

// NewService creates a new audit service
func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

// Log stores an audit log entry in the database
func (s *Service) Log(ctx context.Context, entry LogEntry) error {
	// Marshal details to JSON
	detailsJSON, err := json.Marshal(entry.Details)
	if err != nil {
		return fmt.Errorf("marshal details: %w", err)
	}

	// Insert into database
	query := `
		INSERT INTO audit_logs (user_id, action, resource_type, resource_id, status, details, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`

	var userID *string
	if entry.UserID != "" {
		userID = &entry.UserID
	}

	var ipAddress *string
	if entry.IPAddress != "" {
		ipAddress = &entry.IPAddress
	}

	var userAgent *string
	if entry.UserAgent != "" {
		userAgent = &entry.UserAgent
	}

	err = s.db.QueryRowContext(ctx, query,
		userID,
		entry.Action,
		entry.ResourceType,
		entry.ResourceID,
		entry.Status,
		detailsJSON,
		ipAddress,
		userAgent,
	).Scan(&entry.ID, &entry.CreatedAt)

	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}

	return nil
}

// Query retrieves audit logs based on filters
func (s *Service) Query(ctx context.Context, filters QueryFilters) ([]LogEntry, error) {
	// Build query with filters
	query := `
		SELECT id, user_id, action, resource_type, resource_id, status, details, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE 1=1
	`
	args := []interface{}{}
	argPos := 1

	if filters.UserID != "" {
		query += fmt.Sprintf(" AND user_id = $%d", argPos)
		args = append(args, filters.UserID)
		argPos++
	}

	if filters.Action != "" {
		query += fmt.Sprintf(" AND action = $%d", argPos)
		args = append(args, filters.Action)
		argPos++
	}

	if filters.ResourceType != "" {
		query += fmt.Sprintf(" AND resource_type = $%d", argPos)
		args = append(args, filters.ResourceType)
		argPos++
	}

	if filters.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, filters.Status)
		argPos++
	}

	if !filters.StartTime.IsZero() {
		query += fmt.Sprintf(" AND created_at >= $%d", argPos)
		args = append(args, filters.StartTime)
		argPos++
	}

	if !filters.EndTime.IsZero() {
		query += fmt.Sprintf(" AND created_at <= $%d", argPos)
		args = append(args, filters.EndTime)
		argPos++
	}

	// Order by created_at DESC
	query += " ORDER BY created_at DESC"

	// Apply limit and offset
	limit := filters.Limit
	if limit <= 0 {
		limit = 100 // Default limit
	}
	query += fmt.Sprintf(" LIMIT $%d", argPos)
	args = append(args, limit)
	argPos++

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, filters.Offset)
	}

	// Execute query
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query audit logs: %w", err)
	}
	defer rows.Close()

	// Parse results
	var entries []LogEntry
	for rows.Next() {
		var entry LogEntry
		var userID sql.NullString
		var ipAddress sql.NullString
		var userAgent sql.NullString
		var detailsJSON []byte

		err := rows.Scan(
			&entry.ID,
			&userID,
			&entry.Action,
			&entry.ResourceType,
			&entry.ResourceID,
			&entry.Status,
			&detailsJSON,
			&ipAddress,
			&userAgent,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		if userID.Valid {
			entry.UserID = userID.String
		}
		if ipAddress.Valid {
			entry.IPAddress = ipAddress.String
		}
		if userAgent.Valid {
			entry.UserAgent = userAgent.String
		}

		// Unmarshal details
		if len(detailsJSON) > 0 {
			if err := json.Unmarshal(detailsJSON, &entry.Details); err != nil {
				return nil, fmt.Errorf("unmarshal details: %w", err)
			}
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return entries, nil
}

// GetRecentLogs retrieves recent audit logs for a user
func (s *Service) GetRecentLogs(ctx context.Context, userID string, limit int) ([]LogEntry, error) {
	if limit <= 0 {
		limit = 50
	}

	return s.Query(ctx, QueryFilters{
		UserID: userID,
		Limit:  limit,
	})
}
