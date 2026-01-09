package db

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// HelpTopics is the global help topic database.
var HelpTopics *Store

// Store wraps a SQLite database that tracks registered help topics and their relevant data.
// TODO: Implement "aliases" for topics using a separate linked database
type Store struct {
	db *sql.DB
}

// A TopicId is an opaque reference to a help topic's ID, used for clarity of intent.
type TopicId = int64

const dbPath = "help.db"

type HelpTopic struct {
	ID          TopicId
	Name        string
	Description string
	Text        string
}

// Open creates (or opens) a SQLite database at the default path, initializing it if necessary.
// It will be stored inside [HelpTopics].
func Open() error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	s := &Store{db: db}
	if err := s.init(); err != nil {
		err = errors.Join(err, db.Close())
		return err
	}

	HelpTopics = s
	return nil
}

// Close releases the underlying database connection.
func Close() error {
	if HelpTopics == nil || HelpTopics.db == nil {
		return nil
	}
	return HelpTopics.db.Close()
}

func (s *Store) init() error {
	if _, err := s.db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		return err
	}

	if _, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS topics (
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

// GetHelpTopic retrieves the [HelpTopic] for a given named help topic with the given name.
func (s *Store) GetHelpTopic(name string) (cmd *HelpTopic, err error) {
	var (
		cid        TopicId
		desc, text string
	)
	err = s.db.QueryRow(`
		SELECT description, text, id FROM topics WHERE name = ?
		`, name).Scan(&desc, &text, &cid)
	if err != nil {
		return cmd, err
	}
	cmd = &HelpTopic{
		ID:          cid,
		Name:        name,
		Description: desc,
		Text:        text,
	}
	return cmd, err
}

// AddHelpTopic adds a new help topic to the database.
// If a topic with the same name already exists, it will be overwritten.
func (s *Store) AddHelpTopic(name, description, text string) error {
	_, err := s.db.Exec(`
		INSERT INTO topics (name, description, text) VALUES (?, ?, ?)
		ON CONFLICT DO UPDATE SET description = excluded.description, text = excluded.text`,
		name, description, text)
	if err != nil {
		return fmt.Errorf("failed to add command %s to database: %w", name, err)
	}

	return nil
}

// DeleteTopic deletes the named help topic.
func (s *Store) DeleteTopic(id TopicId) error {
	err := s.db.QueryRow(
		`DELETE FROM topics WHERE id = ? RETURNING id`,
		id).Scan()

	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("command with id %d not found", id)
	} else if err != nil {
		return fmt.Errorf("failed to delete command %d from database: %w", id, err)
	}
	return nil
}
