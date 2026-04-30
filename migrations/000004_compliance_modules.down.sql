-- Copyright 2026 Scott Friedman. Licensed under the Apache License, Version 2.0.

DELETE FROM training_modules WHERE name IN (
  'cui-fundamentals',
  'hipaa-privacy-security',
  'security-awareness',
  'ferpa-basics',
  'itar-export-control',
  'data-classification',
  'nih-research-security'
);
