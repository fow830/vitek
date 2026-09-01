-- name: ClaimNextPendingTask :one
WITH picked AS (
    SELECT id
    FROM tasks
    WHERE status = 'PENDING'
    ORDER BY created_at ASC
    LIMIT 1
    FOR UPDATE SKIP LOCKED
)
UPDATE tasks t
SET status = 'RUNNING'
FROM picked
WHERE t.id = picked.id
RETURNING t.id, t.user_id, t.query, t.status, t.created_at;

-- name: UpdateTaskStatus :one
UPDATE tasks
SET status = $2
WHERE id = $1
RETURNING id, user_id, query, status, created_at;

-- name: GetTask :one
SELECT id, user_id, query, status, created_at
FROM tasks
WHERE id = $1;

-- name: ListTasksByUser :many
SELECT id, user_id, query, status, created_at
FROM tasks
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: InsertTaskItem :exec
INSERT INTO task_items (task_id, item_id, rank)
VALUES ($1, $2, $3);

-- name: ListTaskItems :many
SELECT
    i.id,
    i.avito_id,
    i.title,
    i.created_at,
    ti.rank
FROM task_items ti
JOIN items i ON i.id = ti.item_id
WHERE ti.task_id = $1
ORDER BY ti.rank ASC;

-- name: CountTaskItems :one
SELECT count(*)::bigint AS count
FROM task_items
WHERE task_id = $1;

-- name: ListPriorAvitoIDsForUserQuery :many
SELECT DISTINCT i.avito_id
FROM tasks t
JOIN task_items ti ON ti.task_id = t.id
JOIN items i ON i.id = ti.item_id
WHERE t.user_id = $1
  AND t.query = $2
  AND t.status = 'COMPLETED'
  AND t.id != $3;
