DROP TABLE IF EXISTS listing_fetch_bindings;
DROP TYPE IF EXISTS listing_session_status;
DROP TYPE IF EXISTS listing_binding_status;

ALTER TABLE proxies
    DROP COLUMN IF EXISTS health,
    DROP COLUMN IF EXISTS fail_streak,
    DROP COLUMN IF EXISTS last_err,
    DROP COLUMN IF EXISTS last_ok_at;

DROP TYPE IF EXISTS proxy_health_status;
