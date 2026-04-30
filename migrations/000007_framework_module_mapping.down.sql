-- SPDX-FileCopyrightText: 2026 Scott Friedman
-- SPDX-License-Identifier: Apache-2.0

DROP INDEX IF EXISTS idx_training_modules_frameworks;
ALTER TABLE training_modules
  DROP COLUMN IF EXISTS required_for_frameworks,
  DROP COLUMN IF EXISTS satisfies_controls;
