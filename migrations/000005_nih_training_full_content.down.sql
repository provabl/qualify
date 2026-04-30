-- SPDX-FileCopyrightText: 2026 Scott Friedman
-- SPDX-License-Identifier: Apache-2.0

-- Revert nih-research-security to stub content (restores migration 000004 state).
UPDATE training_modules SET
  title = 'NIH Research Security (NOT-OD-26-017)',
  description = 'Required by NIH NOT-OD-26-017 for all key personnel on NIH-funded projects. Covers foreign influence disclosure, research integrity, and institutional reporting obligations.',
  estimated_minutes = 45,
  content = '{"sections": [{"id": "intro", "title": "NIH Research Security Requirements", "type": "text", "content": "NIH Notice NOT-OD-26-017 requires all key personnel on NIH-funded awards to complete research security training within 12 months of award issuance and every 12 months thereafter."}, {"id": "disclosure", "title": "Foreign Affiliation Disclosure", "type": "text", "content": "Key personnel must disclose all foreign affiliations, positions, and financial interests. Failure to disclose is research misconduct."}, {"id": "reporting", "title": "Reporting Obligations", "type": "text", "content": "You are required to report to your institution any approaches from foreign entities seeking inappropriate access to your research."}]}'
WHERE name = 'nih-research-security';
