// SPDX-FileCopyrightText: 2026 Playground Logic LLC
// SPDX-License-Identifier: Apache-2.0

package training

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

// attest:* IAM tag schema — single source of truth for qualify.
//
// qualify writes these tags to researchers' IAM roles on training completion.
// attest's principal resolver (internal/principal/resolver.go) reads them
// to populate Cedar principal attributes during access evaluation.
//
// IMPORTANT: Both repos must agree on these key strings AND on SchemaVersion. The
// canonical contract is attest-tags-schema.json (embedded below, byte-identical with
// attest's pkg/schema/attest-tags-schema.json); the conformance test in
// tags_schema_test.go locks these constants — and moduleTagMap / ModuleExpiryTag —
// to it. If any key changes, update the schema JSON in BOTH repos and bump
// SchemaVersion in the same release.
// See: https://github.com/provabl/qualify/issues/32
//
// Key naming convention:
//
//	attest:<capability>            boolean flag ("true")
//	attest:<capability>-expiry     RFC3339 timestamp of when the flag expires

// SchemaVersion is the version of the attest:* IAM tag contract shared with attest.
// It MUST equal the "version" field of the embedded canonical schema and the same
// constant in attest's pkg/schema/tags.go. Bump it (in both repos, same release)
// whenever a tag key is added, removed, or renamed.
//
// v2 (2026-06): attest:nih-dua-id (string) → attest:nih-dua-ids (set), an
// attest-written key. qualify declares no constant for it (it is not a
// writer=qualify key), but the shared version still bumps in lockstep. See
// provabl ADR 0002 (compute-to-data-access).
const SchemaVersion = 2

// canonicalTagsSchemaJSON is the byte-identical canonical schema, also present in
// attest at pkg/schema/attest-tags-schema.json.
//
//go:embed attest-tags-schema.json
var canonicalTagsSchemaJSON []byte

// TagSchemaEntry is one row of the canonical schema.
type TagSchemaEntry struct {
	Key    string `json:"key"`
	Writer string `json:"writer"` // "qualify" | "attest" | "legacy"
	Type   string `json:"type"`   // "bool" | "timestamp" | "string" | "set"
	Module string `json:"module,omitempty"`
	Expiry string `json:"expiry,omitempty"`
}

// TagSchema is the parsed canonical schema.
type TagSchema struct {
	Version   int              `json:"version"`
	Namespace string           `json:"namespace"`
	Tags      []TagSchemaEntry `json:"tags"`
}

// LoadTagSchema parses the embedded canonical attest:* tag schema.
func LoadTagSchema() (*TagSchema, error) {
	var s TagSchema
	if err := json.Unmarshal(canonicalTagsSchemaJSON, &s); err != nil {
		return nil, fmt.Errorf("parse canonical attest-tags-schema.json: %w", err)
	}
	return &s, nil
}

// Training completion tags (written by svc.CompleteModule):
const (
	TagCUITraining              = "attest:cui-training"
	TagCUITrainingExpiry        = "attest:cui-training-expiry"
	TagHIPAATraining            = "attest:hipaa-training"
	TagHIPAATrainingExpiry      = "attest:hipaa-training-expiry"
	TagAwarenessTraining        = "attest:awareness-training"
	TagAwarenessTrainingExpiry  = "attest:awareness-training-expiry"
	TagFERPATraining            = "attest:ferpa-training"
	TagFERPATrainingExpiry      = "attest:ferpa-training-expiry"
	TagITARTraining             = "attest:itar-training"
	TagITARTrainingExpiry       = "attest:itar-training-expiry"
	TagDataClassTraining        = "attest:data-class-training"
	TagDataClassTrainingExpiry  = "attest:data-class-training-expiry"
	TagResearchSecurityTraining = "attest:research-security-training"
	TagResearchSecurityExpiry   = "attest:research-security-training-expiry"
	TagCOCCheckCurrent          = "attest:coc-check-current"
	TagCOCCheckExpiry           = "attest:coc-check-expiry"
)

// Countries-of-concern check tags (written by svc.RecordCountryCheck):
const (
	TagCountry = "attest:country" // ISO 3166-1 alpha-2 institutional affiliation country
)

// Identity and lab tags (written by svc.SetIdentityTags):
const (
	TagLabID      = "attest:lab-id"
	TagAdminLevel = "attest:admin-level" // "none" | "env" | "sre"
)

// DefaultPassingScore is the percentage a quiz must reach to pass when the caller
// has no per-module threshold of its own. The interactive CLI overrides this with
// the module content's PassingScore; the backend handler, which receives only a
// raw score, uses this default.
const DefaultPassingScore = 80

// moduleTagMap is the authoritative mapping from qualify training module IDs to
// the attest:* IAM tag key written on completion. Unexported to prevent external
// mutation — use TagForModule() and ModuleIDs() for read access.
var moduleTagMap = map[string]string{
	"cui-fundamentals":               TagCUITraining,
	"hipaa-privacy-security":         TagHIPAATraining,
	"security-awareness":             TagAwarenessTraining,
	"ferpa-basics":                   TagFERPATraining,
	"itar-export-control":            TagITARTraining,
	"data-classification":            TagDataClassTraining,
	"nih-research-security":          TagResearchSecurityTraining,
	"countries-of-concern-awareness": TagCOCCheckCurrent,
}

// TagForModule returns the attest:* IAM tag key for a training module ID.
// Returns "" if the module has no tag mapping.
func TagForModule(moduleID string) string {
	return moduleTagMap[moduleID]
}

// ModuleIDs returns all module IDs that have a tag mapping.
func ModuleIDs() []string {
	ids := make([]string, 0, len(moduleTagMap))
	for id := range moduleTagMap {
		ids = append(ids, id)
	}
	return ids
}

// ModuleExpiryTag returns the expiry tag key for a given training tag key,
// or empty string if no expiry tag is defined for that key.
func ModuleExpiryTag(tagKey string) string {
	switch tagKey {
	case TagCUITraining:
		return TagCUITrainingExpiry
	case TagHIPAATraining:
		return TagHIPAATrainingExpiry
	case TagAwarenessTraining:
		return TagAwarenessTrainingExpiry
	case TagFERPATraining:
		return TagFERPATrainingExpiry
	case TagITARTraining:
		return TagITARTrainingExpiry
	case TagDataClassTraining:
		return TagDataClassTrainingExpiry
	case TagResearchSecurityTraining:
		return TagResearchSecurityExpiry
	case TagCOCCheckCurrent:
		return TagCOCCheckExpiry
	}
	return ""
}
