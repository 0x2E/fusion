package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/0x2E/fusion/internal/model"
)

// ListItemsParams specifies filtering and pagination for item queries.
//
// Pointer fields (FeedID, GroupID, Unread) are optional filters - nil means "no filter".
// OrderBy accepts "pub_date" (default) or "created_at".
// Limit/Offset = 0 means no limit/offset.
type ListItemsParams struct {
	FeedID  *int64
	GroupID *int64
	Unread  *bool
	Limit   int
	Offset  int
	OrderBy string // "pub_date" or "created_at"
}

func (s *Store) ListItems(params ListItemsParams) ([]*model.Item, error) {
	query := `
		SELECT items.id, items.feed_id, items.guid, items.title, items.link, items.content, items.pub_date, items.unread, items.created_at
		FROM items
	`
	args := []interface{}{}

	// Join feeds table if filtering by GroupID
	if params.GroupID != nil {
		query += ` INNER JOIN feeds ON items.feed_id = feeds.id`
	}

	query += ` WHERE 1=1`

	if params.FeedID != nil {
		query += ` AND items.feed_id = :feed_id`
		args = append(args, sql.Named("feed_id", *params.FeedID))
	}
	if params.GroupID != nil {
		query += ` AND feeds.group_id = :group_id`
		args = append(args, sql.Named("group_id", *params.GroupID))
	}
	if params.Unread != nil {
		query += ` AND items.unread = :unread`
		args = append(args, sql.Named("unread", boolToInt(*params.Unread)))
	}

	// ORDER BY cannot use named parameters, validated via allowlist instead
	orderBy := "items.pub_date DESC, items.id DESC"
	if params.OrderBy == "created_at" {
		orderBy = "items.created_at DESC, items.id DESC"
	}
	query += ` ORDER BY ` + orderBy

	if params.Limit > 0 {
		query += ` LIMIT :limit`
		args = append(args, sql.Named("limit", params.Limit))
	}
	if params.Offset > 0 {
		query += ` OFFSET :offset`
		args = append(args, sql.Named("offset", params.Offset))
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []*model.Item{}
	for rows.Next() {
		i := &model.Item{}
		var unread int
		if err := rows.Scan(&i.ID, &i.FeedID, &i.GUID, &i.Title, &i.Link, &i.Content, &i.PubDate, &unread, &i.CreatedAt); err != nil {
			return nil, err
		}
		i.Unread = intToBool(unread)
		items = append(items, i)
	}
	return items, rows.Err()
}

func (s *Store) GetItem(id int64) (*model.Item, error) {
	i := &model.Item{}
	var unread int
	err := s.db.QueryRow(`
		SELECT id, feed_id, guid, title, link, content, pub_date, unread, created_at
		FROM items
		WHERE id = :id
	`, sql.Named("id", id)).Scan(&i.ID, &i.FeedID, &i.GUID, &i.Title, &i.Link, &i.Content, &i.PubDate, &unread, &i.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: item", ErrNotFound)
		}
		return nil, fmt.Errorf("get item: %w", err)
	}

	i.Unread = intToBool(unread)
	return i, nil
}

func (s *Store) CreateItem(feedID int64, guid, title, link, content string, pubDate int64) (*model.Item, error) {
	result, err := s.db.Exec(`
		INSERT INTO items (feed_id, guid, title, link, content, pub_date)
		VALUES (:feed_id, :guid, :title, :link, :content, :pub_date)
	`, sql.Named("feed_id", feedID), sql.Named("guid", guid), sql.Named("title", title),
		sql.Named("link", link), sql.Named("content", content), sql.Named("pub_date", pubDate))
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return s.GetItem(id)
}

func (s *Store) UpdateItemUnread(id int64, unread bool) error {
	result, err := s.db.Exec(`UPDATE items SET unread = :unread WHERE id = :id`,
		sql.Named("unread", boolToInt(unread)), sql.Named("id", id))
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("%w: item", ErrNotFound)
	}
	return nil
}

// BatchUpdateItemsUnread marks multiple items as read/unread in a single query.
// Dynamically builds IN clause with named parameters (:id0, :id1, ...) for safety.
func (s *Store) BatchUpdateItemsUnread(ids []int64, unread bool) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, 0, len(ids)+1)
	args = append(args, sql.Named("unread", boolToInt(unread)))
	for i, id := range ids {
		paramName := fmt.Sprintf("id%d", i)
		placeholders[i] = ":" + paramName
		args = append(args, sql.Named(paramName, id))
	}

	query := fmt.Sprintf(`UPDATE items SET unread = :unread WHERE id IN (%s)`, strings.Join(placeholders, ","))
	_, err := s.db.Exec(query, args...)
	return err
}

// MarkAllAsRead marks items as read. If feedID is nil, marks ALL items across all feeds.
// If feedID is non-nil, only marks items from that specific feed.
func (s *Store) MarkAllAsRead(feedID *int64) error {
	if feedID != nil {
		_, err := s.db.Exec(`UPDATE items SET unread = 0 WHERE feed_id = :feed_id`, sql.Named("feed_id", *feedID))
		return err
	}
	_, err := s.db.Exec(`UPDATE items SET unread = 0`)
	return err
}

func (s *Store) ItemExists(feedID int64, guid string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM items WHERE feed_id = :feed_id AND guid = :guid)`,
		sql.Named("feed_id", feedID), sql.Named("guid", guid)).Scan(&exists)
	return exists, err
}
