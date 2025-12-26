package store

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/0x2E/fusion/internal/model"
)

type ListItemsParams struct {
	FeedID  *int64
	Unread  *bool
	Limit   int
	Offset  int
	OrderBy string // "pub_date" or "created_at"
}

func (s *Store) ListItems(params ListItemsParams) ([]*model.Item, error) {
	query := `
		SELECT id, feed_id, guid, title, link, content, pub_date, unread, created_at
		FROM items
		WHERE 1=1
	`
	args := []interface{}{}

	if params.FeedID != nil {
		query += ` AND feed_id = :feed_id`
		args = append(args, sql.Named("feed_id", *params.FeedID))
	}
	if params.Unread != nil {
		unread := 0
		if *params.Unread {
			unread = 1
		}
		query += ` AND unread = :unread`
		args = append(args, sql.Named("unread", unread))
	}

	orderBy := "pub_date DESC"
	if params.OrderBy == "created_at" {
		orderBy = "created_at DESC"
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

	var items []*model.Item
	for rows.Next() {
		i := &model.Item{}
		var unread int
		if err := rows.Scan(&i.ID, &i.FeedID, &i.GUID, &i.Title, &i.Link, &i.Content, &i.PubDate, &unread, &i.CreatedAt); err != nil {
			return nil, err
		}
		i.Unread = unread != 0
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
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("item not found")
	}
	i.Unread = unread != 0
	return i, err
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
	u := 0
	if unread {
		u = 1
	}
	_, err := s.db.Exec(`UPDATE items SET unread = :unread WHERE id = :id`,
		sql.Named("unread", u), sql.Named("id", id))
	return err
}

func (s *Store) BatchUpdateItemsUnread(ids []int64, unread bool) error {
	if len(ids) == 0 {
		return nil
	}

	u := 0
	if unread {
		u = 1
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, 0, len(ids)+1)
	args = append(args, sql.Named("unread", u))
	for i, id := range ids {
		paramName := fmt.Sprintf("id%d", i)
		placeholders[i] = ":" + paramName
		args = append(args, sql.Named(paramName, id))
	}

	query := fmt.Sprintf(`UPDATE items SET unread = :unread WHERE id IN (%s)`, strings.Join(placeholders, ","))
	_, err := s.db.Exec(query, args...)
	return err
}

func (s *Store) MarkAllAsRead(feedID *int64) error {
	if feedID != nil {
		_, err := s.db.Exec(`UPDATE items SET unread = 0 WHERE feed_id = :feed_id`, sql.Named("feed_id", *feedID))
		return err
	}
	_, err := s.db.Exec(`UPDATE items SET unread = 0`)
	return err
}

func (s *Store) DeleteItem(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE bookmarks SET item_id = NULL WHERE item_id = :id`, sql.Named("id", id)); err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM items WHERE id = :id`, sql.Named("id", id)); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) ItemExists(feedID int64, guid string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM items WHERE feed_id = :feed_id AND guid = :guid)`,
		sql.Named("feed_id", feedID), sql.Named("guid", guid)).Scan(&exists)
	return exists, err
}
