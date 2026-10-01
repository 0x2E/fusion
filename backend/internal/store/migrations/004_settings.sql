-- User preferences persisted in the backend (#167). Fusion is single-user, so
-- this is a single-row table (id = 1). NULL means "the user has never chosen a
-- value"; the API never writes NULL, and a PATCH field that is absent leaves
-- the stored value unchanged.
--
-- Value validation lives in the handler (the only write path), not in CHECK
-- constraints: allow-lists change with product decisions, and SQLite cannot
-- ALTER a CHECK in place, so putting them here would make every adjustment a
-- table rebuild.

CREATE TABLE IF NOT EXISTS settings (
	id                INTEGER PRIMARY KEY CHECK (id = 1),
	locale            TEXT,
	article_page_size INTEGER,
	theme             TEXT,
	updated_at        INTEGER NOT NULL DEFAULT (unixepoch())
);
