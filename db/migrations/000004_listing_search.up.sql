-- listing_search: task results + COMPLETED status.

ALTER TYPE task_status ADD VALUE IF NOT EXISTS 'COMPLETED';

CREATE TABLE task_items (
    task_id uuid NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    item_id uuid NOT NULL REFERENCES items (id) ON DELETE CASCADE,
    rank integer NOT NULL DEFAULT 0 CHECK (rank >= 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (task_id, item_id)
);

CREATE INDEX task_items_task_id_idx ON task_items (task_id);
