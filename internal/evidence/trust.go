// SPDX-FileCopyrightText: 2026 Playground Logic LLC
// SPDX-License-Identifier: Apache-2.0

package qualifyasp

import (
	"crypto/ed25519"
	"fmt"

	"github.com/provabl/evidence/trust"
)

// amKeyID names the attestation-manager key on the Signed evidence node.
const amKeyID = "qualify-training-am-ephemeral"

// EphemeralAM is the kernel's attestation-manager key for one training-completion
// evaluation. It implements BOTH trust.Signer (the SIG built-in) and trust.Store
// (spine verification during appraisal), backed by a freshly generated ed25519
// keypair.
//
// The key is ephemeral on purpose (matching vet). qualify's appraisal is
// in-process and synchronous — Run signs, Appraise verifies, in the same call —
// so the public half that signed the bundle moments earlier is the only key the
// spine check needs. The durable artifact is the attest:* IAM tag set, never the
// evidence bundle, so nothing reads this signature after the call returns. A
// persisted key would add key management for no benefit and would weaken
// freshness, letting a stale bundle verify — exactly what the nonce+SIG spine
// exists to prevent. A persisted/named AM key earns its place only when evidence
// is stored and appraised out-of-process, a non-goal here.
type EphemeralAM struct {
	priv  ed25519.PrivateKey
	pub   ed25519.PublicKey
	keyID string
}

// NewEphemeralAM generates a fresh attestation-manager keypair.
func NewEphemeralAM() (*EphemeralAM, error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, fmt.Errorf("qualify: generate AM key: %w", err)
	}
	return &EphemeralAM{priv: priv, pub: pub, keyID: amKeyID}, nil
}

// Sign implements trust.Signer.
func (a *EphemeralAM) Sign(msg []byte) (sig []byte, keyID string, err error) {
	return ed25519.Sign(a.priv, msg), a.keyID, nil
}

// Verify implements trust.Store: it verifies a signature made by this AM's key.
func (a *EphemeralAM) Verify(keyID string, msg, sig []byte) (bool, error) {
	if keyID != a.keyID {
		return false, nil
	}
	return ed25519.Verify(a.pub, msg, sig), nil
}

// Root implements trust.Store. qualify brings no external trust roots.
func (a *EphemeralAM) Root(string) (trust.Root, bool) { return trust.Root{}, false }
