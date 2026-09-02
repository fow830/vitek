-- name: GetProxy :one
SELECT id, endpoint, status, created_at, label, last_ok_at, last_err, fail_streak, health
FROM proxies
WHERE id = $1;

-- name: RecordProxyHealthOK :exec
UPDATE proxies
SET health = 'OK',
    last_ok_at = now(),
    last_err = NULL,
    fail_streak = 0
WHERE id = $1;

-- name: RecordProxyHealthFail :one
UPDATE proxies
SET last_err = sqlc.arg(last_err),
    fail_streak = fail_streak + 1,
    health = CASE
        WHEN fail_streak + 1 >= sqlc.arg(dead_after)::int THEN 'DEAD'::proxy_health_status
        ELSE 'DEGRADED'::proxy_health_status
    END
WHERE id = sqlc.arg(id)
RETURNING id, endpoint, status, created_at, label, last_ok_at, last_err, fail_streak, health;

-- name: CreateBinding :one
INSERT INTO listing_fetch_bindings (avito_account_id, proxy_id, user_data_dir, status, session_status)
VALUES ($1, $2, $3, 'ACTIVE', 'LOGGED_OUT')
RETURNING id, avito_account_id, proxy_id, user_data_dir, status, session_status, session_err, last_session_at, created_at;

-- name: ListBindings :many
SELECT id, avito_account_id, proxy_id, user_data_dir, status, session_status, session_err, last_session_at, created_at
FROM listing_fetch_bindings
ORDER BY created_at ASC;

-- name: GetBinding :one
SELECT id, avito_account_id, proxy_id, user_data_dir, status, session_status, session_err, last_session_at, created_at
FROM listing_fetch_bindings
WHERE id = $1;

-- name: UpdateBindingStatus :one
UPDATE listing_fetch_bindings
SET status = $2
WHERE id = $1
RETURNING id, avito_account_id, proxy_id, user_data_dir, status, session_status, session_err, last_session_at, created_at;

-- name: UpdateBindingSession :one
UPDATE listing_fetch_bindings
SET session_status = $2,
    session_err = $3,
    last_session_at = now()
WHERE id = $1
RETURNING id, avito_account_id, proxy_id, user_data_dir, status, session_status, session_err, last_session_at, created_at;

-- name: PauseBindingsForProxy :exec
UPDATE listing_fetch_bindings
SET status = 'PAUSED'
WHERE proxy_id = $1
  AND status = 'ACTIVE';

-- name: PickReadyBinding :one
SELECT
    b.id,
    b.avito_account_id,
    b.proxy_id,
    b.user_data_dir,
    b.status,
    b.session_status,
    b.session_err,
    b.last_session_at,
    b.created_at,
    p.endpoint AS proxy_endpoint
FROM listing_fetch_bindings b
JOIN proxies p ON p.id = b.proxy_id
WHERE b.status = 'ACTIVE'
  AND b.session_status = 'READY'
  AND p.status = 'ACTIVE'
  AND p.health <> 'DEAD'
ORDER BY b.created_at ASC
LIMIT 1;
