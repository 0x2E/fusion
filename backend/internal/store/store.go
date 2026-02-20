// Package store provides data access layer for Fusion RSS reader.
//
// All timestamps are stored as Unix epoch seconds (INTEGER in SQLite).
// Boolean fields are stored as INTEGER (0/1) and converted to/from Go bool.
// Named SQL parameters (:param_name) are used throughout for safety.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/0x2E/fusion/internal/model"
	"modernc.org/sqlite"
)

// Storer is the interface implemented by both Store (SQLite) and PGStore (PostgreSQL).
// All handler and puller code depends on this interface, not on a concrete type.
type Storer interface {
	Close() error

	ListFeeds() ([]*model.Feed, error)
	GetFeed(id int64) (*model.Feed, error)
	CreateFeed(groupID int64, name, link, siteURL, proxy string) (*model.Feed, error)
	SearchFeeds(query string) ([]*SearchFeedResult, error)
	UpdateFeed(id int64, params UpdateFeedParams) error
	DeleteFeed(id int64) error
	UpdateFeedFetchSuccess(id int64, params UpdateFeedFetchSuccessParams) error
	UpdateFeedFetchFailure(id int64, params UpdateFeedFetchFailureParams) error
	UpdateFeedSiteURLIfEmpty(id int64, siteURL string) error
	BatchCreateFeeds(inputs []BatchCreateFeedsInput) (*BatchCreateFeedsResult, error)

	ListItems(params ListItemsParams) ([]*model.Item, error)
	GetItem(id int64) (*model.Item, error)
	CreateItem(feedID int64, guid, title, link, content string, pubDate int64) (*model.Item, error)
	BatchCreateItemsIgnore(feedID int64, inputs []BatchCreateItemInput) (int, error)
	UpdateItemUnread(id int64, unread bool) error
	BatchUpdateItemsUnread(ids []int64, unread bool) error
	MarkAllAsRead(feedID *int64) error
	ItemExists(feedID int64, guid string) (bool, error)
	SearchItems(query string, limit int) ([]*SearchItemResult, error)
	CountItems(params ListItemsParams) (int, error)

	ListGroups() ([]*model.Group, error)
	GetGroup(id int64) (*model.Group, error)
	CreateGroup(name string) (*model.Group, error)
	UpdateGroup(id int64, name string) error
	DeleteGroup(id int64) error

	ListBookmarks(limit, offset int) ([]*model.Bookmark, error)
	GetBookmark(id int64) (*model.Bookmark, error)
	CreateBookmark(itemID *int64, link, title, content string, pubDate int64, feedName string) (*model.Bookmark, error)
	DeleteBookmark(id int64) error
	BookmarkExists(link string) (bool, error)
	CountBookmarks() (int, error)
}

type Store struct {
	db *sql.DB
}

var sqliteHookOnce sync.Once

func New(dbPath string) (*Store, error) {
	sqliteHookOnce.Do(func() {
		sqlite.RegisterConnectionHook(func(conn sqlite.ExecQuerierContext, _ string) error {
			ctx := context.Background()
			if _, err := conn.ExecContext(ctx, "PRAGMA foreign_keys = ON", nil); err != nil {
				return fmt.Errorf("enable foreign_keys: %w", err)
			}
			if _, err := conn.ExecContext(ctx, "PRAGMA busy_timeout = 5000", nil); err != nil {
				return fmt.Errorf("set busy_timeout: %w", err)
			}
			if _, err := conn.ExecContext(ctx, "PRAGMA journal_mode = WAL", nil); err != nil {
				return fmt.Errorf("set journal_mode: %w", err)
			}
			return nil
		})
	})

	if err := prepareLegacyDatabase(dbPath); err != nil {
		return nil, fmt.Errorf("prepare legacy database: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate database: %w", err)
	}

	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
