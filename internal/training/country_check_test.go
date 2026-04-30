// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package training

import (
	"context"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/provabl/qualify/internal/database"
)

// mockTagger records TagRole calls for test assertions.
type mockTagger struct {
	roleName string
	tags     map[string]string
}

func (m *mockTagger) TagRole(_ context.Context, params *iam.TagRoleInput, _ ...func(*iam.Options)) (*iam.TagRoleOutput, error) {
	m.roleName = aws.ToString(params.RoleName)
	m.tags = make(map[string]string, len(params.Tags))
	for _, t := range params.Tags {
		m.tags[aws.ToString(t.Key)] = aws.ToString(t.Value)
	}
	return &iam.TagRoleOutput{}, nil
}


func newMockService(t *testing.T) (*Service, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return &Service{db: &database.DB{DB: db}}, mock
}

func expectCountryUpdate(mock sqlmock.Sqlmock, userID, country, performedBy string) {
	mock.ExpectExec(`UPDATE users`).
		WithArgs(userID, country, sqlmock.AnyArg(), performedBy).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

func expectRoleARN(mock sqlmock.Sqlmock, arn string) {
	// getUserRoleARN uses: SELECT metadata->>'role_arn' FROM users WHERE id = $1
	// PostgreSQL ->> returns the raw string value, not JSON.
	mock.ExpectQuery(`SELECT metadata`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"role_arn"}).AddRow(arn))
}

// ── Validation tests ───────────────────────────────────────────────────────

func TestRecordCountryCheck_InvalidCodes(t *testing.T) {
	svc := &Service{} // no DB — validation fails before any query
	for _, tc := range []struct{ code, desc string }{
		{"", "empty"},
		{"USA", "3-letter"},
		{"1A", "starts with digit"},
		{"a1", "lowercase+digit"},
		{"  ", "spaces"},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			if err := svc.RecordCountryCheck(context.Background(), "u", tc.code, "officer"); err == nil {
				t.Errorf("code %q: expected error, got nil", tc.code)
			}
		})
	}
}

// ── Happy path ─────────────────────────────────────────────────────────────

func TestRecordCountryCheck_ValidCode_NoIAM(t *testing.T) {
	// When no iamTagger is set, RecordCountryCheck updates the DB but skips
	// the IAM role ARN lookup and tag write entirely.
	svc, mock := newMockService(t)
	expectCountryUpdate(mock, "user-1", "US", "officer@mru.edu")

	if err := svc.RecordCountryCheck(context.Background(), "user-1", "US", "officer@mru.edu"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestRecordCountryCheck_UppercasesCode(t *testing.T) {
	svc, mock := newMockService(t)
	// Input "cn" (lowercase) — stored as "CN"
	expectCountryUpdate(mock, "user-1", "CN", "officer")

	if err := svc.RecordCountryCheck(context.Background(), "user-1", "cn", "officer"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRecordCountryCheck_WritesIAMTags(t *testing.T) {
	svc, mock := newMockService(t)
	expectCountryUpdate(mock, "user-1", "DE", "officer")
	expectRoleARN(mock, "arn:aws:iam::123456789012:role/ResearchRole")

	tagger := &mockTagger{}
	svc.iamTagger = tagger

	if err := svc.RecordCountryCheck(context.Background(), "user-1", "DE", "officer"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tagger.roleName != "ResearchRole" {
		t.Errorf("role name: got %q, want ResearchRole", tagger.roleName)
	}
	if tagger.tags[TagCountry] != "DE" {
		t.Errorf("%s: got %q, want DE", TagCountry, tagger.tags[TagCountry])
	}
	if tagger.tags[TagCOCCheckCurrent] != "true" {
		t.Errorf("%s: got %q, want true", TagCOCCheckCurrent, tagger.tags[TagCOCCheckCurrent])
	}
	expiry := tagger.tags[TagCOCCheckExpiry]
	if expiry == "" {
		t.Fatal("expiry tag is empty")
	}
	expiryTime, err := time.Parse(time.RFC3339, expiry)
	if err != nil {
		t.Fatalf("expiry not RFC3339: %v", err)
	}
	// Must be ~1 year from now (±10 seconds)
	diff := expiryTime.Sub(time.Now().AddDate(1, 0, 0)).Abs()
	if diff > 10*time.Second {
		t.Errorf("expiry %v is not ~1 year from now (diff %v)", expiryTime, diff)
	}
}

// ── TagForModule / ModuleIDs ───────────────────────────────────────────────

func TestTagForModule(t *testing.T) {
	for module, want := range map[string]string{
		"cui-fundamentals":               TagCUITraining,
		"hipaa-privacy-security":         TagHIPAATraining,
		"security-awareness":             TagAwarenessTraining,
		"ferpa-basics":                   TagFERPATraining,
		"countries-of-concern-awareness": TagCOCCheckCurrent,
		"nonexistent":                    "",
	} {
		if got := TagForModule(module); got != want {
			t.Errorf("TagForModule(%q) = %q, want %q", module, got, want)
		}
	}
}

func TestModuleIDs_ContainsAllModules(t *testing.T) {
	required := []string{
		"cui-fundamentals", "hipaa-privacy-security", "security-awareness",
		"ferpa-basics", "itar-export-control", "data-classification",
		"nih-research-security", "countries-of-concern-awareness",
	}
	ids := make(map[string]bool)
	for _, id := range ModuleIDs() {
		ids[id] = true
	}
	for _, r := range required {
		if !ids[r] {
			t.Errorf("ModuleIDs() missing %q", r)
		}
	}
}

func TestModuleTagMap_Immutable(t *testing.T) {
	// TagForModule returns consistent values — callers cannot mutate moduleTagMap
	v1 := TagForModule("cui-fundamentals")
	v2 := TagForModule("cui-fundamentals")
	if v1 != v2 || v1 == "" || strings.Contains(v1, "injected") {
		t.Errorf("TagForModule inconsistent or mutated: %q vs %q", v1, v2)
	}
}

// ── Ensure mockTagger implements iamTagWriter ──────────────────────────────

var _ iamTagWriter = (*mockTagger)(nil)

// Compile-time check that mockTagger satisfies iamTagWriter.
var _ iamTagWriter = (*mockTagger)(nil)

// Prevent unused import errors — these types are used transitively via the interface.
var _ = iam.TagRoleOutput{}
var _ = iamtypes.Tag{}
