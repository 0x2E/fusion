package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/0x2E/fusion/internal/model"
)

func (s *Store) ListBookmarks(limit, offset int) ([]*model.Bookmark, error) {
	query := `
		SELECT id, item_id, link, title, content, pub_date, feed_name, created_at
		FROM bookmarks
		ORDER BY created_at DESC
	`
	args := []interface{}{}

	if limit > 0 {
		query += ` LIMIT :limit`
		args = append(args, sql.Named("limit", limit))
	}
	if offset > 0 {
		query += ` OFFSET :offset`
		args = append(args, sql.Named("offset", offset))
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

func (s *Store) GetBookmark(id int64) (*model.Bookmark, error) {
	b := &model.Bookmark{}
	err := s.db.QueryRow(`
		SELECT id, item_id, link, title, content, pub_date, feed_name, created_at
		FROM bookmarks
		WHERE id = :id
	`, sql.Named("id", id)).Scan(&b.ID, &b.ItemID, &b.Link, &b.Title, &b.Content, &b.PubDate, &b.FeedName, &b.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: bookmark", ErrNotFound)
		}
		return nil, fmt.Errorf("get bookmark: %w", err)
	}
	return b, nil
}

// CreateBookmark saves a snapshot of content. itemID may be nil if the
// original item was deleted, in which case the bookmark preserves the content.
func (s *Store) CreateBookmark(itemID *int64, link, title, content string, pubDate int64, feedName string) (*model.Bookmark, error) {
	result, err := s.db.Exec(`
		INSERT INTO bookmarks (item_id, link, title, content, pub_date, feed_name)
		VALUES (:item_id, :link, :title, :content, :pub_date, :feed_name)
	`, sql.Named("item_id", itemID), sql.Named("link", link), sql.Named("title", title),
		sql.Named("content", content), sql.Named("pub_date", pubDate), sql.Named("feed_name", feedName))
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return s.GetBookmark(id)
}

func (s *Store) DeleteBookmark(id int64) error {
	result, err := s.db.Exec(`DELETE FROM bookmarks WHERE id = :id`, sql.Named("id", id))
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

func (s *Store) BookmarkExists(link string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM bookmarks WHERE link = :link)`, sql.Named("link", link)).Scan(&exists)
	return exists, err
}
