// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

// Package license handles network-based license validation for the qualify
// commercial tier. The validator calls https://licensing.provabl.co at startup
// and caches the result in the system_config table.
//
// If the license endpoint is unreachable and a valid cached result exists,
// the cached result is used. If neither is available, CommunityLicense() is
// returned — the open-source tier with basic modules and no SSO.
package license

import "time"

// Info describes the active license for this deployment.
type Info struct {
	Valid        bool
	Tier         string            // community | professional | enterprise
	ExpiresAt    time.Time
	Company      string
	ContentPacks []string          // content pack IDs included in license
	Features     map[string]bool   // sso, multi_institution, advanced_reporting, etc.
}

// CommunityLicense returns the default open-source license with no key required.
func CommunityLicense() *Info {
	return &Info{
		Valid:        true,
		Tier:         "community",
		ExpiresAt:    time.Now().AddDate(100, 0, 0), // effectively unlimited
		Company:      "",
		ContentPacks: []string{"basic"},
		Features: map[string]bool{
			"sso":                  false,
			"multi_institution":    false,
			"advanced_reporting":   false,
			"custom_content":       false,
		},
	}
}

// HasFeature returns whether the license includes the named feature.
func (i *Info) HasFeature(feature string) bool {
	if i == nil {
		return false
	}
	return i.Features[feature]
}

// HasContentPack returns whether the license includes the named content pack.
func (i *Info) HasContentPack(pack string) bool {
	if i == nil {
		return false
	}
	for _, p := range i.ContentPacks {
		if p == pack || p == "all" {
			return true
		}
	}
	return false
}
