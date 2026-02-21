package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/patrickjmcd/reedme/internal/model"
)

func (s *Store) ListGroups() ([]*model.Group, error) {
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

func (s *Store) GetGroup(id int64) (*model.Group, error) {
	g := &model.Group{}
	err := s.db.QueryRow(`
		SELECT id, name, created_at, updated_at
		FROM groups
		WHERE id = :id
	`, sql.Named("id", id)).Scan(&g.ID, &g.Name, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: group", ErrNotFound)
		}
		return nil, fmt.Errorf("get group: %w", err)
	}
	return g, nil
}

func (s *Store) CreateGroup(name string) (*model.Group, error) {
	result, err := s.db.Exec(`
		INSERT INTO groups (name) VALUES (:name)
	`, sql.Named("name", name))
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return s.GetGroup(id)
}

func (s *Store) UpdateGroup(id int64, name string) error {
	result, err := s.db.Exec(`
		UPDATE groups
		SET name = :name, updated_at = unixepoch()
		WHERE id = :id
	`, sql.Named("name", name), sql.Named("id", id))
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

// DeleteGroup removes a group and moves all its feeds to the default group (ID=1).
// The default group itself cannot be deleted to ensure all feeds have a valid group.
func (s *Store) DeleteGroup(id int64) error {
	if id == 1 {
		return fmt.Errorf("%w: cannot delete default group", ErrInvalid)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE feeds SET group_id = 1 WHERE group_id = :id`, sql.Named("id", id)); err != nil {
		return err
	}

	result, err := tx.Exec(`DELETE FROM groups WHERE id = :id`, sql.Named("id", id))
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
