package store

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/0x2E/fusion/internal/model"
)

// GReaderUnreadSummary is the unread aggregate for one Google Reader scope
// (a feed, a label/group, or the whole reading list).
type GReaderUnreadSummary struct {
	Count int64
	// NewestCreatedAt is the crawl time of the newest unread item, in unix
	// seconds; 0 when the scope has no unread items.
	NewestCreatedAt int64
}

type GReaderUnreadCounts struct {
	ByFeed  map[int64]GReaderUnreadSummary
	ByGroup map[int64]GReaderUnreadSummary
	Total   GReaderUnreadSummary
}

// GReaderUnreadCounts aggregates unread items per feed, per group and overall.
// Google Reader clients expect an entry for every subscription and label,
// including the ones with zero unread items, so callers default missing keys
// to a zero summary.
func (s *Store) GReaderUnreadCounts() (*GReaderUnreadCounts, error) {
	counts := &GReaderUnreadCounts{
		ByFeed:  map[int64]GReaderUnreadSummary{},
		ByGroup: map[int64]GReaderUnreadSummary{},
	}

	// One read transaction gives the three aggregates a single WAL snapshot.
	// Without it, a pull inserting items between the queries makes a feed
	// disagree with its own label and the total in the same response.
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin unread counts: %w", err)
	}
	defer tx.Rollback()

	feedRows, err := tx.Query(`
		SELECT feed_id, COUNT(*), COALESCE(MAX(created_at), 0)
		FROM items
		WHERE unread = 1
		GROUP BY feed_id
	`)
	if err != nil {
		return nil, fmt.Errorf("count unread items per feed: %w", err)
	}
	defer feedRows.Close()

	for feedRows.Next() {
		var feedID int64
		var summary GReaderUnreadSummary
		if err := feedRows.Scan(&feedID, &summary.Count, &summary.NewestCreatedAt); err != nil {
			return nil, fmt.Errorf("scan unread feed count: %w", err)
		}
		counts.ByFeed[feedID] = summary
	}
	if err := feedRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate unread feed counts: %w", err)
	}

	groupRows, err := tx.Query(`
		SELECT feeds.group_id, COUNT(*), COALESCE(MAX(items.created_at), 0)
		FROM items
		INNER JOIN feeds ON feeds.id = items.feed_id
		WHERE items.unread = 1
		GROUP BY feeds.group_id
	`)
	if err != nil {
		return nil, fmt.Errorf("count unread items per group: %w", err)
	}
	defer groupRows.Close()

	for groupRows.Next() {
		var groupID int64
		var summary GReaderUnreadSummary
		if err := groupRows.Scan(&groupID, &summary.Count, &summary.NewestCreatedAt); err != nil {
			return nil, fmt.Errorf("scan unread group count: %w", err)
		}
		counts.ByGroup[groupID] = summary
	}
	if err := groupRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate unread group counts: %w", err)
	}

	if err := tx.QueryRow(`
		SELECT COUNT(*), COALESCE(MAX(created_at), 0)
		FROM items
		WHERE unread = 1
	`).Scan(&counts.Total.Count, &counts.Total.NewestCreatedAt); err != nil {
		return nil, fmt.Errorf("count all unread items: %w", err)
	}

	return counts, nil
}

// GReaderItemsParams filters Google Reader item queries. Stream order is
// items.id, which is monotonic with crawl time, so a continuation cursor is
// just the last returned id.
type GReaderItemsParams struct {
	FeedID  *int64
	GroupID *int64
	// ReadOnly and UnreadOnly are independent filters; setting both matches
	// nothing, so a client combining it=read with xt=read gets an empty page
	// rather than a contradictory one.
	ReadOnly   bool
	UnreadOnly bool
	// Starred/NotStarred likewise restrict to items (not) bookmarked.
	Starred    bool
	NotStarred bool
	// CreatedBefore/After map to the ot/nt filters. They are exclusive
	// bounds on items.created_at (crawl time), not pub_date.
	CreatedBefore *int64
	CreatedAfter  *int64
	// BeforeID/AfterID are the exclusive cursors for descending/ascending
	// pages; the handler picks the direction from r=o.
	BeforeID  *int64
	AfterID   *int64
	Limit     int
	Ascending bool
}

// GReaderItem is an item plus the feed, group and starred metadata that the
// Google Reader item payload exposes.
type GReaderItem struct {
	model.Item
	FeedName    string
	FeedSiteURL string
	GroupName   string
	Starred     bool
}

// GReaderItemRef is the lightweight projection behind stream/items/ids.
type GReaderItemRef struct {
	ID        int64
	CreatedAt int64
}

const greaderItemsColumns = `items.id, items.feed_id, items.guid, items.title, items.link,
	items.content, items.pub_date, items.unread, items.created_at,
	feeds.name, COALESCE(feeds.site_url, ''), groups.name,
	EXISTS(SELECT 1 FROM bookmarks b WHERE b.item_id = items.id)`

const greaderItemsFrom = `
	FROM items
	INNER JOIN feeds ON feeds.id = items.feed_id
	INNER JOIN groups ON groups.id = feeds.group_id
`

func buildGReaderItemsFilter(params GReaderItemsParams) (string, []any) {
	clauses := []string{"1 = 1"}
	args := []any{}

	if params.FeedID != nil {
		clauses = append(clauses, "items.feed_id = :feed_id")
		args = append(args, sql.Named("feed_id", *params.FeedID))
	}
	if params.GroupID != nil {
		clauses = append(clauses, "feeds.group_id = :group_id")
		args = append(args, sql.Named("group_id", *params.GroupID))
	}
	switch {
	case params.ReadOnly && params.UnreadOnly:
		clauses = append(clauses, "1 = 0")
	case params.ReadOnly:
		clauses = append(clauses, "items.unread = 0")
	case params.UnreadOnly:
		clauses = append(clauses, "items.unread = 1")
	}
	if params.Starred {
		clauses = append(clauses, "EXISTS(SELECT 1 FROM bookmarks b WHERE b.item_id = items.id)")
	}
	if params.NotStarred {
		clauses = append(clauses, "NOT EXISTS(SELECT 1 FROM bookmarks b WHERE b.item_id = items.id)")
	}
	if params.CreatedBefore != nil {
		clauses = append(clauses, "items.created_at < :created_before")
		args = append(args, sql.Named("created_before", *params.CreatedBefore))
	}
	if params.CreatedAfter != nil {
		clauses = append(clauses, "items.created_at > :created_after")
		args = append(args, sql.Named("created_after", *params.CreatedAfter))
	}
	if params.BeforeID != nil {
		clauses = append(clauses, "items.id < :before_id")
		args = append(args, sql.Named("before_id", *params.BeforeID))
	}
	if params.AfterID != nil {
		clauses = append(clauses, "items.id > :after_id")
		args = append(args, sql.Named("after_id", *params.AfterID))
	}

	return strings.Join(clauses, " AND "), args
}

func greaderItemsOrder(params GReaderItemsParams) string {
	if params.Ascending {
		return "items.id ASC"
	}
	return "items.id DESC"
}

func scanGReaderItem(row rowScanner) (*GReaderItem, error) {
	item := &GReaderItem{}
	var unread int
	var starred int
	if err := row.Scan(
		&item.ID, &item.FeedID, &item.GUID, &item.Title, &item.Link, &item.Content,
		&item.PubDate, &unread, &item.CreatedAt,
		&item.FeedName, &item.FeedSiteURL, &item.GroupName, &starred,
	); err != nil {
		return nil, err
	}
	item.Unread = intToBool(unread)
	item.Starred = intToBool(starred)
	return item, nil
}

func (s *Store) ListGReaderItems(params GReaderItemsParams) ([]*GReaderItem, error) {
	where, args := buildGReaderItemsFilter(params)
	query := `SELECT ` + greaderItemsColumns + greaderItemsFrom + ` WHERE ` + where +
		` ORDER BY ` + greaderItemsOrder(params)
	if params.Limit > 0 {
		query += ` LIMIT :limit`
		args = append(args, sql.Named("limit", params.Limit))
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list greader items: %w", err)
	}
	defer rows.Close()

	items := []*GReaderItem{}
	for rows.Next() {
		item, err := scanGReaderItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan greader item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate greader items: %w", err)
	}
	return items, nil
}

func (s *Store) ListGReaderItemRefs(params GReaderItemsParams) ([]GReaderItemRef, error) {
	refs := []GReaderItemRef{}

	where, args := buildGReaderItemsFilter(params)
	query := `SELECT items.id, items.created_at` + greaderItemsFrom + ` WHERE ` + where +
		` ORDER BY ` + greaderItemsOrder(params)
	if params.Limit > 0 {
		query += ` LIMIT :limit`
		args = append(args, sql.Named("limit", params.Limit))
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list greader item refs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ref GReaderItemRef
		if err := rows.Scan(&ref.ID, &ref.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan greader item ref: %w", err)
		}
		refs = append(refs, ref)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate greader item refs: %w", err)
	}
	return refs, nil
}

// ListGReaderItemsByIDs fetches items in no particular order; callers restore
// the order the client asked for.
func (s *Store) ListGReaderItemsByIDs(ids []int64) ([]*GReaderItem, error) {
	if len(ids) == 0 {
		return []*GReaderItem{}, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]any, 0, len(ids))
	for i, id := range ids {
		name := fmt.Sprintf("item_id_%d", i)
		placeholders[i] = ":" + name
		args = append(args, sql.Named(name, id))
	}

	query := `SELECT ` + greaderItemsColumns + greaderItemsFrom +
		` WHERE items.id IN (` + strings.Join(placeholders, ", ") + `)`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list greader items by ids: %w", err)
	}
	defer rows.Close()

	items := []*GReaderItem{}
	for rows.Next() {
		item, err := scanGReaderItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan greader item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate greader items by ids: %w", err)
	}
	return items, nil
}

// MarkGReaderItemsReadParams marks items read by crawl time. FeedID and
// GroupID narrow the scope; nil means every feed.
type MarkGReaderItemsReadParams struct {
	FeedID    *int64
	GroupID   *int64
	CreatedAt int64 // inclusive upper bound on items.created_at
}

func (s *Store) MarkGReaderItemsRead(params MarkGReaderItemsReadParams) error {
	query := `UPDATE items SET unread = 0 WHERE created_at <= :created_at`
	args := []any{sql.Named("created_at", params.CreatedAt)}

	if params.FeedID != nil {
		query += ` AND feed_id = :feed_id`
		args = append(args, sql.Named("feed_id", *params.FeedID))
	}
	if params.GroupID != nil {
		query += ` AND feed_id IN (SELECT id FROM feeds WHERE group_id = :group_id)`
		args = append(args, sql.Named("group_id", *params.GroupID))
	}

	if _, err := s.db.Exec(query, args...); err != nil {
		return fmt.Errorf("mark greader items read: %w", err)
	}
	return nil
}
