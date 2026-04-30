// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package training

// attest:* IAM tag schema — single source of truth for qualify.
//
// qualify writes these tags to researchers' IAM roles on training completion.
// attest's principal resolver (internal/principal/resolver.go) reads them
// to populate Cedar principal attributes during access evaluation.
//
// Schema version: 1
//
// IMPORTANT: Both repos must agree on these key strings. If any key changes,
// update this file AND attest's principal resolver in the same release.
// See: https://github.com/provabl/qualify/issues/32
//
// Key naming convention:
//
//	attest:<capability>            boolean flag ("true")
//	attest:<capability>-expiry     RFC3339 timestamp of when the flag expires
//
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

// ModuleTagMap is the authoritative mapping from qualify training module IDs to
// the attest:* IAM tag key written on completion. This is the single source of
// truth — do not define this mapping anywhere else.
//
// Both svc.CompleteModule (which writes the tag) and the CLI display (which
// shows the user which tag was written) must reference this map.
var ModuleTagMap = map[string]string{
	"cui-fundamentals":               TagCUITraining,
	"hipaa-privacy-security":         TagHIPAATraining,
	"security-awareness":             TagAwarenessTraining,
	"ferpa-basics":                   TagFERPATraining,
	"itar-export-control":            TagITARTraining,
	"data-classification":            TagDataClassTraining,
	"nih-research-security":          TagResearchSecurityTraining,
	"countries-of-concern-awareness": TagCOCCheckCurrent,
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
