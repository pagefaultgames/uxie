package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// Store wraps a SQLite database that tracks registered help commands and their relevant data.
// TODO: Implement "aliases" for commands using a separate linked database
type Store struct {
	db *sql.DB
}

// A CommandId is an opaque reference to a help command's ID, used for clarity of intent.
type CommandId = int64

const dbPath = "help.db"

type Command struct {
	ID          CommandId
	Name        string
	Description string
	Text        string
}

// Open creates (or opens) a SQLite database at the given path, initializing it if necessary.
func Open() (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	s := &Store{db: db}
	if err := s.init(context.Background()); err != nil {
		db.Close()
		return nil, err
	}

	return s, nil
}

// Close releases the underlying database connection.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) init(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `PRAGMA foreign_keys = ON`); err != nil {
		return err
	}

	if _, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS commands (
			id          INTEGER PRIMARY KEY,
			name        VARCHAR NOT NULL UNIQUE,
			description VARCHAR NOT NULL,
			text        VARCHAR NOT NULL,
		);
	`); err != nil {
		return err
	}
	return nil
}

// GetCommand retrieves the [Command] for a given named help command with the given name.
func (s *Store) GetCommand(name string) (cmd *Command, err error) {
	var (
		cid        CommandId
		desc, text string
	)
	err = s.db.QueryRow(`
		SELECT description, text, id FROM commands WHERE name = ?
		`, name).Scan(&desc, &text, &cid)
	if err != nil {
		return cmd, err
	}
	cmd = &Command{
		ID:          cid,
		Name:        name,
		Description: desc,
		Text:        text,
	}
	return cmd, err
}

// HasCommand checks if a command with the given name exists in the database.
func (s *Store) HasCommand(name string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM commands WHERE name = ?)
		`, name).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// AddCommand adds a new help command to the database.
// If a command with the same name already exists, it will be overwritten.
func (s *Store) AddCommand(name, description, text string) error {
	_, err := s.db.Exec(`
		INSERT INTO commands (name, description, text) VALUES (?, ?, ?)
		ON CONFLICT DO UPDATE SET description = excluded.description, text = excluded.text`,
		name, description, text)
	if err != nil {
		return fmt.Errorf("failed to add command %s to database: %w", name, err)
	}

	return nil
}

// DeleteCommand deletes the named command.
func (s *Store) DeleteCommand(id CommandId) error {
	err := s.db.QueryRow(
		`DELETE FROM commands WHERE id = ? RETURNING id`,
		id).Scan()

	if errors.Is(err, sql.ErrNoRows) {
		return 	fmt.Errorf("command with id %d not found", id)
	} else if err != nil {
		return fmt.Errorf("failed to delete command %d from database: %w", id, err)
	}
	return nil
}
