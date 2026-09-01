-- name: GetPlanMaxTasks :one
SELECT max_tasks
FROM plan_limits
WHERE plan_type = $1;

-- name: CreateUser :one
INSERT INTO users (email)
VALUES ($1)
RETURNING id, email, created_at;

-- name: CreateSubscription :one
INSERT INTO subscriptions (user_id, plan_type, active)
VALUES ($1, $2, true)
RETURNING id, user_id, plan_type, active, created_at;

-- name: GetActiveSubscriptionByUser :one
SELECT id, user_id, plan_type, active, created_at
FROM subscriptions
WHERE user_id = $1 AND active = true;

-- name: CountUserTasks :one
SELECT count(*)::bigint AS count
FROM tasks
WHERE user_id = $1
  AND status IN ('PENDING', 'RUNNING', 'PAUSED');

-- name: CreateTask :one
INSERT INTO tasks (user_id, query, status)
VALUES ($1, $2, 'PENDING')
RETURNING id, user_id, query, status, created_at;

-- name: ListActiveProxies :many
SELECT id, endpoint, status, created_at
FROM proxies
WHERE status = 'ACTIVE'
ORDER BY created_at ASC;

-- name: CreateProxy :one
INSERT INTO proxies (endpoint, status)
VALUES ($1, $2)
RETURNING id, endpoint, status, created_at;

-- name: InsertItem :one
INSERT INTO items (avito_id, title)
VALUES ($1, $2)
RETURNING id, avito_id, title, created_at;

-- name: GetItemByAvitoID :one
SELECT id, avito_id, title, created_at
FROM items
WHERE avito_id = $1;
