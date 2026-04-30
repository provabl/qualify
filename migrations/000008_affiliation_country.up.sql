-- Add institutional affiliation country tracking to users.
--
-- Compliance officers perform a countries-of-concern check per NOT-OD-25-083
-- and record the researcher's institutional affiliation country here.
-- qualify writes attest:country to the researcher's IAM role after this check.
-- attest's Cedar PDP evaluates principal.InstitutionalAffiliationCountry for
-- NIH GDS access control and ITAR deemed-export enforcement.

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS institutional_affiliation_country CHAR(2),
    ADD COLUMN IF NOT EXISTS affiliation_check_performed_at    TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS affiliation_check_performed_by    TEXT;

COMMENT ON COLUMN users.institutional_affiliation_country IS 'ISO 3166-1 alpha-2 country code of the researcher''s primary institutional affiliation';
COMMENT ON COLUMN users.affiliation_check_performed_at    IS 'When the countries-of-concern check was performed';
COMMENT ON COLUMN users.affiliation_check_performed_by    IS 'Who performed the countries-of-concern check (compliance officer user ID)';
