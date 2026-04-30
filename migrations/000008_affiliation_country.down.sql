ALTER TABLE users
    DROP COLUMN IF EXISTS institutional_affiliation_country,
    DROP COLUMN IF EXISTS affiliation_check_performed_at,
    DROP COLUMN IF EXISTS affiliation_check_performed_by;
