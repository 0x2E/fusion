-- Initial schema (PostgreSQL)
-- Uses BIGSERIAL instead of INTEGER PRIMARY KEY, BOOLEAN instead of INTEGER for flags,
-- and a tsvector generated column for full-text search instead of SQLite FTS5.

CREATE TABLE IF NOT EXISTS groups (
	id         BIGSERIAL PRIMARY KEY,
	name       TEXT NOT NULL UNIQUE,
	created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::bigint,
	updated_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::bigint
);

INSERT INTO groups (id, name) VALUES (1, 'Default') ON CONFLICT DO NOTHING;
-- Advance the sequence past the explicitly inserted id=1
SELECT setval(pg_get_serial_sequence('groups', 'id'), GREATEST(1, (SELECT MAX(id) FROM groups)));


CREATE TABLE IF NOT EXISTS feeds (
	id         BIGSERIAL PRIMARY KEY,
	group_id   BIGINT NOT NULL DEFAULT 1 REFERENCES groups(id) ON UPDATE CASCADE ON DELETE RESTRICT,
	name       TEXT NOT NULL,
	link       TEXT NOT NULL UNIQUE,
	site_url   TEXT DEFAULT '',
	suspended  BOOLEAN DEFAULT FALSE,
	proxy      TEXT DEFAULT '',
	created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::bigint,
	updated_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::bigint
);

CREATE INDEX IF NOT EXISTS idx_feeds_group_id ON feeds(group_id);


CREATE TABLE IF NOT EXISTS items (
	id         BIGSERIAL PRIMARY KEY,
	feed_id    BIGINT NOT NULL REFERENCES feeds(id) ON UPDATE CASCADE ON DELETE CASCADE,
	guid       TEXT NOT NULL,
	title      TEXT DEFAULT '',
	link       TEXT DEFAULT '',
	content    TEXT DEFAULT '',
	pub_date   BIGINT DEFAULT 0,
	unread     BOOLEAN DEFAULT TRUE,
	created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::bigint
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_items_feed_guid ON items(feed_id, guid);
CREATE INDEX IF NOT EXISTS idx_items_unread ON items(unread) WHERE unread = true;
CREATE INDEX IF NOT EXISTS idx_items_pub_date ON items(pub_date DESC);
CREATE INDEX IF NOT EXISTS idx_items_feed_unread ON items(feed_id, unread);

ALTER TABLE items ADD COLUMN IF NOT EXISTS fts_doc tsvector GENERATED ALWAYS AS (
	to_tsvector('english', coalesce(title, '') || ' ' || coalesce(content, ''))
) STORED;

CREATE INDEX IF NOT EXISTS idx_items_fts ON items USING GIN(fts_doc);


CREATE TABLE IF NOT EXISTS bookmarks (
	id         BIGSERIAL PRIMARY KEY,
	item_id    BIGINT REFERENCES items(id) ON UPDATE CASCADE ON DELETE SET NULL,
	link       TEXT NOT NULL UNIQUE,
	title      TEXT DEFAULT '',
	content    TEXT DEFAULT '',
	pub_date   BIGINT DEFAULT 0,
	feed_name  TEXT DEFAULT '',
	created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::bigint
);

CREATE INDEX IF NOT EXISTS idx_bookmarks_created_at ON bookmarks(created_at DESC);
