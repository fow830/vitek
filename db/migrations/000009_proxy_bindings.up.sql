-- wave B: proxy health + fetch bindings (account↔proxy↔profile).

CREATE TYPE proxy_health_status AS ENUM ('UNKNOWN', 'OK', 'DEGRADED', 'DEAD');
CREATE TYPE listing_binding_status AS ENUM ('ACTIVE', 'PAUSED', 'DISABLED');
CREATE TYPE listing_session_status AS ENUM ('LOGGED_OUT', 'LOGGING_IN', 'READY', 'CHALLENGE', 'ERROR');

ALTER TABLE proxies
    ADD COLUMN last_ok_at timestamptz,
    ADD COLUMN last_err text,
    ADD COLUMN fail_streak integer NOT NULL DEFAULT 0,
    ADD COLUMN health proxy_health_status NOT NULL DEFAULT 'UNKNOWN';

CREATE TABLE listing_fetch_bindings (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    avito_account_id uuid NOT NULL REFERENCES avito_accounts (id) ON DELETE CASCADE,
    proxy_id uuid NOT NULL REFERENCES proxies (id) ON DELETE RESTRICT,
    user_data_dir text NOT NULL,
    status listing_binding_status NOT NULL DEFAULT 'ACTIVE',
    session_status listing_session_status NOT NULL DEFAULT 'LOGGED_OUT',
    session_err text,
    last_session_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX listing_fetch_bindings_account_active_uidx
    ON listing_fetch_bindings (avito_account_id)
    WHERE status = 'ACTIVE';

CREATE UNIQUE INDEX listing_fetch_bindings_proxy_active_uidx
    ON listing_fetch_bindings (proxy_id)
    WHERE status = 'ACTIVE';

CREATE INDEX listing_fetch_bindings_status_idx ON listing_fetch_bindings (status, session_status);
