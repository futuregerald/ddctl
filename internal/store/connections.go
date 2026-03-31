// internal/store/connections.go
package store

import (
	"database/sql"
	"fmt"
	"time"
)

type Connection struct {
	Name      string
	Site      string
	IsDefault bool
	CreatedAt time.Time
}

func (s *Store) AddConnection(name, site string) error {
	// Check if this is the first connection
	var count int
	if err := s.db.QueryRow("SELECT count(*) FROM connections").Scan(&count); err != nil {
		return err
	}
	isDefault := count == 0

	_, err := s.db.Exec(
		"INSERT INTO connections (name, site, is_default) VALUES (?, ?, ?)",
		name, site, isDefault,
	)
	if err != nil {
		return fmt.Errorf("adding connection %q: %w", name, err)
	}
	return nil
}

func (s *Store) GetConnection(name string) (*Connection, error) {
	conn := &Connection{}
	err := s.db.QueryRow(
		"SELECT name, site, is_default, created_at FROM connections WHERE name = ?",
		name,
	).Scan(&conn.Name, &conn.Site, &conn.IsDefault, &conn.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("connection %q not found", name)
	}
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (s *Store) GetDefaultConnection() (*Connection, error) {
	conn := &Connection{}
	err := s.db.QueryRow(
		"SELECT name, site, is_default, created_at FROM connections WHERE is_default = 1",
	).Scan(&conn.Name, &conn.Site, &conn.IsDefault, &conn.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no default connection configured")
	}
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (s *Store) ListConnections() ([]Connection, error) {
	rows, err := s.db.Query("SELECT name, site, is_default, created_at FROM connections ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conns []Connection
	for rows.Next() {
		var c Connection
		if err := rows.Scan(&c.Name, &c.Site, &c.IsDefault, &c.CreatedAt); err != nil {
			return nil, err
		}
		conns = append(conns, c)
	}
	return conns, rows.Err()
}

func (s *Store) SetDefaultConnection(name string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Verify connection exists
	var exists int
	if err := tx.QueryRow("SELECT count(*) FROM connections WHERE name = ?", name).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("connection %q not found", name)
	}

	// Clear existing default
	if _, err := tx.Exec("UPDATE connections SET is_default = 0"); err != nil {
		return err
	}
	// Set new default
	if _, err := tx.Exec("UPDATE connections SET is_default = 1 WHERE name = ?", name); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) RemoveConnection(name string) error {
	result, err := s.db.Exec("DELETE FROM connections WHERE name = ?", name)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("connection %q not found", name)
	}
	return nil
}
