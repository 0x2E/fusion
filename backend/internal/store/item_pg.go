package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/0x2E/fusion/internal/model"
)

func (s *PGStore) ListItems(params ListItemsParams) ([]*model.Item, error) {
	query := `
		SELECT items.id, items.feed_id, items.guid, items.title, items.link, items.content, items.pub_date, items.unread, items.created_at
		FROM items
	`
	args := []interface{}{}
	paramIdx := 1

	if params.GroupID != nil {
		query += ` INNER JOIN feeds ON items.feed_id = feeds.id`
	}

	query += ` WHERE 1=1`

	if params.FeedID != nil {
		query += fmt.Sprintf(` AND items.feed_id = $%d`, paramIdx)
		args = append(args, *params.FeedID)
		paramIdx++
	}
	if params.GroupID != nil {
		query += fmt.Sprintf(` AND feeds.group_id = $%d`, paramIdx)
		args = append(args, *params.GroupID)
		paramIdx++
	}
	if params.Unread != nil {
		query += fmt.Sprintf(` AND items.unread = $%d`, paramIdx)
		args = append(args, *params.Unread)
		paramIdx++
	}

	orderBy := "items.pub_date DESC, items.id DESC"
	if params.OrderBy == "created_at" {
		orderBy = "items.created_at DESC, items.id DESC"
	}
	query += ` ORDER BY ` + orderBy

	if params.Limit > 0 {
		query += fmt.Sprintf(` LIMIT $%d`, paramIdx)
		args = append(args, params.Limit)
		paramIdx++
	}
	if params.Offset > 0 {
		query += fmt.Sprintf(` OFFSET $%d`, paramIdx)
		args = append(args, params.Offset)
		paramIdx++
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []*model.Item{}
	for rows.Next() {
		i := &model.Item{}
		if err := rows.Scan(&i.ID, &i.FeedID, &i.GUID, &i.Title, &i.Link, &i.Content, &i.PubDate, &i.Unread, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (s *PGStore) GetItem(id int64) (*model.Item, error) {
	i := &model.Item{}
	err := s.db.QueryRow(`
		SELECT id, feed_id, guid, title, link, content, pub_date, unread, created_at
		FROM items
		WHERE id = $1
	`, id).Scan(&i.ID, &i.FeedID, &i.GUID, &i.Title, &i.Link, &i.Content, &i.PubDate, &i.Unread, &i.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: item", ErrNotFound)
		}
		return nil, fmt.Errorf("get item: %w", err)
	}
	return i, nil
}

func (s *PGStore) CreateItem(feedID int64, guid, title, link, content string, pubDate int64) (*model.Item, error) {
	var id int64
	if err := s.db.QueryRow(`
		INSERT INTO items (feed_id, guid, title, link, content, pub_date)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, feedID, guid, title, link, content, pubDate).Scan(&id); err != nil {
		return nil, err
	}

	return s.GetItem(id)
}

func (s *PGStore) BatchCreateItemsIgnore(feedID int64, inputs []BatchCreateItemInput) (int, error) {
	if len(inputs) == 0 {
		return 0, nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO items (feed_id, guid, title, link, content, pub_date)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT(feed_id, guid) DO NOTHING
	`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	created := 0
	for _, input := range inputs {
		result, err := stmt.Exec(feedID, input.GUID, input.Title, input.Link, input.Content, input.PubDate)
		if err != nil {
			return 0, err
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return 0, err
		}
		if affected > 0 {
			created++
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return created, nil
}

func (s *PGStore) UpdateItemUnread(id int64, unread bool) error {
	result, err := s.db.Exec(`UPDATE items SET unread = $1 WHERE id = $2`, unread, id)
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

func (s *PGStore) BatchUpdateItemsUnread(ids []int64, unread bool) error {
	if len(ids) == 0 {
		return nil
	}

	const chunkSize = 500
	for start := 0; start < len(ids); start += chunkSize {
		end := start + chunkSize
		if end > len(ids) {
			end = len(ids)
		}

		if err := s.pgBatchUpdateItemsUnreadChunk(ids[start:end], unread); err != nil {
			return err
		}
	}

	return nil
}

func (s *PGStore) pgBatchUpdateItemsUnreadChunk(ids []int64, unread bool) error {
	if len(ids) == 0 {
		return nil
	}

	// $1 = unread, ids start at $2
	args := make([]interface{}, 0, len(ids)+1)
	args = append(args, unread)
	placeholders := make([]string, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, id)
	}

	query := fmt.Sprintf(`UPDATE items SET unread = $1 WHERE id IN (%s)`, strings.Join(placeholders, ","))
	_, err := s.db.Exec(query, args...)
	return err
}

func (s *PGStore) MarkAllAsRead(feedID *int64) error {
	if feedID != nil {
		_, err := s.db.Exec(`UPDATE items SET unread = false WHERE feed_id = $1`, *feedID)
		return err
	}
	_, err := s.db.Exec(`UPDATE items SET unread = false`)
	return err
}

func (s *PGStore) ItemExists(feedID int64, guid string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM items WHERE feed_id = $1 AND guid = $2)`,
		feedID, guid).Scan(&exists)
	return exists, err
}

func (s *PGStore) SearchItems(query string, limit int) ([]*SearchItemResult, error) {
	if strings.TrimSpace(query) == "" {
		return s.pgSearchItemsLike(query, limit)
	}

	rows, err := s.db.Query(`
		SELECT i.id, i.feed_id, i.title, i.pub_date
		FROM items i
		WHERE i.fts_doc @@ plainto_tsquery('english', $1)
		ORDER BY i.pub_date DESC, i.id DESC
		LIMIT $2
	`, query, limit)
	if err != nil {
		return s.pgSearchItemsLike(query, limit)
	}
	defer rows.Close()

	items := []*SearchItemResult{}
	for rows.Next() {
		i := &SearchItemResult{}
		if err := rows.Scan(&i.ID, &i.FeedID, &i.Title, &i.PubDate); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return s.pgSearchItemsLike(query, limit)
	}
	return items, nil
}

func (s *PGStore) pgSearchItemsLike(query string, limit int) ([]*SearchItemResult, error) {
	rows, err := s.db.Query(`
		SELECT id, feed_id, title, pub_date
		FROM items
		WHERE title LIKE $1 OR content LIKE $1
		ORDER BY pub_date DESC, id DESC
		LIMIT $2
	`, "%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []*SearchItemResult{}
	for rows.Next() {
		i := &SearchItemResult{}
		if err := rows.Scan(&i.ID, &i.FeedID, &i.Title, &i.PubDate); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (s *PGStore) CountItems(params ListItemsParams) (int, error) {
	query := `SELECT COUNT(*) FROM items`
	args := []interface{}{}
	paramIdx := 1

	if params.GroupID != nil {
		query += ` INNER JOIN feeds ON items.feed_id = feeds.id`
	}

	query += ` WHERE 1=1`

	if params.FeedID != nil {
		query += fmt.Sprintf(` AND items.feed_id = $%d`, paramIdx)
		args = append(args, *params.FeedID)
		paramIdx++
	}
	if params.GroupID != nil {
		query += fmt.Sprintf(` AND feeds.group_id = $%d`, paramIdx)
		args = append(args, *params.GroupID)
		paramIdx++
	}
	if params.Unread != nil {
		query += fmt.Sprintf(` AND items.unread = $%d`, paramIdx)
		args = append(args, *params.Unread)
		paramIdx++
	}

	var count int
	err := s.db.QueryRow(query, args...).Scan(&count)
	return count, err
}
