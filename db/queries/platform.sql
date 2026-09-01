-- name: CreateMagicLinkChallenge :one
INSERT INTO magic_link_challenges (email, token_hash, role_hint, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, email, token_hash, role_hint, expires_at, consumed_at, created_at;

-- name: ConsumeMagicLinkChallenge :one
UPDATE magic_link_challenges
SET consumed_at = now()
WHERE token_hash = $1
  AND consumed_at IS NULL
  AND expires_at > now()
RETURNING id, email, token_hash, role_hint, expires_at, consumed_at, created_at;

-- name: ListShippedProductServices :many
SELECT code, title, shipped, created_at
FROM product_services
WHERE shipped = true
ORDER BY code ASC;

-- name: GetProductService :one
SELECT code, title, shipped, created_at
FROM product_services
WHERE code = $1;

-- name: GrantUserService :one
INSERT INTO user_service_entitlements (user_id, service_code)
VALUES ($1, $2)
RETURNING user_id, service_code, created_at;

-- name: HasUserService :one
SELECT EXISTS (
    SELECT 1
    FROM user_service_entitlements
    WHERE user_id = $1 AND service_code = $2
);

-- name: ListUserServices :many
SELECT e.user_id, e.service_code, e.created_at
FROM user_service_entitlements e
WHERE e.user_id = $1
ORDER BY e.service_code ASC;

-- name: CreateAvitoAccount :one
INSERT INTO avito_accounts (label, status, external_ref)
VALUES ($1, $2, $3)
RETURNING id, label, status, external_ref, created_at, updated_at;

-- name: CountAvitoAccounts :one
SELECT count(*)::bigint AS count
FROM avito_accounts;

-- name: CreateUserWithRole :one
INSERT INTO users (email, role)
VALUES ($1, $2)
RETURNING id, email, created_at, role;
