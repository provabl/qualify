// SPDX-FileCopyrightText: 2026 Playground Logic LLC
// SPDX-License-Identifier: Apache-2.0

// Package qualifyasp is qualify's (ASP, appraiser) pair for the provabl/evidence
// kernel. It is the second re-pointing in phase two of ADR 0001 (after vet): a
// training-completion is no longer turned into IAM tags by a mechanical map —
// the verdict is produced by running a Copland term through the kernel's CVM and
// appraising the resulting evidence, then lowering the verdict to the attest:*
// Cedar/IAM attributes the attest principal resolver reads.
//
// The pair lives in qualify (not in evidence/providers) on purpose, matching the
// vet precedent: the kernel's CLAUDE.md says a provider may live in its tool's
// repo, and the no-ASP-branch invariant constrains only the kernel packages.
//
// CONTRACT-DRIFT GUARD: this package never types an "attest:*" key literally.
// The tag keys arrive pre-resolved in Input.TagKey / Input.ExpiryTagKey, sourced
// from internal/training/tags.go (TagForModule / ModuleExpiryTag — the single
// source of truth shared with attest). Because lower.ToAttributes performs no key
// munging, a Claim.Key becomes the IAM tag key verbatim. Keeping the resolution
// in tags.go and only passing strings here means there is no second place a key
// string could drift from attest's resolver.
package qualifyasp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/provabl/evidence/asp"
	"github.com/provabl/evidence/ev"
	"github.com/provabl/evidence/term"
)

// ID keys this pair in the kernel registry. "training" names what is measured
// (a training-completion event), leaving room for future qualify ASPs
// (identity, countries-of-concern) under their own IDs.
const ID term.ASPID = "training"

// TargetScheme is the opaque scheme the kernel routes on; the role ARN — the
// principal the verdict is about — is carried after it. The kernel never parses
// past the scheme.
const TargetScheme = "principal://"

// Target builds the kernel target for a principal's IAM role ARN.
func Target(roleARN string) term.Target { return term.Target(TargetScheme + roleARN) }

// Input is the training-completion event handed to the provider. The service
// constructs one per CompleteModule call and injects it via Provider. TagKey and
// ExpiryTagKey are pre-resolved from internal/training/tags.go; ExpiresAt is
// precomputed (RFC3339) by the service so the appraiser stays a pure function.
type Input struct {
	ModuleID     string `json:"module_id"`
	TagKey       string `json:"tag_key"`        // attest:* key, or "" if the module has no mapping
	ExpiryTagKey string `json:"expiry_tag_key"` // attest:*-expiry key, or "" if none
	Score        int    `json:"score"`
	PassingScore int    `json:"passing_score"`
	ExpiresAt    string `json:"expires_at"` // RFC3339
}

// Provider assembles the training pair from an injected completion event.
func Provider(in Input) asp.Provider {
	return asp.Provider{
		ID:        ID,
		Measurer:  measurer{in: in},
		Appraiser: appraiser{},
	}
}

// --- measurer: gather, do not judge -----------------------------------------

type measurer struct{ in Input }

func (m measurer) Measure(_ context.Context, _ asp.MeasureIn) (ev.Measurement, error) {
	if m.in.TagKey == "" {
		// The module has no attest tag mapping — there is nothing to measure.
		// This is the kernel-native form of the old "if !ok { return nil }": a
		// NotApplicable measurement, distinct from a failed one.
		return ev.Measurement{
			Status: ev.NotApplicable,
			Detail: fmt.Sprintf("module %q has no attest tag mapping", m.in.ModuleID),
		}, nil
	}
	payload, err := json.Marshal(m.in)
	if err != nil {
		return ev.Measurement{}, fmt.Errorf("qualify: marshal completion: %w", err)
	}
	return ev.Measurement{Payload: payload, Status: ev.Collected}, nil
}

// --- appraiser: decode, judge pass/fail, emit attest:* claims ----------------

type appraiser struct{}

func (appraiser) Appraise(_ context.Context, in asp.AppraiseIn) (asp.Verdict, error) {
	var p Input
	if err := json.Unmarshal(in.Meas.Payload, &p); err != nil {
		return asp.Verdict{}, fmt.Errorf("qualify: decode completion: %w", err)
	}

	if p.Score < p.PassingScore {
		// Failed: emit NO tag claims (distinct from NotApplicable — the module
		// was attempted, just not passed). No attest:* tag is written.
		return asp.Verdict{
			Pass:   false,
			Reason: fmt.Sprintf("module %s score %d below passing %d", p.ModuleID, p.Score, p.PassingScore),
		}, nil
	}

	claims := []asp.Claim{{Key: p.TagKey, Value: "true", Type: "bool"}}
	if p.ExpiryTagKey != "" {
		claims = append(claims, asp.Claim{Key: p.ExpiryTagKey, Value: p.ExpiresAt, Type: "string"})
	}
	return asp.Verdict{
		Pass:   true,
		Claims: claims,
		Reason: fmt.Sprintf("module %s completed (score %d >= %d)", p.ModuleID, p.Score, p.PassingScore),
	}, nil
}

// IsAttestTag reports whether a lowered attribute key is an attest:* IAM tag (as
// opposed to a synthetic kernel marker like "attested" or "training.applicable").
// Exposed so the single tag-writing chokepoint in the training service can filter
// on it without re-deriving the namespace.
func IsAttestTag(key string) bool { return strings.HasPrefix(key, "attest:") }

// --- attributes provider: identity & countries-of-concern -------------------
//
// SetIdentityTags (attest:lab-id, attest:admin-level) and RecordCountryCheck
// (attest:country, attest:coc-check-current, attest:coc-check-expiry) are direct
// attestations of recorded facts — there is no quiz to pass. They route through
// the kernel for the SAME reason training does (one lowering path; a freshness-
// bound, appraised assertion rather than a hand-built tag set), but the appraiser
// always passes: the verdict is "these recorded attributes are attested as of
// now", and the claims are the attest:* tags to write.

// AttrID keys the attributes pair in the registry. Distinct from the training
// ID so a single registry could hold both, and so the kernel's synthetic markers
// read "attrs.*" rather than "training.*".
const AttrID term.ASPID = "attrs"

// Tag is a pre-resolved attest:* attribute to attest. Key is the full IAM tag key
// (e.g. "attest:lab-id"), sourced from the caller (tags.go constants) so this
// package never types an attest:* literal — same drift guard as the training pair.
type Tag struct {
	Key   string
	Value string
}

// AttributesInput is the set of attest:* tags to attest for a principal.
type AttributesInput struct {
	Tags []Tag `json:"tags"`
}

// AttributesProvider assembles the attributes pair from an injected tag set.
func AttributesProvider(in AttributesInput) asp.Provider {
	return asp.Provider{
		ID:        AttrID,
		Measurer:  attrMeasurer{in: in},
		Appraiser: attrAppraiser{},
	}
}

type attrMeasurer struct{ in AttributesInput }

func (m attrMeasurer) Measure(_ context.Context, _ asp.MeasureIn) (ev.Measurement, error) {
	if len(m.in.Tags) == 0 {
		// Nothing to attest — the kernel-native form of "no tags to write".
		return ev.Measurement{Status: ev.NotApplicable, Detail: "no attributes to attest"}, nil
	}
	payload, err := json.Marshal(m.in)
	if err != nil {
		return ev.Measurement{}, fmt.Errorf("qualify: marshal attributes: %w", err)
	}
	return ev.Measurement{Payload: payload, Status: ev.Collected}, nil
}

type attrAppraiser struct{}

func (attrAppraiser) Appraise(_ context.Context, in asp.AppraiseIn) (asp.Verdict, error) {
	var p AttributesInput
	if err := json.Unmarshal(in.Meas.Payload, &p); err != nil {
		return asp.Verdict{}, fmt.Errorf("qualify: decode attributes: %w", err)
	}
	// No pass/fail: these are recorded facts being attested. Emit each as a claim;
	// types default to "string" in lowering, which is correct for lab-id, country,
	// admin-level, RFC3339 expiry; coc-check-current is the string "true".
	claims := make([]asp.Claim, 0, len(p.Tags))
	for _, t := range p.Tags {
		claims = append(claims, asp.Claim{Key: t.Key, Value: t.Value, Type: "string"})
	}
	return asp.Verdict{Pass: true, Claims: claims, Reason: "attributes attested"}, nil
}
