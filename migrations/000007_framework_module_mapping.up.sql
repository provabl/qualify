-- SPDX-FileCopyrightText: 2026 Scott Friedman
-- SPDX-License-Identifier: Apache-2.0

-- Add framework_id and control_id columns to training_modules.
-- Enables qualify to surface required modules based on active attest frameworks.
-- Maps the qualify training pipeline to specific attest framework controls.

ALTER TABLE training_modules
  ADD COLUMN IF NOT EXISTS required_for_frameworks JSONB NOT NULL DEFAULT '[]',
  ADD COLUMN IF NOT EXISTS satisfies_controls JSONB NOT NULL DEFAULT '[]';

-- Map each compliance module to the frameworks and controls it satisfies.

UPDATE training_modules SET
  required_for_frameworks = '["nist-800-171-r2", "cmmc-level-1", "cmmc-level-2", "cmmc-level-3"]',
  satisfies_controls = '[{"framework": "nist-800-171-r2", "control_id": "3.2.1", "title": "Limit system access to authorized users"},
                          {"framework": "nist-800-171-r2", "control_id": "3.2.2", "title": "Limit system access — CUI handling"},
                          {"framework": "cmmc-level-2", "control_id": "AC.L2-3.1.1", "title": "Authorized Access Control"}]'
WHERE name = 'cui-fundamentals';

UPDATE training_modules SET
  required_for_frameworks = '["hipaa"]',
  satisfies_controls = '[{"framework": "hipaa", "control_id": "164.308(a)(5)", "title": "Security awareness and training"},
                          {"framework": "hipaa", "control_id": "164.312(a)(2)(i)", "title": "Unique user identification"}]'
WHERE name = 'hipaa-privacy-security';

UPDATE training_modules SET
  required_for_frameworks = '["nist-800-171-r2", "cmmc-level-1", "cmmc-level-2", "hipaa", "ferpa", "itar", "nih-gds", "gdpr"]',
  satisfies_controls = '[{"framework": "nist-800-171-r2", "control_id": "3.2.2", "title": "Ensure CUI users are aware of security risks"},
                          {"framework": "cmmc-level-1", "control_id": "AC.L1-3.1.1", "title": "Authorized Access Control"},
                          {"framework": "hipaa", "control_id": "164.308(a)(5)(ii)(A)", "title": "Security reminders"}]'
WHERE name = 'security-awareness';

UPDATE training_modules SET
  required_for_frameworks = '["ferpa"]',
  satisfies_controls = '[{"framework": "ferpa", "control_id": "ferpa-access-control", "title": "Appropriate access to education records"}]'
WHERE name = 'ferpa-basics';

UPDATE training_modules SET
  required_for_frameworks = '["itar"]',
  satisfies_controls = '[{"framework": "itar", "control_id": "itar-workforce-training", "title": "Export control workforce training"}]'
WHERE name = 'itar-export-control';

UPDATE training_modules SET
  required_for_frameworks = '["nist-800-171-r2", "cmmc-level-2", "hipaa", "nih-gds"]',
  satisfies_controls = '[{"framework": "nist-800-171-r2", "control_id": "3.4.1", "title": "Baseline configurations"},
                          {"framework": "hipaa", "control_id": "164.308(a)(5)(ii)(B)", "title": "Protection from malicious software"}]'
WHERE name = 'data-classification';

UPDATE training_modules SET
  required_for_frameworks = '["nih-gds", "nist-800-171-r2"]',
  satisfies_controls = '[{"framework": "nih-gds", "control_id": "nih-gds-3.1", "title": "DUA administration — key personnel training"},
                          {"framework": "nist-800-171-r2", "control_id": "3.2.1", "title": "Awareness and training"}]'
WHERE name = 'nih-research-security';

UPDATE training_modules SET
  required_for_frameworks = '["nih-gds"]',
  satisfies_controls = '[{"framework": "nih-gds", "control_id": "nih-gds-1.2", "title": "Countries of concern — institutional affiliation check"}]'
WHERE name = 'countries-of-concern-awareness';

-- Index for fast framework lookup
CREATE INDEX IF NOT EXISTS idx_training_modules_frameworks
  ON training_modules USING GIN (required_for_frameworks);
