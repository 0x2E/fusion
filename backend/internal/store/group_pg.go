package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/patrickjmcd/reedme/internal/model"
)

func (s *PGStore) ListGroups() ([]*model.Group, error) {
	rows, err := s.db.Query(`
		SELECT id, name, created_at, updated_at
		FROM groups
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := []*model.Group{}
	for rows.Next() {
		g := &model.Group{}
		if err := rows.Scan(&g.ID, &g.Name, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

func (s *PGStore) GetGroup(id int64) (*model.Group, error) {
	g := &model.Group{}
	err := s.db.QueryRow(`
		SELECT id, name, created_at, updated_at
		FROM groups
		WHERE id = $1
	`, id).Scan(&g.ID, &g.Name, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: group", ErrNotFound)
		}
		return nil, fmt.Errorf("get group: %w", err)
	}
	return g, nil
}

func (s *PGStore) CreateGroup(name string) (*model.Group, error) {
	var id int64
	if err := s.db.QueryRow(`
		INSERT INTO groups (name) VALUES ($1) RETURNING id
	`, name).Scan(&id); err != nil {
		return nil, err
	}

	return s.GetGroup(id)
}

func (s *PGStore) UpdateGroup(id int64, name string) error {
	result, err := s.db.Exec(`
		UPDATE groups
		SET name = $1, updated_at = EXTRACT(EPOCH FROM NOW())::bigint
		WHERE id = $2
	`, name, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("%w: group", ErrNotFound)
	}
	return nil
}

func (s *PGStore) DeleteGroup(id int64) error {
	if id == 1 {
		return fmt.Errorf("%w: cannot delete default group", ErrInvalid)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE feeds SET group_id = 1 WHERE group_id = $1`, id); err != nil {
		return err
	}

	result, err := tx.Exec(`DELETE FROM groups WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("%w: group", ErrNotFound)
	}

	return tx.Commit()
}
