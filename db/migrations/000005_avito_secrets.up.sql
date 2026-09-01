-- Avito account passwords for listing_search worker (admin-managed, never exposed via GET).

CREATE TABLE avito_account_secrets (
    account_id uuid PRIMARY KEY REFERENCES avito_accounts (id) ON DELETE CASCADE,
    password text NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now()
);
