package store

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/0x2E/fusion/internal/model"
)

func (s *Store) ListFeeds() ([]*model.Feed, error) {
	rows, err := s.db.Query(`
		SELECT id, group_id, name, link, site_url, last_build, failure, failures, suspended, proxy, created_at, updated_at
		FROM feeds
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	feeds := []*model.Feed{}
	for rows.Next() {
		f := &model.Feed{}
		var suspended int
		if err := rows.Scan(&f.ID, &f.GroupID, &f.Name, &f.Link, &f.SiteURL, &f.LastBuild, &f.Failure, &f.Failures, &suspended, &f.Proxy, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		f.Suspended = intToBool(suspended)
		feeds = append(feeds, f)
	}
	return feeds, rows.Err()
}

func (s *Store) GetFeed(id int64) (*model.Feed, error) {
	f := &model.Feed{}
	var suspended int
	err := s.db.QueryRow(`
		SELECT id, group_id, name, link, site_url, last_build, failure, failures, suspended, proxy, created_at, updated_at
		FROM feeds
		WHERE id = :id
	`, sql.Named("id", id)).Scan(&f.ID, &f.GroupID, &f.Name, &f.Link, &f.SiteURL, &f.LastBuild, &f.Failure, &f.Failures, &suspended, &f.Proxy, &f.CreatedAt, &f.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("feed not found")
	}
	f.Suspended = intToBool(suspended)
	return f, err
}

func (s *Store) CreateFeed(groupID int64, name, link, siteURL, proxy string) (*model.Feed, error) {
	result, err := s.db.Exec(`
		INSERT INTO feeds (group_id, name, link, site_url, proxy)
		VALUES (:group_id, :name, :link, :site_url, :proxy)
	`, sql.Named("group_id", groupID), sql.Named("name", name), sql.Named("link", link),
		sql.Named("site_url", siteURL), sql.Named("proxy", proxy))
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return s.GetFeed(id)
}

// UpdateFeedParams supports partial updates. Only non-nil fields will be updated.
// Pointer fields distinguish between "not set" (nil) and "set to zero value" (e.g., &false).
type UpdateFeedParams struct {
	GroupID   *int64
	Name      *string
	SiteURL   *string
	Suspended *bool
	Proxy     *string
}

// UpdateFeed performs partial update of feed fields using a single dynamic UPDATE query.
func (s *Store) UpdateFeed(id int64, params UpdateFeedParams) error {
	setClauses := []string{}
	args := []interface{}{sql.Named("id", id)}

	if params.GroupID != nil {
		setClauses = append(setClauses, "group_id = :group_id")
		args = append(args, sql.Named("group_id", *params.GroupID))
	}
	if params.Name != nil {
		setClauses = append(setClauses, "name = :name")
		args = append(args, sql.Named("name", *params.Name))
	}
	if params.SiteURL != nil {
		setClauses = append(setClauses, "site_url = :site_url")
		args = append(args, sql.Named("site_url", *params.SiteURL))
	}
	if params.Suspended != nil {
		setClauses = append(setClauses, "suspended = :suspended")
		args = append(args, sql.Named("suspended", boolToInt(*params.Suspended)))
	}
	if params.Proxy != nil {
		setClauses = append(setClauses, "proxy = :proxy")
		args = append(args, sql.Named("proxy", *params.Proxy))
	}

	if len(setClauses) == 0 {
		return nil
	}

	setClauses = append(setClauses, "updated_at = unixepoch()")
	query := fmt.Sprintf("UPDATE feeds SET %s WHERE id = :id", strings.Join(setClauses, ", "))
	_, err := s.db.Exec(query, args...)
	return err
}

// DeleteFeed removes a feed and all its items in a transaction.
// Bookmarks are preserved by setting their item_id to NULL, maintaining
// the snapshot of content even after original items are deleted.
func (s *Store) DeleteFeed(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		UPDATE bookmarks SET item_id = NULL
		WHERE item_id IN (SELECT id FROM items WHERE feed_id = :feed_id)
	`, sql.Named("feed_id", id)); err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM items WHERE feed_id = :feed_id`, sql.Named("feed_id", id)); err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM feeds WHERE id = :id`, sql.Named("id", id)); err != nil {
		return err
	}

	return tx.Commit()
}

// UpdateFeedLastBuild records successful feed fetch and resets failure counters.
// This allows feeds to auto-recover from temporary network issues.
func (s *Store) UpdateFeedLastBuild(id int64, lastBuild int64) error {
	_, err := s.db.Exec(`
		UPDATE feeds
		SET last_build = :last_build, failures = 0, failure = '', updated_at = unixepoch()
		WHERE id = :id
	`, sql.Named("last_build", lastBuild), sql.Named("id", id))
	return err
}

func (s *Store) UpdateFeedFailure(id int64, failure string) error {
	_, err := s.db.Exec(`
		UPDATE feeds
		SET failures = failures + 1, failure = :failure, updated_at = unixepoch()
		WHERE id = :id
	`, sql.Named("failure", failure), sql.Named("id", id))
	return err
}
