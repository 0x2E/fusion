package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/0x2E/fusion/internal/model"
)

func (s *PGStore) ListBookmarks(limit, offset int) ([]*model.Bookmark, error) {
	query := `
		SELECT id, item_id, link, title, content, pub_date, feed_name, created_at
		FROM bookmarks
		ORDER BY created_at DESC, id DESC
	`
	args := []interface{}{}
	paramIdx := 1

	if limit > 0 {
		query += fmt.Sprintf(` LIMIT $%d`, paramIdx)
		args = append(args, limit)
		paramIdx++
	}
	if offset > 0 {
		query += fmt.Sprintf(` OFFSET $%d`, paramIdx)
		args = append(args, offset)
		paramIdx++
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bookmarks := []*model.Bookmark{}
	for rows.Next() {
		b := &model.Bookmark{}
		if err := rows.Scan(&b.ID, &b.ItemID, &b.Link, &b.Title, &b.Content, &b.PubDate, &b.FeedName, &b.CreatedAt); err != nil {
			return nil, err
		}
		bookmarks = append(bookmarks, b)
	}
	return bookmarks, rows.Err()
}

func (s *PGStore) GetBookmark(id int64) (*model.Bookmark, error) {
	b := &model.Bookmark{}
	err := s.db.QueryRow(`
		SELECT id, item_id, link, title, content, pub_date, feed_name, created_at
		FROM bookmarks
		WHERE id = $1
	`, id).Scan(&b.ID, &b.ItemID, &b.Link, &b.Title, &b.Content, &b.PubDate, &b.FeedName, &b.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: bookmark", ErrNotFound)
		}
		return nil, fmt.Errorf("get bookmark: %w", err)
	}
	return b, nil
}

func (s *PGStore) CreateBookmark(itemID *int64, link, title, content string, pubDate int64, feedName string) (*model.Bookmark, error) {
	var id int64
	if err := s.db.QueryRow(`
		INSERT INTO bookmarks (item_id, link, title, content, pub_date, feed_name)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, itemID, link, title, content, pubDate, feedName).Scan(&id); err != nil {
		return nil, err
	}

	return s.GetBookmark(id)
}

func (s *PGStore) DeleteBookmark(id int64) error {
	result, err := s.db.Exec(`DELETE FROM bookmarks WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("%w: bookmark", ErrNotFound)
	}
	return nil
}

func (s *PGStore) BookmarkExists(link string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM bookmarks WHERE link = $1)`, link).Scan(&exists)
	return exists, err
}

func (s *PGStore) CountBookmarks() (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM bookmarks`).Scan(&count)
	return count, err
}
