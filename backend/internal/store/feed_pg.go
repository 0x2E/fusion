package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/0x2E/fusion/internal/model"
	"github.com/0x2E/fusion/internal/pullpolicy"
)

func (s *PGStore) ListFeeds() ([]*model.Feed, error) {
	rows, err := s.db.Query(`
		SELECT f.id, f.group_id, f.name, f.link, f.site_url,
		       f.suspended, f.proxy, f.created_at, f.updated_at,
		       COALESCE(fs.etag, ''), COALESCE(fs.last_modified, ''), COALESCE(fs.cache_control, ''),
		       COALESCE(fs.expires_at, 0), COALESCE(fs.last_checked_at, 0), COALESCE(fs.next_check_at, 0),
		       COALESCE(fs.last_http_status, 0), COALESCE(fs.retry_after_until, 0), COALESCE(fs.last_success_at, 0),
		       COALESCE(fs.last_error_at, 0), COALESCE(fs.last_error, ''), COALESCE(fs.consecutive_failures, 0),
		       COALESCE(SUM(CASE WHEN i.unread = true THEN 1 ELSE 0 END), 0) AS unread_count,
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
		if err := rows.Scan(
			&f.ID,
			&f.GroupID,
			&f.Name,
			&f.Link,
			&f.SiteURL,
			&f.Suspended,
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
		feeds = append(feeds, f)
	}
	return feeds, rows.Err()
}

func (s *PGStore) GetFeed(id int64) (*model.Feed, error) {
	f := &model.Feed{}
	err := s.db.QueryRow(`
		SELECT f.id, f.group_id, f.name, f.link, f.site_url,
		       f.suspended, f.proxy, f.created_at, f.updated_at,
		       COALESCE(fs.etag, ''), COALESCE(fs.last_modified, ''), COALESCE(fs.cache_control, ''),
		       COALESCE(fs.expires_at, 0), COALESCE(fs.last_checked_at, 0), COALESCE(fs.next_check_at, 0),
		       COALESCE(fs.last_http_status, 0), COALESCE(fs.retry_after_until, 0), COALESCE(fs.last_success_at, 0),
		       COALESCE(fs.last_error_at, 0), COALESCE(fs.last_error, ''), COALESCE(fs.consecutive_failures, 0)
		FROM feeds f
		LEFT JOIN feed_fetch_state fs ON fs.feed_id = f.id
		WHERE f.id = $1
	`, id).Scan(
		&f.ID,
		&f.GroupID,
		&f.Name,
		&f.Link,
		&f.SiteURL,
		&f.Suspended,
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

	return f, nil
}

func (s *PGStore) CreateFeed(groupID int64, name, link, siteURL, proxy string) (*model.Feed, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var id int64
	if err := tx.QueryRow(`
		INSERT INTO feeds (group_id, name, link, site_url, proxy)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, groupID, name, link, siteURL, proxy).Scan(&id); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(`
		INSERT INTO feed_fetch_state (feed_id, next_check_at)
		VALUES ($1, EXTRACT(EPOCH FROM NOW())::bigint)
	`, id); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s.GetFeed(id)
}

func (s *PGStore) SearchFeeds(query string) ([]*SearchFeedResult, error) {
	rows, err := s.db.Query(`
		SELECT id, name, link, site_url
		FROM feeds
		WHERE name LIKE $1
		ORDER BY id
	`, "%"+query+"%")
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

func (s *PGStore) UpdateFeed(id int64, params UpdateFeedParams) error {
	setClauses := []string{}
	args := []interface{}{}
	paramIdx := 1

	if params.GroupID != nil {
		setClauses = append(setClauses, fmt.Sprintf("group_id = $%d", paramIdx))
		args = append(args, *params.GroupID)
		paramIdx++
	}
	if params.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", paramIdx))
		args = append(args, *params.Name)
		paramIdx++
	}
	if params.Link != nil {
		setClauses = append(setClauses, fmt.Sprintf("link = $%d", paramIdx))
		args = append(args, *params.Link)
		paramIdx++
	}
	if params.SiteURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("site_url = $%d", paramIdx))
		args = append(args, *params.SiteURL)
		paramIdx++
	}
	if params.Suspended != nil {
		setClauses = append(setClauses, fmt.Sprintf("suspended = $%d", paramIdx))
		args = append(args, *params.Suspended)
		paramIdx++
	}
	if params.Proxy != nil {
		setClauses = append(setClauses, fmt.Sprintf("proxy = $%d", paramIdx))
		args = append(args, *params.Proxy)
		paramIdx++
	}

	if len(setClauses) == 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	setClauses = append(setClauses, "updated_at = EXTRACT(EPOCH FROM NOW())::bigint")
	query := fmt.Sprintf("UPDATE feeds SET %s WHERE id = $%d", strings.Join(setClauses, ", "), paramIdx)
	args = append(args, id)

	result, err := tx.Exec(query, args...)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
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
				$1,
				'',
				'',
				'',
				0,
				0,
				EXTRACT(EPOCH FROM NOW())::bigint,
				0,
				0,
				0,
				0,
				'',
				0,
				EXTRACT(EPOCH FROM NOW())::bigint
			)
			ON CONFLICT(feed_id) DO UPDATE SET
				etag = '',
				last_modified = '',
				cache_control = '',
				expires_at = 0,
				last_checked_at = 0,
				next_check_at = EXTRACT(EPOCH FROM NOW())::bigint,
				last_http_status = 0,
				retry_after_until = 0,
				last_success_at = 0,
				last_error_at = 0,
				last_error = '',
				consecutive_failures = 0,
				updated_at = EXTRACT(EPOCH FROM NOW())::bigint
		`, id); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *PGStore) DeleteFeed(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		UPDATE bookmarks SET item_id = NULL
		WHERE item_id IN (SELECT id FROM items WHERE feed_id = $1)
	`, id); err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM items WHERE feed_id = $1`, id); err != nil {
		return err
	}

	result, err := tx.Exec(`DELETE FROM feeds WHERE id = $1`, id)
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

func (s *PGStore) UpdateFeedFetchSuccess(id int64, params UpdateFeedFetchSuccessParams) error {
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
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			0,
			'',
			0,
			EXTRACT(EPOCH FROM NOW())::bigint
		)
		ON CONFLICT(feed_id) DO UPDATE SET
			etag = EXCLUDED.etag,
			last_modified = EXCLUDED.last_modified,
			cache_control = EXCLUDED.cache_control,
			expires_at = EXCLUDED.expires_at,
			last_checked_at = EXCLUDED.last_checked_at,
			next_check_at = EXCLUDED.next_check_at,
			last_http_status = EXCLUDED.last_http_status,
			retry_after_until = EXCLUDED.retry_after_until,
			last_success_at = EXCLUDED.last_success_at,
			last_error_at = 0,
			last_error = '',
			consecutive_failures = 0,
			updated_at = EXTRACT(EPOCH FROM NOW())::bigint
	`,
		id,
		strings.TrimSpace(params.ETag),
		strings.TrimSpace(params.LastModified),
		strings.TrimSpace(params.CacheControl),
		params.ExpiresAt,
		params.CheckedAt,
		params.NextCheckAt,
		params.HTTPStatus,
		params.RetryAfterUntil,
		params.CheckedAt,
	)
	return err
}

func (s *PGStore) UpdateFeedFetchFailure(id int64, params UpdateFeedFetchFailureParams) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// RETURNING consecutive_failures avoids a separate SELECT query.
	var newFailures int64
	if err := tx.QueryRow(`
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
		VALUES ($1, $2, $3, $4, $5, $6, $7, 1, EXTRACT(EPOCH FROM NOW())::bigint)
		ON CONFLICT(feed_id) DO UPDATE SET
			last_checked_at = EXCLUDED.last_checked_at,
			next_check_at = EXCLUDED.next_check_at,
			last_http_status = EXCLUDED.last_http_status,
			retry_after_until = EXCLUDED.retry_after_until,
			last_error_at = EXCLUDED.last_error_at,
			last_error = EXCLUDED.last_error,
			consecutive_failures = feed_fetch_state.consecutive_failures + 1,
			updated_at = EXTRACT(EPOCH FROM NOW())::bigint
		RETURNING consecutive_failures
	`,
		id,
		params.CheckedAt,
		params.CheckedAt,
		params.HTTPStatus,
		params.RetryAfterUntil,
		params.CheckedAt,
		strings.TrimSpace(params.LastError),
	).Scan(&newFailures); err != nil {
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
		SET next_check_at = $1, updated_at = EXTRACT(EPOCH FROM NOW())::bigint
		WHERE feed_id = $2
	`, nextCheckAt, id); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *PGStore) UpdateFeedSiteURLIfEmpty(id int64, siteURL string) error {
	siteURL = strings.TrimSpace(siteURL)
	if siteURL == "" {
		return nil
	}

	_, err := s.db.Exec(`
		UPDATE feeds
		SET site_url = $1, updated_at = EXTRACT(EPOCH FROM NOW())::bigint
		WHERE id = $2 AND (site_url IS NULL OR TRIM(site_url) = '')
	`, siteURL, id)
	return err
}

func (s *PGStore) BatchCreateFeeds(inputs []BatchCreateFeedsInput) (*BatchCreateFeedsResult, error) {
	result := &BatchCreateFeedsResult{}

	if len(inputs) == 0 {
		return result, nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// RETURNING id retrieves the inserted row's id; returns no rows on conflict.
	stmt, err := tx.Prepare(`
		INSERT INTO feeds (group_id, name, link, site_url, proxy)
		VALUES ($1, $2, $3, $4, '')
		ON CONFLICT(link) DO NOTHING
		RETURNING id
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	stateStmt, err := tx.Prepare(`
		INSERT INTO feed_fetch_state (feed_id, next_check_at)
		VALUES ($1, EXTRACT(EPOCH FROM NOW())::bigint)
		ON CONFLICT(feed_id) DO UPDATE SET
			next_check_at = EXCLUDED.next_check_at,
			updated_at = EXTRACT(EPOCH FROM NOW())::bigint
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

		var id int64
		err := stmt.QueryRow(input.GroupID, input.Name, input.Link, input.SiteURL).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			result.Errors = append(result.Errors, fmt.Sprintf("duplicate feed: %s", input.Link))
			continue
		}
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to create %s: %v", input.Link, err))
			continue
		}

		if _, err := stateStmt.Exec(id); err != nil {
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
