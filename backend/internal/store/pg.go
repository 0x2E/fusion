package store

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// PGStore is the PostgreSQL implementation of Storer.
type PGStore struct {
	db *sql.DB
}

// NewPostgres opens a PostgreSQL connection, pings it, and runs migrations.
func NewPostgres(connString string) (*PGStore, error) {
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	s := &PGStore{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate database: %w", err)
	}

	return s, nil
}

func (s *PGStore) Close() error {
	return s.db.Close()
}
