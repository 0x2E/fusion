ALTER TABLE feeds ADD COLUMN crawl INTEGER NOT NULL DEFAULT 0;
ALTER TABLE items ADD COLUMN crawled_at INTEGER DEFAULT NULL;
CREATE INDEX idx_items_feed_uncrawled ON items(feed_id, created_at DESC) WHERE crawled_at IS NULL;
