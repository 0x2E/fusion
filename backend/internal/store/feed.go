package store

import (
	"database/sql"
	"fmt"

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

	var feeds []*model.Feed
	for rows.Next() {
		f := &model.Feed{}
		var suspended int
		if err := rows.Scan(&f.ID, &f.GroupID, &f.Name, &f.Link, &f.SiteURL, &f.LastBuild, &f.Failure, &f.Failures, &suspended, &f.Proxy, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		f.Suspended = suspended != 0
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
	f.Suspended = suspended != 0
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

type UpdateFeedParams struct {
	GroupID   *int64
	Name      *string
	SiteURL   *string
	Suspended *bool
	Proxy     *string
}

// FIX build sql based on params instead of multiple executions
func (s *Store) UpdateFeed(id int64, params UpdateFeedParams) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if params.GroupID != nil {
		if _, err := tx.Exec(`UPDATE feeds SET group_id = :group_id, updated_at = unixepoch() WHERE id = :id`,
			sql.Named("group_id", *params.GroupID), sql.Named("id", id)); err != nil {
			return err
		}
	}
	if params.Name != nil {
		if _, err := tx.Exec(`UPDATE feeds SET name = :name, updated_at = unixepoch() WHERE id = :id`,
			sql.Named("name", *params.Name), sql.Named("id", id)); err != nil {
			return err
		}
	}
	if params.SiteURL != nil {
		if _, err := tx.Exec(`UPDATE feeds SET site_url = :site_url, updated_at = unixepoch() WHERE id = :id`,
			sql.Named("site_url", *params.SiteURL), sql.Named("id", id)); err != nil {
			return err
		}
	}
	if params.Suspended != nil {
		suspended := 0
		if *params.Suspended {
			suspended = 1
		}
		if _, err := tx.Exec(`UPDATE feeds SET suspended = :suspended, updated_at = unixepoch() WHERE id = :id`,
			sql.Named("suspended", suspended), sql.Named("id", id)); err != nil {
			return err
		}
	}
	if params.Proxy != nil {
		if _, err := tx.Exec(`UPDATE feeds SET proxy = :proxy, updated_at = unixepoch() WHERE id = :id`,
			sql.Named("proxy", *params.Proxy), sql.Named("id", id)); err != nil {
			return err
		}
	}

	return tx.Commit()
}

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
