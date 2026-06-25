# qualify — required AWS permissions

`qualify preflight` verifies the calling AWS principal holds these actions, using
read-only `iam:SimulatePrincipalPolicy` against the caller ARN (from
`sts:GetCallerIdentity`). It **evaluates, it never acts** — running preflight changes
nothing. A denied action prints a remediation and the command exits non-zero.

Most of qualify's flow needs **no AWS at all**: `train start`, `train status`, and
`train certificate` only read and write the local `~/.qualify/` store (progress,
certificates, audit log) and talk to the qualify backend over HTTP. The one
AWS-touching path is the IAM-tag write that records a completed module — the action
below.

| Action | Needed by | Status |
|--------|-----------|--------|
| `sts:GetCallerIdentity` | preflight itself (resolves the caller ARN to simulate) | live |
| `iam:SimulatePrincipalPolicy` | preflight itself (the permission self-check) | live |
| `iam:TagRole` | qualify's core write — on training-module completion qualify tags the researcher's IAM role with `attest:<training-id> = true` (and a companion `attest:<training-id>-expiry`) | live |

## The `attest:*` tag write

`iam:TagRole` is the whole reason qualify touches AWS. When a researcher passes a
module, `svc.CompleteModule()` writes IAM role tags in the `attest:*` namespace
(`attest:cui-training`, `attest:coc-check-current`, …) onto the role registered via
`qualify lab register-role`. attest's Cedar PDP reads those tags as `principal.*`
attributes on the next access request; until the tag is present and unexpired, access
to the gated environment is denied.

The namespace is **`attest:*`, not `qualify:*`** — it is attest's integration
interface, the lowered-attribute contract the PDP consumes, not a qualify-private
key space. The schema is versioned and shared byte-for-byte with attest
(`internal/training/attest-tags-schema.json`); see the README's "Integration with
attest" section and attest's `docs/integrations/qualify.md`.

## Why preflight is its own action list

This check mirrors attest's and ground's caller-permission preflight (provabl#16).
The suite tools are deliberately decoupled — the evidence kernel is the only shared
dependency, and it is stdlib-only — so each tool carries its own small copy of the
generic check rather than depending on a shared AWS-SDK library. To scope a principal
to qualify's live paths, grant `sts:GetCallerIdentity` +
`iam:SimulatePrincipalPolicy` (preflight) and `iam:TagRole` (the completion write);
the local training flow needs nothing else.

## Boundary

qualify **trains the person and records their approvals; it never decides access**
(attest does, via the Cedar PDP reading the `attest:*` tags qualify writes) and never
moves or governs data (steward does). The IAM tag is qualify's only durable AWS
output — qualify produces the evidence, attest consumes it. See the README boundary
note and the provabl suite glossary.
