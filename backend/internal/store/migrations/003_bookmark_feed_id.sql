-- Add feed_id to bookmarks as a soft association to a feed for per-feed /
-- per-group filtering. Bookmarks are intentionally independent of feeds: when
-- the source feed is deleted the bookmark survives (its content/feed_name
-- snapshot remains), and feed_id is cleared to NULL by the foreign key.
-- feed_name (a snapshot string) is unreliable after a feed is renamed, so
-- feed_id is the authoritative link while the feed exists.

ALTER TABLE bookmarks ADD COLUMN feed_id INTEGER REFERENCES feeds(id) ON UPDATE CASCADE ON DELETE SET NULL;

UPDATE bookmarks
SET feed_id = (
	SELECT i.feed_id
	FROM items i
	WHERE i.id = bookmarks.item_id
)
WHERE item_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_bookmarks_feed_id ON bookmarks(feed_id);
