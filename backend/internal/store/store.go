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

	"modernc.org/sqlite"
)

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
