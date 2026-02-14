-- Split runtime fetch metadata from static feed config.

CREATE TABLE IF NOT EXISTS feed_fetch_state (
	feed_id              INTEGER PRIMARY KEY REFERENCES feeds(id) ON UPDATE CASCADE ON DELETE CASCADE,
	etag                 TEXT NOT NULL DEFAULT '',
	last_modified        TEXT NOT NULL DEFAULT '',
	cache_control        TEXT NOT NULL DEFAULT '',
	expires_at           INTEGER NOT NULL DEFAULT 0,
	last_checked_at      INTEGER NOT NULL DEFAULT 0,
	next_check_at        INTEGER NOT NULL DEFAULT 0,
	last_http_status     INTEGER NOT NULL DEFAULT 0,
	retry_after_until    INTEGER NOT NULL DEFAULT 0,
	last_success_at      INTEGER NOT NULL DEFAULT 0,
	last_error_at        INTEGER NOT NULL DEFAULT 0,
	last_error           TEXT NOT NULL DEFAULT '',
	consecutive_failures INTEGER NOT NULL DEFAULT 0,
	updated_at           INTEGER NOT NULL DEFAULT (unixepoch())
);

INSERT INTO feed_fetch_state (
	feed_id,
	last_success_at,
	last_error_at,
	last_error,
	consecutive_failures,
	last_checked_at,
	next_check_at,
	updated_at
)
SELECT
	id,
	last_build,
	last_failure_at,
	failure,
	failures,
	CASE
		WHEN last_failure_at > last_build THEN last_failure_at
		ELSE last_build
	END,
	0,
	updated_at
FROM feeds;

CREATE INDEX IF NOT EXISTS idx_feed_fetch_state_next_check_at ON feed_fetch_state(next_check_at);

ALTER TABLE feeds DROP COLUMN last_build;
ALTER TABLE feeds DROP COLUMN last_failure_at;
ALTER TABLE feeds DROP COLUMN failure;
ALTER TABLE feeds DROP COLUMN failures;
