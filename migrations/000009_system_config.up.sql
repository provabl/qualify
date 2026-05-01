-- System configuration table for caching license validation results and
-- other deployment-level settings. The license validator caches its response
-- here to avoid calling the Provabl licensing server on every startup.

CREATE TABLE IF NOT EXISTS system_config (
    key        TEXT        PRIMARY KEY,
    value      JSONB       NOT NULL,
    expires_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE system_config IS 'Deployment-level key/value config. Used for license cache and feature flags.';
