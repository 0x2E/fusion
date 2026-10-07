-- Add per-feed refresh interval override.

ALTER TABLE feeds ADD COLUMN refresh_interval_seconds INTEGER DEFAULT NULL;

-- The scheduler derives its sleep from MIN(next_check_at); a legacy zero
-- (feeds never pulled since the 002 backfill) would pin the minimum to
-- "due immediately" forever and mask the real earliest due time.
UPDATE feed_fetch_state SET next_check_at = unixepoch(), updated_at = unixepoch()
WHERE next_check_at = 0;
