// SPDX-FileCopyrightText: 2026 Playground Logic LLC
// SPDX-License-Identifier: Apache-2.0

package training

import (
	"sort"
	"testing"
)

// declaredTagConstants is every attest:* tag key constant declared in tags.go.
// The conformance test locks this set to the canonical schema, so adding, removing,
// or renaming a constant without updating attest-tags-schema.json (and attest's
// byte-identical copy, same release) fails CI. The guard against the silent
// qualify↔attest drift in qualify#32.
var declaredTagConstants = []string{
	TagCUITraining,
	TagCUITrainingExpiry,
	TagHIPAATraining,
	TagHIPAATrainingExpiry,
	TagAwarenessTraining,
	TagAwarenessTrainingExpiry,
	TagFERPATraining,
	TagFERPATrainingExpiry,
	TagITARTraining,
	TagITARTrainingExpiry,
	TagDataClassTraining,
	TagDataClassTrainingExpiry,
	TagResearchSecurityTraining,
	TagResearchSecurityExpiry,
	TagCOCCheckCurrent,
	TagCOCCheckExpiry,
	TagCountry,
	TagLabID,
	TagAdminLevel,
}

// TestSchemaVersionMatchesCanonical guards that qualify's SchemaVersion constant and
// the embedded canonical schema agree. The same assertion runs in attest against its
// byte-identical copy — together they make SchemaVersion a meaningful cross-repo
// signal: a key change forces a bump that both conformance tests check.
func TestSchemaVersionMatchesCanonical(t *testing.T) {
	s, err := LoadTagSchema()
	if err != nil {
		t.Fatalf("LoadTagSchema: %v", err)
	}
	if s.Version != SchemaVersion {
		t.Errorf("canonical schema version = %d, SchemaVersion const = %d — bump them together", s.Version, SchemaVersion)
	}
	if s.Namespace != "attest:" {
		t.Errorf("schema namespace = %q, want %q", s.Namespace, "attest:")
	}
}

// TestQualifyConstantsMatchSchema asserts qualify declares a constant for every
// attest:* key it is the writer of (writer == "qualify"), and declares no qualify
// constant absent from the schema. (qualify does not declare the attest-written NIH
// keys or the legacy key — those belong to attest's side of the contract.)
func TestQualifyConstantsMatchSchema(t *testing.T) {
	s, err := LoadTagSchema()
	if err != nil {
		t.Fatalf("LoadTagSchema: %v", err)
	}

	qualifyKeys := make(map[string]bool)
	for _, e := range s.Tags {
		if e.Writer == "qualify" {
			qualifyKeys[e.Key] = true
		}
	}
	declared := make(map[string]bool, len(declaredTagConstants))
	for _, k := range declaredTagConstants {
		declared[k] = true
	}

	for k := range qualifyKeys {
		if !declared[k] {
			t.Errorf("schema key %q (writer=qualify) has no qualify constant — add it to tags.go and declaredTagConstants", k)
		}
	}
	for k := range declared {
		if !qualifyKeys[k] {
			t.Errorf("qualify constant %q is not a writer=qualify key in the canonical schema — fix the schema (both repos) and bump SchemaVersion", k)
		}
	}
	if len(declaredTagConstants) != len(qualifyKeys) {
		t.Errorf("count mismatch: %d qualify constants vs %d writer=qualify schema keys\n declared: %v\n schema:   %v",
			len(declaredTagConstants), len(qualifyKeys), sortedKeys(declared), sortedKeys(qualifyKeys))
	}
}

// TestModuleTagMapMatchesSchema asserts the module → tag mapping qualify writes by
// agrees with the schema's "module" metadata, both directions: every schema entry
// with a module maps to it in moduleTagMap, and every moduleTagMap entry is backed
// by a schema "module" field. This is what keeps TagForModule honest.
func TestModuleTagMapMatchesSchema(t *testing.T) {
	s, err := LoadTagSchema()
	if err != nil {
		t.Fatalf("LoadTagSchema: %v", err)
	}

	schemaModuleToKey := make(map[string]string)
	for _, e := range s.Tags {
		if e.Module != "" {
			schemaModuleToKey[e.Module] = e.Key
		}
	}

	for module, wantKey := range schemaModuleToKey {
		if got := TagForModule(module); got != wantKey {
			t.Errorf("TagForModule(%q) = %q, schema says %q", module, got, wantKey)
		}
	}
	for module, key := range moduleTagMap {
		if schemaModuleToKey[module] != key {
			t.Errorf("moduleTagMap[%q] = %q has no matching schema module entry — add a \"module\" field in attest-tags-schema.json (both repos)", module, key)
		}
	}
	if len(moduleTagMap) != len(schemaModuleToKey) {
		t.Errorf("count mismatch: %d moduleTagMap entries vs %d schema entries with a module", len(moduleTagMap), len(schemaModuleToKey))
	}
}

// TestModuleExpiryTagMatchesSchema asserts ModuleExpiryTag agrees with the schema's
// "expiry" metadata for every key that has one. Scoped to writer=qualify keys:
// ModuleExpiryTag is a qualify-side helper and deliberately does not know attest's
// own keys (e.g. the NIH approval pair), so the schema's writer field is the
// authority on which expiry pairings qualify is responsible for.
func TestModuleExpiryTagMatchesSchema(t *testing.T) {
	s, err := LoadTagSchema()
	if err != nil {
		t.Fatalf("LoadTagSchema: %v", err)
	}
	for _, e := range s.Tags {
		if e.Expiry == "" || e.Writer != "qualify" {
			continue
		}
		if got := ModuleExpiryTag(e.Key); got != e.Expiry {
			t.Errorf("ModuleExpiryTag(%q) = %q, schema says %q", e.Key, got, e.Expiry)
		}
	}
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
