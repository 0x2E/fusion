-- User preferences persisted in the backend (#167). Fusion is single-user, so
-- this is a single-row table (id = 1). NULL means "the user has never chosen a
-- value"; the API never writes NULL, and a PATCH field that is absent leaves
-- the stored value unchanged. The allow-lists below are duplicated in the
-- handler validation and the frontend constants; note that SQLite cannot ALTER
-- a CHECK constraint in place, so extending a list requires a table rebuild.

CREATE TABLE IF NOT EXISTS settings (
	id                INTEGER PRIMARY KEY CHECK (id = 1),
	locale            TEXT CHECK (locale IN ('en', 'zh', 'de', 'fr', 'es', 'ru', 'pt', 'sv') OR locale IS NULL),
	article_page_size INTEGER CHECK (article_page_size IN (10, 20, 30, 50, 100) OR article_page_size IS NULL),
	theme             TEXT CHECK (theme IN ('light', 'dark', 'system') OR theme IS NULL),
	updated_at        INTEGER NOT NULL DEFAULT (unixepoch())
);
