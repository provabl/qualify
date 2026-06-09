// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package training

import (
	"context"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/provabl/qualify/internal/database"
)

// countingTagger records whether TagRole was called and with what tags.
type countingTagger struct {
	calls int
	tags  map[string]string
}

func (m *countingTagger) TagRole(_ context.Context, params *iam.TagRoleInput, _ ...func(*iam.Options)) (*iam.TagRoleOutput, error) {
	m.calls++
	m.tags = map[string]string{}
	for _, t := range params.Tags {
		m.tags[aws.ToString(t.Key)] = aws.ToString(t.Value)
	}
	return &iam.TagRoleOutput{}, nil
}

// newCompleteModuleService wires a sqlmock DB with relaxed ordering and the
// queries CompleteModule issues (progress upsert, activity name-lookup + insert,
// role ARN lookup), plus the IAM tagger.
func newCompleteModuleService(t *testing.T, roleARN string, tagger iamTagWriter) (*Service, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	mock.MatchExpectationsInOrder(false)

	mock.ExpectExec(`INSERT INTO user_training_progress`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	// RecordActivity: name lookup (best-effort) + insert.
	mock.ExpectQuery(`SELECT name FROM training_modules`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Module"))
	mock.ExpectExec(`INSERT INTO training_activities`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	// getUserRoleARN lookup.
	mock.ExpectQuery(`SELECT metadata`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"role_arn"}).AddRow(roleARN))

	svc := &Service{db: &database.DB{DB: db}, iamTagger: tagger}
	return svc, mock
}

func TestCompleteModule_PassWritesTags(t *testing.T) {
	tagger := &countingTagger{}
	svc, _ := newCompleteModuleService(t, "arn:aws:iam::123456789012:role/ResearchRole", tagger)

	if err := svc.CompleteModule(context.Background(), "user-1", "cui-fundamentals", 90, 80); err != nil {
		t.Fatalf("CompleteModule: %v", err)
	}
	if tagger.calls != 1 {
		t.Fatalf("TagRole calls = %d, want 1", tagger.calls)
	}
	if tagger.tags[TagCUITraining] != "true" {
		t.Errorf("%s = %q, want true", TagCUITraining, tagger.tags[TagCUITraining])
	}
	exp := tagger.tags[TagCUITrainingExpiry]
	if _, err := time.Parse(time.RFC3339, exp); err != nil {
		t.Errorf("expiry %q not RFC3339: %v", exp, err)
	}
}

func TestCompleteModule_FailWritesNothing(t *testing.T) {
	tagger := &countingTagger{}
	svc, _ := newCompleteModuleService(t, "arn:aws:iam::123456789012:role/ResearchRole", tagger)

	// Score below passing — the appraiser fails, so no tags are written. This is
	// the behavior change that fixes the backend writing tags for failing scores.
	if err := svc.CompleteModule(context.Background(), "user-1", "cui-fundamentals", 50, 80); err != nil {
		t.Fatalf("CompleteModule: %v", err)
	}
	if tagger.calls != 0 {
		t.Errorf("TagRole calls = %d, want 0 (failing score must not write tags)", tagger.calls)
	}
}

func TestCompleteModule_UnmappedModuleWritesNothing(t *testing.T) {
	tagger := &countingTagger{}
	svc, _ := newCompleteModuleService(t, "arn:aws:iam::123456789012:role/ResearchRole", tagger)

	// A module with no attest tag mapping is NotApplicable — no tags written.
	if err := svc.CompleteModule(context.Background(), "user-1", "some-unmapped-module", 100, 80); err != nil {
		t.Fatalf("CompleteModule: %v", err)
	}
	if tagger.calls != 0 {
		t.Errorf("TagRole calls = %d, want 0 (unmapped module must not write tags)", tagger.calls)
	}
}
