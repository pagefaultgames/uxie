package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/pagefaultgames/oranguru/utils"
)

// Store wraps a SQLite database that tracks registered help topics and their relevant data.
// TODO: Implement "aliases" for topics using a separate linked database?
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
	//  `SELECT id, text FROM topics WHERE name = ?1`
	getHelpTopic statementName = "getHelpTopic"
	// Get all help topics.
	//  `SELECT id, name, text FROM topics`
	getAllTopics statementName = "getAllTopics"
	// Add or update a help topic.
	//  `INSERT INTO topics (name, text) VALUES (?1, ?2)
	//   ON CONFLICT DO UPDATE SET text = excluded.text`
	addHelpTopic statementName = "addHelpTopic"
	// Delete a help topic.
	//  `DELETE FROM topics WHERE name = ?1 RETURNING id`
	deleteTopic statementName = "deleteTopic"
)

// A topicId is a reference to a [HelpTopic]'s rowid, used for clarity of intent.
type topicId = int64

// A HelpTopic represents a single help topic stored in the database.
type HelpTopic struct {
	// The topic's name.
	Name string
	// The topic's text.
	Text string
	// The topic's internal ID.
	id topicId
}

// init initializes the database using the provided context.
func (s *Store) init(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS topics (
			id          INTEGER PRIMARY KEY,
			name        VARCHAR NOT NULL UNIQUE,
			text        VARCHAR NOT NULL
		);`); err != nil {
		return err
	}
	return s.prepareStatements(ctx)
}

// prepareStatements prepares commonly used SQL statements.
func (s *Store) prepareStatements(ctx context.Context) (err error) {
	s.statements = make(map[statementName]*sql.Stmt, 3)
	s.statements[getHelpTopic], err = s.db.PrepareContext(ctx, `
		SELECT  id, text FROM topics WHERE name = ?1
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare getHelpTopic statement: %w", err)
	}

	s.statements[addHelpTopic], err = s.db.PrepareContext(ctx, `
		INSERT INTO topics (name, text) VALUES (?1, ?2)
		ON CONFLICT DO UPDATE SET text = excluded.text
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare addHelpTopic statement: %w", err)
	}

	s.statements[getAllTopics], err = s.db.PrepareContext(ctx, `
		SELECT id, name, text FROM topics
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare getAllTopics statement: %w", err)
	}

	s.statements[deleteTopic], err = s.db.PrepareContext(ctx, `
		DELETE FROM topics WHERE name = ?1 RETURNING id
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare deleteTopic statement: %w", err)
	}
	return nil
}

func (s *Store) close() (err error) {
	for name, stmt := range s.statements {
		if stmtCloseErr := stmt.Close(); err != nil {
			utils.ErrorAttrs(
				"Failed to close prepared statement",
				slog.String("name", string(name)),
				slog.Any("error", stmtCloseErr),
			)
			err = errors.Join(
				err,
				fmt.Errorf("failed to close prepared statement %s: %w", name, stmtCloseErr),
			)
		}
	}

	if closeErr := s.db.Close(); closeErr != nil {
		utils.ErrorAttrs("failed to close database", slog.Any("error", closeErr))
		err = errors.Join(err, fmt.Errorf("failed to close database: %w", closeErr))
	}

	return err
}

func (s *Store) getHelpTopic(name string) (topic *HelpTopic, err error) {
	var (
		cid  topicId
		text string
	)
	err = s.statements[getHelpTopic].QueryRow(name).Scan(&cid, &text)
	if err != nil {
		return nil, fmt.Errorf("failed to get help topic %s: %w", name, err)
	}
	return &HelpTopic{
		id:   cid,
		Name: name,
		Text: text,
	}, nil
}

func (s *Store) getAllTopics() (topics []HelpTopic, err error) {
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
			id   topicId
			name string
			text string
		)
		if serr := rows.Scan(&id, &name, &text); serr != nil {
			return nil, fmt.Errorf("failed to scan help topic row: %w", serr)
		}
		topics = append(topics, HelpTopic{
			id:   id,
			Name: name,
			Text: text,
		})
	}

	return topics, rows.Err()
}

func (s *Store) addHelpTopic(name, text string) error {
	_, err := s.statements[addHelpTopic].Exec(name, text)
	if err != nil {
		return fmt.Errorf("failed to add help topic %s to database: %w", name, err)
	}

	return nil
}

func (s *Store) deleteTopic(name string) error {
	if err := s.statements[deleteTopic].QueryRow(name).Scan(); err != nil {
		return fmt.Errorf("failed to delete help topic %s from database: %w", name, err)
	}
	return nil
}
