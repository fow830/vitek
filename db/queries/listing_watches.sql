-- name: UpsertFilterWatch :one
INSERT INTO listing_filter_watches (user_id, filter_key, query, status)
VALUES ($1, $2, $3, 'ACTIVE')
ON CONFLICT (user_id, filter_key) DO UPDATE
SET query = EXCLUDED.query,
    status = 'ACTIVE',
    last_polled_at = NULL
RETURNING id, user_id, filter_key, query, status, last_polled_at, created_at;

-- name: GetFilterWatch :one
SELECT id, user_id, filter_key, query, status, last_polled_at, created_at
FROM listing_filter_watches
WHERE id = $1;

-- name: GetFilterWatchForUser :one
SELECT id, user_id, filter_key, query, status, last_polled_at, created_at
FROM listing_filter_watches
WHERE id = $1
  AND user_id = $2;

-- name: ListDueFilterWatches :many
SELECT id, user_id, filter_key, query, status, last_polled_at, created_at
FROM listing_filter_watches
WHERE status = 'ACTIVE'
  AND (
    last_polled_at IS NULL
    OR last_polled_at <= now() - interval '1 minute'
  )
ORDER BY last_polled_at NULLS FIRST, created_at ASC
LIMIT 20
FOR UPDATE SKIP LOCKED;

-- name: TouchFilterWatchPolled :exec
UPDATE listing_filter_watches
SET last_polled_at = now()
WHERE id = $1;

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
