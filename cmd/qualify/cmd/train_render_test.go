// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"strings"
	"testing"
)

func TestRenderTextSQLUnescape(t *testing.T) {
	// SQL-escaped apostrophes must be unescaped.
	out := renderText("institution''s policy")
	if !strings.Contains(out, "institution's policy") {
		t.Errorf("expected unescaped apostrophe, got: %q", out)
	}
}

func TestRenderTextBulletLists(t *testing.T) {
	input := "Intro\n- First item\n- Second item"
	out := renderText(input)
	if !strings.Contains(out, "• First item") {
		t.Errorf("expected bullet point, got: %q", out)
	}
	if !strings.Contains(out, "• Second item") {
		t.Errorf("expected second bullet, got: %q", out)
	}
}

func TestRenderTextNumberedLists(t *testing.T) {
	input := "Steps:\n1. Do this first\n2. Then do this"
	out := renderText(input)
	if !strings.Contains(out, "1. Do this first") {
		t.Errorf("expected numbered list, got: %q", out)
	}
}

func TestRenderTextStripsBoldInNonTTY(t *testing.T) {
	// isTTY is false in test environments (no real terminal).
	// Bold markers should be stripped.
	input := "This is **important** information."
	out := renderText(input)
	if strings.Contains(out, "**") {
		t.Errorf("expected bold markers stripped, got: %q", out)
	}
	if !strings.Contains(out, "important") {
		t.Errorf("expected bold text content preserved, got: %q", out)
	}
}

func TestRenderTextStripsHeaders(t *testing.T) {
	input := "# Section Title\n\nBody text here."
	out := renderText(input)
	if strings.Contains(out, "# ") {
		t.Errorf("expected header marker stripped, got: %q", out)
	}
	if !strings.Contains(out, "Section Title") {
		t.Errorf("expected header text preserved, got: %q", out)
	}
}

func TestRenderTextWrapsLongLines(t *testing.T) {
	long := strings.Repeat("word ", 30) // ~150 chars
	out := renderText(long)
	for _, line := range strings.Split(out, "\n") {
		if len(line) > 95 { // wrapWidth(78) + some ANSI slack
			t.Errorf("line too long (%d chars): %q", len(line), line)
		}
	}
}

func TestWrapText(t *testing.T) {
	cases := []struct {
		input string
		width int
		check func(string) bool
		desc  string
	}{
		{"short", 80, func(s string) bool { return s == "short" }, "short line unchanged"},
		{strings.Repeat("x ", 50), 20, func(s string) bool {
			for _, line := range strings.Split(s, "\n") {
				if len(line) > 22 {
					return false
				}
			}
			return true
		}, "long line wrapped at width"},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			out := wrapText(tc.input, tc.width)
			if !tc.check(out) {
				t.Errorf("wrapText(%q, %d) = %q", tc.input, tc.width, out)
			}
		})
	}
}
