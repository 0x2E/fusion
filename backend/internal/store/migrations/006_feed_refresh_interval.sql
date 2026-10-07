-- Add per-feed refresh interval override.

ALTER TABLE feeds ADD COLUMN refresh_interval_seconds INTEGER DEFAULT NULL;
