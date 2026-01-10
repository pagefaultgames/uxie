package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pagefaultgames/oranguru/utils"
)

// HelpTopics is the global help topic database.
var HelpTopics *Store

// Store wraps a SQLite database that tracks registered help topics and their relevant data.
// TODO: Implement "aliases" for topics using a separate linked database
type Store struct {
	// The underlying database connection.
	db *sql.DB

	// A map containing prepared statements for common queries.
	statements map[statementName]*sql.Stmt
}

// statementName is a string enum used for identifying prepared statements.
type statementName string

const (
	// Get a single help topic.
	//  `SELECT description, text, id FROM topics WHERE name = ?1`
	getHelpTopic statementName = "getHelpTopic"
	// Get all help topics.
	//  `SELECT id, name, description, text FROM topics`
	getAllTopics statementName = "getAllTopics"
	// Add or update a help topic.
	//  `INSERT INTO topics (name, description, text) VALUES (?1, ?2, ?3)
	//  ON CONFLICT DO UPDATE SET description = excluded.description, text = excluded.text`
	addHelpTopic statementName = "addHelpTopic"
	// Delete a help topic.
	//  `DELETE FROM topics WHERE name = ?1 RETURNING id`
	deleteTopic statementName = "deleteTopic"
)

// A TopicId is an opaque reference to a help topic's ID, used for clarity of intent.
type TopicId = int64

type HelpTopic struct {
	// The topic's name.
	Name string
	// The topic's description.
	Description string
	// The topic's text.
	Text string
	// The topic's internal ID.
	id TopicId
}

const dbPath = "help.db"

// Open creates (or opens) a SQLite database at the default path, initializing it if necessary and storing it inside [HelpTopics].
// It returns any error encountered.
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

// init initializes the database.
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
	return s.prepareStatements()
}

// prepareStatements prepares all commonly used SQL statements.
func (s *Store) prepareStatements() (err error) {
	s.statements = make(map[statementName]*sql.Stmt, 3)
	s.statements[getHelpTopic], err = s.db.Prepare(`
		SELECT description, text, id FROM topics WHERE name = ?1
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare getHelpTopic statement: %w", err)
	}

	s.statements[addHelpTopic], err = s.db.Prepare(`
		INSERT INTO topics (name, description, text) VALUES (?1, ?2, ?3)
		ON CONFLICT DO UPDATE SET description = excluded.description, text = excluded.text
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare addHelpTopic statement: %w", err)
	}

	s.statements[getAllTopics], err = s.db.Prepare(`
		SELECT id, name, description, text FROM topics
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare getAllTopics statement: %w", err)
	}

	s.statements[deleteTopic], err = s.db.Prepare(`
		DELETE FROM topics WHERE name = ?1 RETURNING id
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare deleteTopic statement: %w", err)
	}
	return nil
}

// Close releases the underlying database connection.
func Close() error {
	if HelpTopics == nil || HelpTopics.db == nil {
		return nil
	}
	var err error
	for name, stmt := range HelpTopics.statements {
		if stmtCloseErr := stmt.Close(); err != nil {
			utils.ErrorAttrs(
				"failed to close prepared statement",
				slog.String("name", string(name)),
				slog.Any("error", stmtCloseErr),
			)
			err = errors.Join(
				err,
				fmt.Errorf("failed to close prepared statement %s: %w", name, stmtCloseErr),
			)
		}
	}

	if closeErr := HelpTopics.db.Close(); closeErr != nil {
		utils.ErrorAttrs("failed to close database", slog.Any("error", closeErr))
		err = errors.Join(err, fmt.Errorf("failed to close database: %w", closeErr))
	}

	return err
}

// GetHelpTopic retrieves the [HelpTopic] for the help topic with the given name.
// If no such topic exists, an error implementing [sql.ErrNoRows] will be returned.
func (s *Store) GetHelpTopic(name string) (topic *HelpTopic, err error) {
	var (
		cid        TopicId
		desc, text string
	)
	err = s.statements[getHelpTopic].QueryRow(name).Scan(&desc, &text, &cid)
	if err != nil {
		return nil, fmt.Errorf("failed to get help topic %s: %w", name, err)
	}
	return &HelpTopic{
		id:          cid,
		Name:        name,
		Description: desc,
		Text:        text,
	}, nil
}

// GetAllTopics retrieves all help topics stored in the database.
func (s *Store) GetAllTopics() (topics []HelpTopic, err error) {
	rows, err := s.statements[getAllTopics].Query()
	if err != nil {
		return nil, fmt.Errorf("failed to query help topics: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			utils.ErrorAttrs("failed to close db rows query", slog.Any("error", cerr))
			err = errors.Join(err, fmt.Errorf("failed to close db rows query: %w", cerr))
		}
	}()

	for rows.Next() {
		var (
			id          TopicId
			name        string
			description string
			text        string
		)
		if serr := rows.Scan(&id, &name, &description, &text); serr != nil {
			return nil, fmt.Errorf("failed to scan help topic row: %w", serr)
		}
		topics = append(topics, HelpTopic{
			id:          id,
			Name:        name,
			Description: description,
			Text:        text,
		})
	}

	return topics, rows.Err()
}

// AddHelpTopic adds a new help topic to the database.
// If a topic with the same name already exists, it will be overwritten.
// (This technically makes it an UPSERT operation.)
func (s *Store) AddHelpTopic(name, description, text string) error {
	_, err := s.statements[addHelpTopic].Exec(name, description, text)
	if err != nil {
		return fmt.Errorf("failed to add help topic %s to database: %w", name, err)
	}

	return nil
}

// DeleteTopic deletes the named help topic, returning whether a topic was deleted and any error produced.
// A nonexistent topic will return false, nil.
func (s *Store) DeleteTopic(name string) (found bool, err error) {
	err = s.statements[deleteTopic].QueryRow(name).Scan()

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to delete help topic %s from database: %w", name, err)
	}
	return true, nil
}
