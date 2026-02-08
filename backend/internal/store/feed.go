package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/0x2E/fusion/internal/model"
)

func (s *Store) ListFeeds() ([]*model.Feed, error) {
	rows, err := s.db.Query(`
		SELECT f.id, f.group_id, f.name, f.link, f.site_url, f.last_build,
		       f.failure, f.failures, f.suspended, f.proxy, f.created_at, f.updated_at,
		       (SELECT COUNT(*) FROM items WHERE feed_id = f.id AND unread = 1) AS unread_count,
		       (SELECT COUNT(*) FROM items WHERE feed_id = f.id) AS item_count
		FROM feeds f
		ORDER BY f.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	feeds := []*model.Feed{}
	for rows.Next() {
		f := &model.Feed{}
		var suspended int
		if err := rows.Scan(&f.ID, &f.GroupID, &f.Name, &f.Link, &f.SiteURL, &f.LastBuild, &f.Failure, &f.Failures, &suspended, &f.Proxy, &f.CreatedAt, &f.UpdatedAt, &f.UnreadCount, &f.ItemCount); err != nil {
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
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: feed", ErrNotFound)
		}
		return nil, fmt.Errorf("get feed: %w", err)
	}

	f.Suspended = intToBool(suspended)
	return f, nil
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

type SearchFeedResult struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Link    string `json:"link"`
	SiteURL string `json:"site_url"`
}

func (s *Store) SearchFeeds(query string) ([]*SearchFeedResult, error) {
	rows, err := s.db.Query(`
		SELECT id, name, link, site_url
		FROM feeds
		WHERE name LIKE :query
		ORDER BY id
	`, sql.Named("query", "%"+query+"%"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	feeds := []*SearchFeedResult{}
	for rows.Next() {
		f := &SearchFeedResult{}
		if err := rows.Scan(&f.ID, &f.Name, &f.Link, &f.SiteURL); err != nil {
			return nil, err
		}
		feeds = append(feeds, f)
	}
	return feeds, rows.Err()
}

// UpdateFeedParams supports partial updates. Only non-nil fields will be updated.
// Pointer fields distinguish between "not set" (nil) and "set to zero value" (e.g., &false).
type UpdateFeedParams struct {
	GroupID   *int64
	Name      *string
	Link      *string
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
	if params.Link != nil {
		setClauses = append(setClauses, "link = :link")
		args = append(args, sql.Named("link", *params.Link))
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
	result, err := s.db.Exec(query, args...)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("%w: feed", ErrNotFound)
	}
	return nil
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

	result, err := tx.Exec(`DELETE FROM feeds WHERE id = :id`, sql.Named("id", id))
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("%w: feed", ErrNotFound)
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

func (s *Store) UpdateFeedSiteURLIfEmpty(id int64, siteURL string) error {
	siteURL = strings.TrimSpace(siteURL)
	if siteURL == "" {
		return nil
	}

	_, err := s.db.Exec(`
		UPDATE feeds
		SET site_url = :site_url, updated_at = unixepoch()
		WHERE id = :id AND (site_url IS NULL OR TRIM(site_url) = '')
	`, sql.Named("site_url", siteURL), sql.Named("id", id))
	return err
}

// BatchCreateFeedsInput holds input for batch feed creation.
type BatchCreateFeedsInput struct {
	GroupID int64
	Name    string
	Link    string
	SiteURL string
}

// BatchCreateFeedsResult holds the result of batch feed creation.
type BatchCreateFeedsResult struct {
	Created    int
	CreatedIDs []int64
	Errors     []string
}

// BatchCreateFeeds creates multiple feeds in a single transaction.
// It skips feeds with duplicate links and returns statistics.
func (s *Store) BatchCreateFeeds(inputs []BatchCreateFeedsInput) (*BatchCreateFeedsResult, error) {
	result := &BatchCreateFeedsResult{}

	if len(inputs) == 0 {
		return result, nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Check existing links to avoid duplicates
	existingLinks := make(map[string]bool)
	rows, err := tx.Query(`SELECT link FROM feeds`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var link string
		if err := rows.Scan(&link); err != nil {
			rows.Close()
			return nil, err
		}
		existingLinks[link] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO feeds (group_id, name, link, site_url, proxy)
		VALUES (:group_id, :name, :link, :site_url, '')
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	for _, input := range inputs {
		if existingLinks[input.Link] {
			result.Errors = append(result.Errors, fmt.Sprintf("duplicate feed: %s", input.Link))
			continue
		}

		res, err := stmt.Exec(
			sql.Named("group_id", input.GroupID),
			sql.Named("name", input.Name),
			sql.Named("link", input.Link),
			sql.Named("site_url", input.SiteURL),
		)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to create %s: %v", input.Link, err))
			continue
		}

		id, err := res.LastInsertId()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to get id for %s: %v", input.Link, err))
			continue
		}

		existingLinks[input.Link] = true
		result.Created++
		result.CreatedIDs = append(result.CreatedIDs, id)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return result, nil
}
