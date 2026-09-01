-- name: GetUserByEmail :one
SELECT id, email, created_at, role
FROM users
WHERE email = $1;

-- name: CreateSession :one
INSERT INTO sessions (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, token_hash, expires_at, revoked_at, created_at;

-- name: GetActiveSessionByHash :one
SELECT
    s.id,
    s.user_id,
    s.token_hash,
    s.expires_at,
    s.revoked_at,
    s.created_at,
    u.email,
    u.role
FROM sessions s
JOIN users u ON u.id = s.user_id
WHERE s.token_hash = $1
  AND s.revoked_at IS NULL
  AND s.expires_at > now();

-- name: RevokeSessionByHash :exec
UPDATE sessions
SET revoked_at = now()
WHERE token_hash = $1
  AND revoked_at IS NULL;

-- name: ListAllProxies :many
SELECT id, endpoint, status, created_at, label
FROM proxies
ORDER BY created_at ASC;

-- name: UpdateProxy :one
UPDATE proxies
SET endpoint = $2,
    status = $3,
    label = $4
WHERE id = $1
RETURNING id, endpoint, status, created_at, label;

-- name: ListAvitoAccounts :many
SELECT id, label, status, external_ref, created_at, updated_at
FROM avito_accounts
ORDER BY created_at ASC;

-- name: UpdateAvitoAccount :one
UPDATE avito_accounts
SET label = $2,
    status = $3,
    external_ref = $4,
    updated_at = now()
WHERE id = $1
RETURNING id, label, status, external_ref, created_at, updated_at;

-- name: UpsertAvitoAccountSecret :exec
INSERT INTO avito_account_secrets (account_id, password)
VALUES ($1, $2)
ON CONFLICT (account_id) DO UPDATE
SET password = EXCLUDED.password,
    updated_at = now();

-- name: PickActiveAvitoAccount :one
SELECT a.id, a.label, a.external_ref, s.password
FROM avito_accounts a
JOIN avito_account_secrets s ON s.account_id = a.id
WHERE a.status = 'ACTIVE'
ORDER BY a.created_at ASC
LIMIT 1;
