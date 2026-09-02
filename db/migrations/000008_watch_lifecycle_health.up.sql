-- listing_search wave A: watch health + plan max_watches.

ALTER TABLE plan_limits
    ADD COLUMN max_watches integer NOT NULL DEFAULT 1;

UPDATE plan_limits SET max_watches = 1 WHERE plan_type = 'FREE';
UPDATE plan_limits SET max_watches = 20 WHERE plan_type = 'PRO';
UPDATE plan_limits SET max_watches = 100 WHERE plan_type = 'ULTRA';

ALTER TABLE plan_limits
    ADD CONSTRAINT plan_limits_max_watches_positive CHECK (max_watches > 0);

ALTER TABLE listing_filter_watches
    ADD COLUMN last_error text,
    ADD COLUMN last_error_at timestamptz,
    ADD COLUMN consecutive_failures integer NOT NULL DEFAULT 0,
    ADD COLUMN last_success_at timestamptz;
