// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package localaudit

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLogWritesJSONL(t *testing.T) {
	dir := t.TempDir()
	l := &Logger{path: filepath.Join(dir, "audit.log")}

	l.ModuleStarted("user-1", "cui-fundamentals")
	l.SectionCompleted("user-1", "cui-fundamentals", 1, 3)
	l.QuizAttempt("user-1", "cui-fundamentals", 1, 80, 80, true)
	l.ModuleCompleted("user-1", "cui-fundamentals", 80)

	f, err := os.Open(l.path)
	if err != nil {
		t.Fatalf("open audit log: %v", err)
	}
	defer f.Close()

	var entries []Entry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var e Entry
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			t.Fatalf("unmarshal line: %v", err)
		}
		entries = append(entries, e)
	}

	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}

	events := []string{"module_started", "section_completed", "quiz_attempt", "module_completed"}
	for i, want := range events {
		if entries[i].Event != want {
			t.Errorf("entry %d: want event %q, got %q", i, want, entries[i].Event)
		}
		if entries[i].UserID != "user-1" {
			t.Errorf("entry %d: want user-1, got %q", i, entries[i].UserID)
		}
		if entries[i].Module != "cui-fundamentals" {
			t.Errorf("entry %d: want cui-fundamentals, got %q", i, entries[i].Module)
		}
	}
}

func TestLogTimestampsUTC(t *testing.T) {
	dir := t.TempDir()
	l := &Logger{path: filepath.Join(dir, "audit.log")}
	before := time.Now().UTC()
	l.ModuleStarted("u", "m")
	after := time.Now().UTC()

	f, _ := os.Open(l.path)
	defer f.Close()
	var e Entry
	json.NewDecoder(f).Decode(&e) //nolint:errcheck

	if e.Timestamp.Before(before) || e.Timestamp.After(after) {
		t.Errorf("timestamp %v outside [%v, %v]", e.Timestamp, before, after)
	}
	if e.Timestamp.Location() != time.UTC {
		t.Errorf("timestamp not UTC: %v", e.Timestamp.Location())
	}
}

func TestLogQuizDetails(t *testing.T) {
	dir := t.TempDir()
	l := &Logger{path: filepath.Join(dir, "audit.log")}
	l.QuizAttempt("u", "m", 2, 60, 80, false)

	f, _ := os.Open(l.path)
	defer f.Close()
	var e Entry
	json.NewDecoder(f).Decode(&e) //nolint:errcheck

	if e.Details["passed"] != false {
		t.Errorf("expected passed=false, got %v", e.Details["passed"])
	}
	if e.Details["attempt"].(float64) != 2 {
		t.Errorf("expected attempt=2, got %v", e.Details["attempt"])
	}
}

func TestLogSilentOnError(t *testing.T) {
	// Logging to an unwritable path must not panic or propagate error.
	l := &Logger{path: "/nonexistent-dir/that/cannot/exist/audit.log"}
	l.ModuleStarted("u", "m") // must not panic
}
