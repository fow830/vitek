-- golang-migrate style applied by suite (exec *.up.sql in order)

CREATE TYPE plan_type AS ENUM ('FREE', 'PRO', 'ULTRA');
CREATE TYPE task_status AS ENUM ('PENDING', 'RUNNING', 'PAUSED', 'FAILED');
CREATE TYPE proxy_status AS ENUM ('ACTIVE', 'BANNED', 'DISABLED');

CREATE TABLE plan_limits (
    plan_type plan_type PRIMARY KEY,
    max_tasks integer NOT NULL CHECK (max_tasks > 0)
);

INSERT INTO plan_limits (plan_type, max_tasks) VALUES
    ('FREE', 1),
    ('PRO', 20),
    ('ULTRA', 100);

CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE subscriptions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    plan_type plan_type NOT NULL,
    active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id)
);

CREATE TABLE tasks (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    query text NOT NULL,
    status task_status NOT NULL DEFAULT 'PENDING',
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX tasks_user_id_idx ON tasks (user_id);
CREATE INDEX tasks_user_active_idx ON tasks (user_id)
    WHERE status IN ('PENDING', 'RUNNING', 'PAUSED');

CREATE TABLE proxies (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    endpoint text NOT NULL UNIQUE,
    status proxy_status NOT NULL DEFAULT 'ACTIVE',
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX proxies_status_idx ON proxies (status);

CREATE TABLE items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    avito_id text NOT NULL UNIQUE,
    title text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now()
);
