-- name: UpsertFilterWatch :one
INSERT INTO listing_filter_watches (user_id, filter_key, query, status, meta_status)
VALUES ($1, $2, $3, 'ACTIVE', 'PENDING')
ON CONFLICT (user_id, filter_key) DO UPDATE
SET query = EXCLUDED.query,
    status = 'ACTIVE',
    meta_status = CASE
        WHEN listing_filter_watches.query IS DISTINCT FROM EXCLUDED.query THEN 'PENDING'::listing_watch_meta_status
        ELSE listing_filter_watches.meta_status
    END
RETURNING id, user_id, filter_key, query, status, last_polled_at, created_at,
    last_error, last_error_at, consecutive_failures, last_success_at, meta_status, meta_json;

-- name: ListFilterWatchesByUser :many
SELECT id, user_id, filter_key, query, status, last_polled_at, created_at,
    last_error, last_error_at, consecutive_failures, last_success_at, meta_status, meta_json
FROM listing_filter_watches
WHERE user_id = $1
  AND status IN ('ACTIVE', 'PAUSED')
ORDER BY created_at DESC;

-- name: GetFilterWatch :one
SELECT id, user_id, filter_key, query, status, last_polled_at, created_at,
    last_error, last_error_at, consecutive_failures, last_success_at, meta_status, meta_json
FROM listing_filter_watches
WHERE id = $1;

-- name: GetFilterWatchForUser :one
SELECT id, user_id, filter_key, query, status, last_polled_at, created_at,
    last_error, last_error_at, consecutive_failures, last_success_at, meta_status, meta_json
FROM listing_filter_watches
WHERE id = $1
  AND user_id = $2;

-- name: GetFilterWatchByUserFilter :one
SELECT id, user_id, filter_key, query, status, last_polled_at, created_at,
    last_error, last_error_at, consecutive_failures, last_success_at, meta_status, meta_json
FROM listing_filter_watches
WHERE user_id = $1
  AND filter_key = $2;

-- name: CountUserWatches :one
SELECT count(*)::bigint AS count
FROM listing_filter_watches
WHERE user_id = $1
  AND status IN ('ACTIVE', 'PAUSED');

-- name: UpdateFilterWatchStatus :one
UPDATE listing_filter_watches
SET status = $2
WHERE id = $1
  AND user_id = $3
RETURNING id, user_id, filter_key, query, status, last_polled_at, created_at,
    last_error, last_error_at, consecutive_failures, last_success_at, meta_status, meta_json;

-- name: UpdateFilterWatchMeta :exec
UPDATE listing_filter_watches
SET meta_status = $2,
    meta_json = $3
WHERE id = $1;

-- name: ListDueFilterWatches :many
SELECT id, user_id, filter_key, query, status, last_polled_at, created_at,
    last_error, last_error_at, consecutive_failures, last_success_at, meta_status, meta_json
FROM listing_filter_watches
WHERE status = 'ACTIVE'
  AND (
    last_polled_at IS NULL
    OR last_polled_at <= now() - interval '1 minute'
  )
ORDER BY last_polled_at NULLS FIRST, created_at ASC
LIMIT 20;

-- name: TouchFilterWatchPolled :exec
UPDATE listing_filter_watches
SET last_polled_at = now(),
    last_success_at = now(),
    last_error = NULL,
    last_error_at = NULL,
    consecutive_failures = 0
WHERE id = $1;

-- name: RecordFilterWatchPollFailure :one
UPDATE listing_filter_watches
SET last_error = sqlc.arg(last_error),
    last_error_at = now(),
    consecutive_failures = consecutive_failures + 1,
    meta_status = CASE
        WHEN meta_status = 'PENDING' THEN 'FAILED'::listing_watch_meta_status
        ELSE meta_status
    END,
    status = CASE
        WHEN consecutive_failures + 1 >= sqlc.arg(max_failures)::int THEN 'PAUSED'::listing_watch_status
        ELSE status
    END
WHERE id = sqlc.arg(id)
RETURNING id, user_id, filter_key, query, status, last_polled_at, created_at,
    last_error, last_error_at, consecutive_failures, last_success_at, meta_status, meta_json;

-- name: InsertWatchHit :exec
INSERT INTO listing_watch_hits (watch_id, item_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: ListWatchHits :many
SELECT
    i.id,
    i.avito_id,
    i.title,
    i.created_at,
    wh.found_at
FROM listing_watch_hits wh
JOIN items i ON i.id = wh.item_id
WHERE wh.watch_id = $1
ORDER BY wh.found_at DESC;

-- name: DeleteFilterSeenForUserFilter :exec
DELETE FROM listing_filter_seen
WHERE user_id = $1
  AND filter_key = $2;

-- name: ClearWatchHits :exec
DELETE FROM listing_watch_hits
WHERE watch_id = $1;

-- name: ResetFilterWatchBaseline :exec
UPDATE listing_filter_watches
SET last_polled_at = NULL
WHERE id = $1;
