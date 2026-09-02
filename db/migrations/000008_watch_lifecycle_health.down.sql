ALTER TABLE listing_filter_watches
    DROP COLUMN IF EXISTS last_success_at,
    DROP COLUMN IF EXISTS consecutive_failures,
    DROP COLUMN IF EXISTS last_error_at,
    DROP COLUMN IF EXISTS last_error;

ALTER TABLE plan_limits
    DROP CONSTRAINT IF EXISTS plan_limits_max_watches_positive;

ALTER TABLE plan_limits
    DROP COLUMN IF EXISTS max_watches;
