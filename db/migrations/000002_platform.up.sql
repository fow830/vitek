-- +goose Up
-- Platform: Magic Link auth, multi-service catalog, Avito accounts, admin-managed surface.

CREATE TYPE user_role AS ENUM ('USER', 'ADMIN');
CREATE TYPE avito_account_status AS ENUM ('ACTIVE', 'DISABLED', 'ERROR');

ALTER TABLE users
    ADD COLUMN role user_role NOT NULL DEFAULT 'USER';

ALTER TABLE proxies
    ADD COLUMN label text NOT NULL DEFAULT '';

-- Magic Link challenges for both USER and ADMIN (no passwords).
CREATE TABLE magic_link_challenges (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL,
    token_hash text NOT NULL UNIQUE,
    role_hint user_role NOT NULL DEFAULT 'USER',
    expires_at timestamptz NOT NULL,
    consumed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX magic_link_challenges_email_idx ON magic_link_challenges (email);

-- Product services: listing_search is shipped; listing_warmup reserved (not shipped).
CREATE TABLE product_services (
    code text PRIMARY KEY,
    title text NOT NULL,
    shipped boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO product_services (code, title, shipped) VALUES
    ('listing_search', 'Similar listings search', true),
    ('listing_warmup', 'Listing warmup (clicks)', false);

CREATE TABLE user_service_entitlements (
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    service_code text NOT NULL REFERENCES product_services (code),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, service_code)
);

CREATE INDEX user_service_entitlements_service_idx ON user_service_entitlements (service_code);

-- Dozens/hundreds of Avito accounts — managed in admin UI.
CREATE TABLE avito_accounts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    label text NOT NULL,
    status avito_account_status NOT NULL DEFAULT 'DISABLED',
    external_ref text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX avito_accounts_status_idx ON avito_accounts (status);
