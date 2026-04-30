# Open-Core Model

qualify is open-source software (Apache 2.0). The core training infrastructure — CLI, backend API, IAM tag integration, and Cedar access gating — is free, self-hostable, and will always remain open source.

The commercial tier, **qualify Cloud**, extends the open-source base with features designed for institutional operations teams managing compliance at scale. Commercial code lives in a separate private repository (`provabl/qualify-cloud`) that imports qualify as a dependency.

---

## What is open source

| Component | Description |
|---|---|
| `qualify` CLI | `qualify train start`, `qualify train status`, `qualify lab record-check`, etc. |
| Training engine | Module storage, quiz scoring, completion records, IAM tag writes |
| Backend API | HTTP API for training progress, dashboard stats, user management |
| IAM tag integration | Writing/reading `attest:*` tags that Cedar PDP evaluates |
| attest integration | The tag schema contract between qualify and attest |
| Local audit log | `~/.qualify/audit.log` JSONL event log |
| Completion certificates | `~/.qualify/certificates/` text certificates |
| Basic web dashboard | Training module list, completion status, S3 gate |
| All migrations | Database schema and seed content structure |
| Docker Compose | Local development environment |
| CI/CD workflows | GitHub Actions build, test, and release pipelines |

**Foundation training modules** (security-awareness, data-classification, CUI, HIPAA, FERPA, ITAR) are included in the open-source migrations as examples. They are community-maintained.

---

## What is commercial (qualify Cloud)

| Feature | Why commercial |
|---|---|
| **Expert-validated training content packs** | Compliance framework packs authored and maintained by domain experts (NIH GDS, CMMC, FedRAMP, HIPAA). Updated when regulations change. Sold per-institution per-year. |
| **Web dashboard — advanced tier** | Multi-user progress tracking, compliance officer views, bulk certificate management, training gap analysis |
| **Multi-institution management** | Manage training compliance across multiple SREs / institutions from a single pane |
| **SSO / LDAP integration** | Institutional identity provider integration (Shibboleth, SAML, LDAP). Syncs groups and attributes automatically |
| **Compliance report generation** | Exportable PDF/Word compliance reports for auditors. Maps training completion to framework controls |
| **Automated expiry management** | Proactive notifications to users and compliance officers before training expires |
| **Custom content portal** | Institution-specific training module authoring and upload UI |
| **White-labeling** | Custom branding for institutional deployments |
| **SLA + support** | Priority issue response, dedicated Slack channel |

---

## The boundary in practice

```
qualify (OSS)                     qualify Cloud (commercial)
────────────────────────────       ──────────────────────────────────
CLI commands                       Advanced web dashboard
Backend API                        SSO / LDAP sync
Training engine                    Expert content packs
Basic web dashboard                Multi-institution management
IAM tag writes                     Compliance report exports
Audit log                          Automated expiry notifications
                                   Custom content portal
                                   White-labeling + SLA
```

An institution can run qualify entirely open-source using the CLI and community content. The commercial tier adds institutional-grade operations tooling and expert-maintained compliance content.

---

## Contributing

Contributions to the open-source core are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md).

If you're building an integration or extension, the qualify backend API is the stable surface to build against. The `attest:*` IAM tag schema is versioned in `internal/training/tags.go`.

For commercial licensing, contact [hello@provabl.dev](mailto:hello@provabl.dev).
