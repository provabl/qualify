// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

// Package localaudit writes a JSONL audit log to ~/.qualify/audit.log.
//
// This log is always available — it does not require the backend database.
// Each line is a JSON object with at minimum:
//
//	{"ts":"<RFC3339>","event":"<action>","user":"<id>","module":"<id>",...}
//
// Events:
//   - module_started       user began a training module
//   - section_completed    user advanced past a section
//   - quiz_attempt         user submitted quiz answers (includes score, pass/fail)
//   - module_completed     module marked complete, IAM tags written
//   - module_failed        user exhausted retries without passing
//   - operation_blocked    an AWS operation was denied due to incomplete training
//   - iam_tag_written      IAM tag was written after completion
package localaudit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Entry is a single audit log record.
type Entry struct {
	Timestamp time.Time      `json:"ts"`
	Event     string         `json:"event"`
	UserID    string         `json:"user,omitempty"`
	Module    string         `json:"module,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
}

// Logger writes audit entries to ~/.qualify/audit.log.
type Logger struct {
	mu   sync.Mutex
	path string
}

// New creates a Logger writing to ~/.qualify/audit.log.
// Returns an error only if the home directory cannot be determined.
func New() (*Logger, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("qualify audit: home dir: %w", err)
	}
	dir := filepath.Join(home, ".qualify")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("qualify audit: mkdir: %w", err)
	}
	return &Logger{path: filepath.Join(dir, "audit.log")}, nil
}

// Log appends an audit entry. Errors are silently dropped — audit logging
// must never block or fail the primary operation.
func (l *Logger) Log(event, userID, moduleID string, details map[string]any) {
	entry := Entry{
		Timestamp: time.Now().UTC(),
		Event:     event,
		UserID:    userID,
		Module:    moduleID,
		Details:   details,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600) // #nosec G304 — fixed path under ~/.qualify
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = fmt.Fprintf(f, "%s\n", data)
}

// Convenience methods for common events.

func (l *Logger) ModuleStarted(userID, moduleID string) {
	l.Log("module_started", userID, moduleID, nil)
}

func (l *Logger) SectionCompleted(userID, moduleID string, sectionIndex, total int) {
	l.Log("section_completed", userID, moduleID, map[string]any{
		"section": sectionIndex,
		"total":   total,
	})
}

func (l *Logger) QuizAttempt(userID, moduleID string, attempt, score, passing int, passed bool) {
	l.Log("quiz_attempt", userID, moduleID, map[string]any{
		"attempt": attempt,
		"score":   score,
		"passing": passing,
		"passed":  passed,
	})
}

func (l *Logger) ModuleCompleted(userID, moduleID string, score int) {
	l.Log("module_completed", userID, moduleID, map[string]any{"score": score})
}

func (l *Logger) ModuleFailed(userID, moduleID string, finalScore, passing int) {
	l.Log("module_failed", userID, moduleID, map[string]any{
		"final_score": finalScore,
		"passing":     passing,
	})
}

func (l *Logger) IAMTagWritten(userID, moduleID, tagKey string) {
	l.Log("iam_tag_written", userID, moduleID, map[string]any{"tag": tagKey})
}
