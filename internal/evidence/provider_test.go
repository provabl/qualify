// SPDX-FileCopyrightText: 2026 Playground Logic LLC
// SPDX-License-Identifier: Apache-2.0

package qualifyasp_test

import (
	"context"
	"testing"
	"time"

	"github.com/provabl/evidence/asp"
	"github.com/provabl/evidence/cvm"
	"github.com/provabl/evidence/lower"
	"github.com/provabl/evidence/term"

	qualifyasp "github.com/provabl/qualify/internal/evidence"
)

// appraise builds a CVM with the ephemeral AM, runs the canonical term for the
// given completion input, and appraises — the same path writeAttestTags drives.
func appraise(t *testing.T, in qualifyasp.Input) asp.Verdict {
	t.Helper()
	reg := asp.NewRegistry()
	if err := reg.Register(qualifyasp.Provider(in)); err != nil {
		t.Fatalf("register: %v", err)
	}
	am, err := qualifyasp.NewEphemeralAM()
	if err != nil {
		t.Fatalf("am: %v", err)
	}
	c := cvm.New(reg, am, am, nil)
	protocol := term.Seq(
		term.Nonce(),
		term.Seq(term.Meas(term.Self, qualifyasp.ID, qualifyasp.Target("arn:aws:iam::123456789012:role/r"), term.Params{}), term.Sig()),
	)
	bundle, ch, err := c.Run(context.Background(), protocol)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	v, err := c.Appraise(context.Background(), bundle, ch)
	if err != nil {
		t.Fatalf("appraise: %v", err)
	}
	return v
}

func passInput() qualifyasp.Input {
	return qualifyasp.Input{
		ModuleID:     "cui-fundamentals",
		TagKey:       "attest:cui-training",
		ExpiryTagKey: "attest:cui-training-expiry",
		Score:        90,
		PassingScore: 80,
		ExpiresAt:    time.Now().Add(365 * 24 * time.Hour).UTC().Format(time.RFC3339),
	}
}

func TestProvider_PassEmitsTagClaims(t *testing.T) {
	v := appraise(t, passInput())
	if !v.Pass {
		t.Fatalf("expected pass, reason: %s", v.Reason)
	}
	attrs := lower.ToAttributes(v)

	// Exact attest:* key set (plus synthetic "attested"), nothing else.
	want := map[string]string{
		"attest:cui-training":        "bool",
		"attest:cui-training-expiry": "string",
		"attested":                   "bool",
	}
	if len(attrs) != len(want) {
		t.Errorf("attr set has %d keys, want %d: %v", len(attrs), len(want), attrs)
	}
	for k, typ := range want {
		a, ok := attrs[k]
		if !ok {
			t.Errorf("missing attribute %q", k)
			continue
		}
		if a.Type != typ {
			t.Errorf("attr %q type = %q, want %q", k, a.Type, typ)
		}
	}
	if attrs["attest:cui-training"].Value != "true" {
		t.Errorf("cui-training = %q, want true", attrs["attest:cui-training"].Value)
	}
	if _, err := time.Parse(time.RFC3339, attrs["attest:cui-training-expiry"].Value); err != nil {
		t.Errorf("expiry %q not RFC3339: %v", attrs["attest:cui-training-expiry"].Value, err)
	}
	if attrs["attested"].Value != "true" {
		t.Errorf("attested = %q, want true", attrs["attested"].Value)
	}
}

func TestProvider_FailEmitsNoTagClaims(t *testing.T) {
	in := passInput()
	in.Score = 50 // below 80
	v := appraise(t, in)
	if v.Pass {
		t.Fatal("expected fail for score below passing")
	}
	attrs := lower.ToAttributes(v)
	for k := range attrs {
		if qualifyasp.IsAttestTag(k) {
			t.Errorf("failing verdict must emit no attest:* claims, got %q", k)
		}
	}
	if attrs["attested"].Value != "false" {
		t.Errorf("attested = %q, want false", attrs["attested"].Value)
	}
}

// An unmapped module (no tag key) is NotApplicable — distinct from a failure.
// The appraiser is never invoked; the kernel emits training.applicable=false and
// no attest:* claims, and overall pass is NOT degraded.
func TestProvider_UnmappedModuleNotApplicable(t *testing.T) {
	in := qualifyasp.Input{ModuleID: "some-unmapped-module", TagKey: "", Score: 100, PassingScore: 80}
	v := appraise(t, in)
	attrs := lower.ToAttributes(v)
	for k := range attrs {
		if qualifyasp.IsAttestTag(k) {
			t.Errorf("NotApplicable must emit no attest:* claims, got %q", k)
		}
	}
	if attrs["training.applicable"].Value != "false" {
		t.Errorf("expected training.applicable=false, got %q", attrs["training.applicable"].Value)
	}
}

// Freshness: appraising against a different challenge than the one issued must
// fail at the spine before any claim is trusted.
func TestProvider_FreshnessNonceMismatch(t *testing.T) {
	reg := asp.NewRegistry()
	if err := reg.Register(qualifyasp.Provider(passInput())); err != nil {
		t.Fatal(err)
	}
	am, err := qualifyasp.NewEphemeralAM()
	if err != nil {
		t.Fatal(err)
	}
	c := cvm.New(reg, am, am, nil)
	protocol := term.Seq(
		term.Nonce(),
		term.Seq(term.Meas(term.Self, qualifyasp.ID, qualifyasp.Target("x"), term.Params{}), term.Sig()),
	)
	bundle, _, err := c.Run(context.Background(), protocol)
	if err != nil {
		t.Fatal(err)
	}
	stale := cvm.Challenge{Nonce: []byte("a-different-32-byte-challenge....")}
	v, err := c.Appraise(context.Background(), bundle, stale)
	if err != nil {
		t.Fatal(err)
	}
	if v.Pass {
		t.Fatal("expected freshness failure on nonce mismatch")
	}
}

// appraiseAttrs runs the attributes (identity / coc) pair through a CVM and
// returns the lowered attributes — the same path SetIdentityTags/RecordCountryCheck drive.
func appraiseAttrs(t *testing.T, in qualifyasp.AttributesInput) map[string]lower.Attr {
	t.Helper()
	reg := asp.NewRegistry()
	if err := reg.Register(qualifyasp.AttributesProvider(in)); err != nil {
		t.Fatalf("register: %v", err)
	}
	am, err := qualifyasp.NewEphemeralAM()
	if err != nil {
		t.Fatalf("am: %v", err)
	}
	c := cvm.New(reg, am, am, nil)
	protocol := term.Seq(
		term.Nonce(),
		term.Seq(term.Meas(term.Self, qualifyasp.AttrID, qualifyasp.Target("arn:aws:iam::123456789012:role/r"), term.Params{}), term.Sig()),
	)
	bundle, ch, err := c.Run(context.Background(), protocol)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	v, err := c.Appraise(context.Background(), bundle, ch)
	if err != nil {
		t.Fatalf("appraise: %v", err)
	}
	if !v.Pass {
		t.Fatalf("attributes pair must always pass, reason: %s", v.Reason)
	}
	return lower.ToAttributes(v)
}

func TestAttributesProvider_EmitsTagsVerbatim(t *testing.T) {
	attrs := appraiseAttrs(t, qualifyasp.AttributesInput{Tags: []qualifyasp.Tag{
		{Key: "attest:lab-id", Value: "genomics-lab"},
		{Key: "attest:admin-level", Value: "env"},
	}})
	if attrs["attest:lab-id"].Value != "genomics-lab" {
		t.Errorf("attest:lab-id = %q, want genomics-lab", attrs["attest:lab-id"].Value)
	}
	if attrs["attest:admin-level"].Value != "env" {
		t.Errorf("attest:admin-level = %q, want env", attrs["attest:admin-level"].Value)
	}
	// Only the two attest:* keys plus the synthetic "attested".
	attestKeys := 0
	for k := range attrs {
		if qualifyasp.IsAttestTag(k) {
			attestKeys++
		}
	}
	if attestKeys != 2 {
		t.Errorf("expected 2 attest:* attrs, got %d: %v", attestKeys, attrs)
	}
}

func TestAttributesProvider_EmptyIsNotApplicable(t *testing.T) {
	reg := asp.NewRegistry()
	if err := reg.Register(qualifyasp.AttributesProvider(qualifyasp.AttributesInput{})); err != nil {
		t.Fatal(err)
	}
	am, _ := qualifyasp.NewEphemeralAM()
	c := cvm.New(reg, am, am, nil)
	protocol := term.Seq(term.Nonce(), term.Seq(term.Meas(term.Self, qualifyasp.AttrID, qualifyasp.Target("x"), term.Params{}), term.Sig()))
	bundle, ch, err := c.Run(context.Background(), protocol)
	if err != nil {
		t.Fatal(err)
	}
	v, err := c.Appraise(context.Background(), bundle, ch)
	if err != nil {
		t.Fatal(err)
	}
	attrs := lower.ToAttributes(v)
	for k := range attrs {
		if qualifyasp.IsAttestTag(k) {
			t.Errorf("NotApplicable must emit no attest:* claims, got %q", k)
		}
	}
}
