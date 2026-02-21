package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/patrickjmcd/reedme/internal/model"
	"github.com/patrickjmcd/reedme/internal/pullpolicy"
)

func (s *Store) ListFeeds() ([]*model.Feed, error) {
	rows, err := s.db.Query(`
		SELECT f.id, f.group_id, f.name, f.link, f.site_url,
		       f.suspended, f.proxy, f.created_at, f.updated_at,
		       COALESCE(fs.etag, ''), COALESCE(fs.last_modified, ''), COALESCE(fs.cache_control, ''),
		       COALESCE(fs.expires_at, 0), COALESCE(fs.last_checked_at, 0), COALESCE(fs.next_check_at, 0),
		       COALESCE(fs.last_http_status, 0), COALESCE(fs.retry_after_until, 0), COALESCE(fs.last_success_at, 0),
		       COALESCE(fs.last_error_at, 0), COALESCE(fs.last_error, ''), COALESCE(fs.consecutive_failures, 0),
		       COALESCE(SUM(CASE WHEN i.unread = 1 THEN 1 ELSE 0 END), 0) AS unread_count,
		       COALESCE(COUNT(i.id), 0) AS item_count
		FROM feeds f
		LEFT JOIN feed_fetch_state fs ON fs.feed_id = f.id
		LEFT JOIN items i ON i.feed_id = f.id
		GROUP BY f.id, f.group_id, f.name, f.link, f.site_url,
		         f.suspended, f.proxy, f.created_at, f.updated_at,
		         fs.etag, fs.last_modified, fs.cache_control, fs.expires_at, fs.last_checked_at,
		         fs.next_check_at, fs.last_http_status, fs.retry_after_until, fs.last_success_at,
		         fs.last_error_at, fs.last_error, fs.consecutive_failures
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
		if err := rows.Scan(
			&f.ID,
			&f.GroupID,
			&f.Name,
			&f.Link,
			&f.SiteURL,
			&suspended,
			&f.Proxy,
			&f.CreatedAt,
			&f.UpdatedAt,
			&f.FetchState.ETag,
			&f.FetchState.LastModified,
			&f.FetchState.CacheControl,
			&f.FetchState.ExpiresAt,
			&f.FetchState.LastCheckedAt,
			&f.FetchState.NextCheckAt,
			&f.FetchState.LastHTTPStatus,
			&f.FetchState.RetryAfterUntil,
			&f.FetchState.LastSuccessAt,
			&f.FetchState.LastErrorAt,
			&f.FetchState.LastError,
			&f.FetchState.ConsecutiveFailures,
			&f.UnreadCount,
			&f.ItemCount,
		); err != nil {
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
		SELECT f.id, f.group_id, f.name, f.link, f.site_url,
		       f.suspended, f.proxy, f.created_at, f.updated_at,
		       COALESCE(fs.etag, ''), COALESCE(fs.last_modified, ''), COALESCE(fs.cache_control, ''),
		       COALESCE(fs.expires_at, 0), COALESCE(fs.last_checked_at, 0), COALESCE(fs.next_check_at, 0),
		       COALESCE(fs.last_http_status, 0), COALESCE(fs.retry_after_until, 0), COALESCE(fs.last_success_at, 0),
		       COALESCE(fs.last_error_at, 0), COALESCE(fs.last_error, ''), COALESCE(fs.consecutive_failures, 0)
		FROM feeds f
		LEFT JOIN feed_fetch_state fs ON fs.feed_id = f.id
		WHERE f.id = :id
	`, sql.Named("id", id)).Scan(
		&f.ID,
		&f.GroupID,
		&f.Name,
		&f.Link,
		&f.SiteURL,
		&suspended,
		&f.Proxy,
		&f.CreatedAt,
		&f.UpdatedAt,
		&f.FetchState.ETag,
		&f.FetchState.LastModified,
		&f.FetchState.CacheControl,
		&f.FetchState.ExpiresAt,
		&f.FetchState.LastCheckedAt,
		&f.FetchState.NextCheckAt,
		&f.FetchState.LastHTTPStatus,
		&f.FetchState.RetryAfterUntil,
		&f.FetchState.LastSuccessAt,
		&f.FetchState.LastErrorAt,
		&f.FetchState.LastError,
		&f.FetchState.ConsecutiveFailures,
	)
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
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
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

	if _, err := tx.Exec(`
		INSERT INTO feed_fetch_state (feed_id, next_check_at)
		VALUES (:feed_id, unixepoch())
	`, sql.Named("feed_id", id)); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
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

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	setClauses = append(setClauses, "updated_at = unixepoch()")
	query := fmt.Sprintf("UPDATE feeds SET %s WHERE id = :id", strings.Join(setClauses, ", "))
	result, err := tx.Exec(query, args...)
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

	if params.Link != nil {
		if _, err := tx.Exec(`
			INSERT INTO feed_fetch_state (
				feed_id,
				etag,
				last_modified,
				cache_control,
				expires_at,
				last_checked_at,
				next_check_at,
				last_http_status,
				retry_after_until,
				last_success_at,
				last_error_at,
				last_error,
				consecutive_failures,
				updated_at
			)
			VALUES (
				:feed_id,
				'',
				'',
				'',
				0,
				0,
				unixepoch(),
				0,
				0,
				0,
				0,
				'',
				0,
				unixepoch()
			)
			ON CONFLICT(feed_id) DO UPDATE SET
				etag = '',
				last_modified = '',
				cache_control = '',
				expires_at = 0,
				last_checked_at = 0,
				next_check_at = unixepoch(),
				last_http_status = 0,
				retry_after_until = 0,
				last_success_at = 0,
				last_error_at = 0,
				last_error = '',
				consecutive_failures = 0,
				updated_at = unixepoch()
		`, sql.Named("feed_id", id)); err != nil {
			return err
		}
	}

	return tx.Commit()
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

type UpdateFeedFetchSuccessParams struct {
	CheckedAt       int64
	HTTPStatus      int
	ETag            string
	LastModified    string
	CacheControl    string
	ExpiresAt       int64
	RetryAfterUntil int64
	NextCheckAt     int64
}

func (s *Store) UpdateFeedFetchSuccess(id int64, params UpdateFeedFetchSuccessParams) error {
	_, err := s.db.Exec(`
		INSERT INTO feed_fetch_state (
			feed_id,
			etag,
			last_modified,
			cache_control,
			expires_at,
			last_checked_at,
			next_check_at,
			last_http_status,
			retry_after_until,
			last_success_at,
			last_error_at,
			last_error,
			consecutive_failures,
			updated_at
		)
		VALUES (
			:feed_id,
			:etag,
			:last_modified,
			:cache_control,
			:expires_at,
			:last_checked_at,
			:next_check_at,
			:last_http_status,
			:retry_after_until,
			:last_success_at,
			0,
			'',
			0,
			unixepoch()
		)
		ON CONFLICT(feed_id) DO UPDATE SET
			etag = excluded.etag,
			last_modified = excluded.last_modified,
			cache_control = excluded.cache_control,
			expires_at = excluded.expires_at,
			last_checked_at = excluded.last_checked_at,
			next_check_at = excluded.next_check_at,
			last_http_status = excluded.last_http_status,
			retry_after_until = excluded.retry_after_until,
			last_success_at = excluded.last_success_at,
			last_error_at = 0,
			last_error = '',
			consecutive_failures = 0,
			updated_at = unixepoch()
	`,
		sql.Named("feed_id", id),
		sql.Named("etag", strings.TrimSpace(params.ETag)),
		sql.Named("last_modified", strings.TrimSpace(params.LastModified)),
		sql.Named("cache_control", strings.TrimSpace(params.CacheControl)),
		sql.Named("expires_at", params.ExpiresAt),
		sql.Named("last_checked_at", params.CheckedAt),
		sql.Named("next_check_at", params.NextCheckAt),
		sql.Named("last_http_status", params.HTTPStatus),
		sql.Named("retry_after_until", params.RetryAfterUntil),
		sql.Named("last_success_at", params.CheckedAt),
	)
	return err
}

type UpdateFeedFetchFailureParams struct {
	CheckedAt       int64
	HTTPStatus      int
	LastError       string
	RetryAfterUntil int64
	IntervalSeconds int64
	MaxBackoff      int64
}

func (s *Store) UpdateFeedFetchFailure(id int64, params UpdateFeedFetchFailureParams) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		INSERT INTO feed_fetch_state (
			feed_id,
			last_checked_at,
			next_check_at,
			last_http_status,
			retry_after_until,
			last_error_at,
			last_error,
			consecutive_failures,
			updated_at
		)
		VALUES (
			:feed_id,
			:last_checked_at,
			:next_check_at,
			:last_http_status,
			:retry_after_until,
			:last_error_at,
			:last_error,
			1,
			unixepoch()
		)
		ON CONFLICT(feed_id) DO UPDATE SET
			last_checked_at = excluded.last_checked_at,
			next_check_at = excluded.next_check_at,
			last_http_status = excluded.last_http_status,
			retry_after_until = excluded.retry_after_until,
			last_error_at = excluded.last_error_at,
			last_error = excluded.last_error,
			consecutive_failures = feed_fetch_state.consecutive_failures + 1,
			updated_at = unixepoch()
	`,
		sql.Named("feed_id", id),
		sql.Named("last_checked_at", params.CheckedAt),
		sql.Named("next_check_at", params.CheckedAt),
		sql.Named("last_http_status", params.HTTPStatus),
		sql.Named("retry_after_until", params.RetryAfterUntil),
		sql.Named("last_error_at", params.CheckedAt),
		sql.Named("last_error", strings.TrimSpace(params.LastError)),
	); err != nil {
		return err
	}

	var newFailures int64
	if err := tx.QueryRow(`
		SELECT consecutive_failures
		FROM feed_fetch_state
		WHERE feed_id = :feed_id
	`, sql.Named("feed_id", id)).Scan(&newFailures); err != nil {
		return err
	}

	nextCheckAt := pullpolicy.ComputeNextCheckAtSeconds(
		params.CheckedAt,
		params.IntervalSeconds,
		params.MaxBackoff,
		newFailures,
		params.RetryAfterUntil,
		"",
		0,
	)

	if _, err := tx.Exec(`
		UPDATE feed_fetch_state
		SET next_check_at = :next_check_at, updated_at = unixepoch()
		WHERE feed_id = :feed_id
	`, sql.Named("next_check_at", nextCheckAt), sql.Named("feed_id", id)); err != nil {
		return err
	}

	return tx.Commit()
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

	stmt, err := tx.Prepare(`
		INSERT INTO feeds (group_id, name, link, site_url, proxy)
		VALUES (:group_id, :name, :link, :site_url, '')
		ON CONFLICT(link) DO NOTHING
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	stateStmt, err := tx.Prepare(`
		INSERT INTO feed_fetch_state (feed_id, next_check_at)
		VALUES (:feed_id, unixepoch())
		ON CONFLICT(feed_id) DO UPDATE SET next_check_at = excluded.next_check_at, updated_at = unixepoch()
	`)
	if err != nil {
		return nil, err
	}
	defer stateStmt.Close()

	seenLinks := make(map[string]bool, len(inputs))

	for _, input := range inputs {
		if seenLinks[input.Link] {
			result.Errors = append(result.Errors, fmt.Sprintf("duplicate feed: %s", input.Link))
			continue
		}
		seenLinks[input.Link] = true

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

		affected, err := res.RowsAffected()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to inspect result for %s: %v", input.Link, err))
			continue
		}
		if affected == 0 {
			result.Errors = append(result.Errors, fmt.Sprintf("duplicate feed: %s", input.Link))
			continue
		}

		id, err := res.LastInsertId()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to get id for %s: %v", input.Link, err))
			continue
		}

		if _, err := stateStmt.Exec(sql.Named("feed_id", id)); err != nil {
			return nil, fmt.Errorf("init fetch state for %s: %w", input.Link, err)
		}

		result.Created++
		result.CreatedIDs = append(result.CreatedIDs, id)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return result, nil
}
