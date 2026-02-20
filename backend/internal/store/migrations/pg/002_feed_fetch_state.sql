-- Split runtime fetch metadata from static feed config.
-- Fresh PostgreSQL installs have no legacy columns to migrate.

CREATE TABLE IF NOT EXISTS feed_fetch_state (
	feed_id              BIGINT PRIMARY KEY REFERENCES feeds(id) ON UPDATE CASCADE ON DELETE CASCADE,
	etag                 TEXT NOT NULL DEFAULT '',
	last_modified        TEXT NOT NULL DEFAULT '',
	cache_control        TEXT NOT NULL DEFAULT '',
	expires_at           BIGINT NOT NULL DEFAULT 0,
	last_checked_at      BIGINT NOT NULL DEFAULT 0,
	next_check_at        BIGINT NOT NULL DEFAULT 0,
	last_http_status     BIGINT NOT NULL DEFAULT 0,
	retry_after_until    BIGINT NOT NULL DEFAULT 0,
	last_success_at      BIGINT NOT NULL DEFAULT 0,
	last_error_at        BIGINT NOT NULL DEFAULT 0,
	last_error           TEXT NOT NULL DEFAULT '',
	consecutive_failures BIGINT NOT NULL DEFAULT 0,
	updated_at           BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::bigint
);

-- Insert a default fetch state row for any feeds that don't have one yet
INSERT INTO feed_fetch_state (feed_id, next_check_at, updated_at)
SELECT id, 0, EXTRACT(EPOCH FROM NOW())::bigint
FROM feeds
ON CONFLICT DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_feed_fetch_state_next_check_at ON feed_fetch_state(next_check_at);
