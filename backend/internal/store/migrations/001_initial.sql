-- Initial schema

CREATE TABLE IF NOT EXISTS groups (
	id         INTEGER PRIMARY KEY,
	name       TEXT NOT NULL UNIQUE,
	created_at INTEGER NOT NULL DEFAULT (unixepoch()),
	updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);
INSERT OR IGNORE INTO groups (id, name) VALUES (1, 'Default');


CREATE TABLE IF NOT EXISTS feeds (
	id         INTEGER PRIMARY KEY,
	group_id   INTEGER NOT NULL DEFAULT 1,
	name       TEXT NOT NULL,
	link       TEXT NOT NULL UNIQUE,
	site_url   TEXT DEFAULT '',
	last_build INTEGER DEFAULT 0,
	last_failure_at INTEGER NOT NULL DEFAULT 0,
	failure    TEXT DEFAULT '',
	failures   INTEGER DEFAULT 0,
	suspended  INTEGER DEFAULT 0,
	proxy      TEXT DEFAULT '',
	created_at INTEGER NOT NULL DEFAULT (unixepoch()),
	updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE INDEX IF NOT EXISTS idx_feeds_group_id ON feeds(group_id);


CREATE TABLE IF NOT EXISTS items (
	id         INTEGER PRIMARY KEY,
	feed_id    INTEGER NOT NULL,
	guid       TEXT NOT NULL,
	title      TEXT DEFAULT '',
	link       TEXT DEFAULT '',
	content    TEXT DEFAULT '',
	pub_date   INTEGER DEFAULT 0,
	unread     INTEGER DEFAULT 1,
	created_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_items_feed_guid ON items(feed_id, guid);
CREATE INDEX IF NOT EXISTS idx_items_unread ON items(unread) WHERE unread = 1;
CREATE INDEX IF NOT EXISTS idx_items_pub_date ON items(pub_date DESC);
CREATE INDEX IF NOT EXISTS idx_items_feed_unread ON items(feed_id, unread);

CREATE VIRTUAL TABLE IF NOT EXISTS items_fts USING fts5(
	title,
	content,
	tokenize = 'unicode61'
);

INSERT INTO items_fts(rowid, title, content)
SELECT id, title, content
FROM items
WHERE id NOT IN (SELECT rowid FROM items_fts);

CREATE TRIGGER IF NOT EXISTS items_fts_items_ai AFTER INSERT ON items BEGIN
	INSERT INTO items_fts(rowid, title, content)
	VALUES (new.id, new.title, new.content);
END;

CREATE TRIGGER IF NOT EXISTS items_fts_items_ad AFTER DELETE ON items BEGIN
	DELETE FROM items_fts WHERE rowid = old.id;
END;

CREATE TRIGGER IF NOT EXISTS items_fts_items_au AFTER UPDATE ON items BEGIN
	DELETE FROM items_fts WHERE rowid = old.id;
	INSERT INTO items_fts(rowid, title, content)
	VALUES (new.id, new.title, new.content);
END;


CREATE TABLE IF NOT EXISTS bookmarks (
	id         INTEGER PRIMARY KEY,
	item_id    INTEGER,
	link       TEXT NOT NULL UNIQUE,
	title      TEXT DEFAULT '',
	content    TEXT DEFAULT '',
	pub_date   INTEGER DEFAULT 0,
	feed_name  TEXT DEFAULT '',
	created_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE INDEX IF NOT EXISTS idx_bookmarks_created_at ON bookmarks(created_at DESC);
