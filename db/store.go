// SPDX-FileCopyrightText: 2026 Pagefault Games
// SPDX-FileContributor: Bertie690
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/pagefaultgames/uxie/utils"
)

// Store wraps a SQLite database that tracks registered help topics and their relevant data.
// TODO: Figure out whether to make names case insensitive/enforce a particular case
type Store struct {
	// The underlying connection to the HelpTopics database.
	db *sql.DB

	// A map containing prepared statements for common queries.
	statements map[statementName]*sql.Stmt

	// TODO: Pass in top level context from main function (for cancellation on shutdown)
}

// statementName is a string enum used for identifying prepared statements.
type statementName string

const (
	// Get a single help topic.
	//  `SELECT id, text, updated_at FROM topics WHERE name = ?`
	getHelpTopic statementName = "getHelpTopic"
	// Get all help topics.
	//  `SELECT id, name, text, updated_at FROM topics`
	getAllTopics statementName = "getAllTopics"
	// Add or update a help topic.
	//  `UPDATE topics
	//  SET text = ?, updated_at = CURRENT_TIMESTAMP(6)
	//  WHERE name = ? AND updated_at = ?`
	addHelpTopic statementName = "addHelpTopic"
	// Delete a help topic.
	//  `DELETE FROM topics WHERE name = ? RETURNING id, text, updated_at`
	deleteTopic statementName = "deleteTopic"
)

// A topicId is a reference to a [HelpTopic]'s rowid, used for clarity of intent.
type topicId = int64

// A HelpTopic represents information about a single help topic stored in the database.
type HelpTopic struct {
	// The topic's name.
	Name string
	// The topic's text.
	Text string
	// The time at which the topic was last updated.
	UpdatedAt time.Time
	// The topic's internal ID.
	id topicId
}

// init initializes the database using the provided context.
func (s *Store) init(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS topics (
			id          INTEGER PRIMARY KEY AUTO_INCREMENT,
			name        VARCHAR(100) NOT NULL UNIQUE,
			text        TEXT NOT NULL,
			updated_at  TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)
				ON UPDATE CURRENT_TIMESTAMP(6),
			CHECK (length(name) > 0 AND length(text) > 0)
		);`); err != nil {
		return err
	}

	if _, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_version (
  			id 			TINYINT NOT NULL PRIMARY KEY,
  			version 	INT NOT NULL,
			CONSTRAINT one_row CHECK (id = 1)
		);`); err != nil {
		return err
	}

	if err := s.prepareStatements(ctx); err != nil {
		return err
	}
	return runMigrations(ctx, s.db)
}

// prepareStatements prepares commonly used SQL statements.
func (s *Store) prepareStatements(ctx context.Context) (err error) {
	s.statements = make(map[statementName]*sql.Stmt, 3)
	s.statements[getHelpTopic], err = s.db.PrepareContext(ctx, `
		SELECT id, text, updated_at FROM topics WHERE name = ?
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare getHelpTopic statement: %w", err)
	}

	s.statements[addHelpTopic], err = s.db.PrepareContext(ctx, `
		  UPDATE topics
		  SET text = ?, updated_at = CURRENT_TIMESTAMP(6)
		  WHERE name = ? AND updated_at = ?
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare addHelpTopic statement: %w", err)
	}

	s.statements[getAllTopics], err = s.db.PrepareContext(ctx, `
		SELECT id, name, text, updated_at FROM topics
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare getAllTopics statement: %w", err)
	}

	s.statements[deleteTopic], err = s.db.PrepareContext(ctx, `
		DELETE FROM topics WHERE name = ? RETURNING id, text, updated_at
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
				fmt.Errorf("failed to close prepared statement %q: %w", name, stmtCloseErr),
			)
		}
	}

	if closeErr := s.db.Close(); closeErr != nil {
		utils.ErrorAttrs("failed to close database", slog.Any("error", closeErr))
		err = errors.Join(err, fmt.Errorf("failed to close database: %w", closeErr))
	}

	return err
}

func (s *Store) getHelpTopic(name string) (topic HelpTopic, err error) {
	var (
		id        topicId
		text      string
		updatedAt time.Time
	)
	err = s.statements[getHelpTopic].QueryRow(name).Scan(&id, &text, &updatedAt)
	if err != nil {
		return HelpTopic{}, fmt.Errorf("failed to get help topic %q: %w", name, err)
	}
	return HelpTopic{
		id:        id,
		Name:      name,
		Text:      text,
		UpdatedAt: updatedAt,
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
			id        topicId
			name      string
			text      string
			updatedAt time.Time
		)
		if serr := rows.Scan(&id, &name, &text, &updatedAt); serr != nil {
			return nil, fmt.Errorf("failed to scan help topic row: %w", serr)
		}
		topics = append(topics, HelpTopic{
			id:        id,
			Name:      name,
			Text:      text,
			UpdatedAt: updatedAt,
		})
	}

	return topics, rows.Err()
}

func (s *Store) addHelpTopic(name, text string) error {
	_, err := s.statements[addHelpTopic].Exec(name, text)
	if err != nil {
		return fmt.Errorf("failed to add help topic %q to database: %w", name, err)
	}

	return nil
}

// deleteTopic deletes the help topic with the given name, returning the deleted topic and any error produced.
// If no topic with the given name exists, an error implementing [sql.ErrNoRows] will be returned.
func (s *Store) deleteTopic(topicName string) (HelpTopic, error) {
	row := s.statements[deleteTopic].QueryRow(topicName)
	var (
		id        topicId
		text      string
		updatedAt time.Time
	)

	if err := row.Scan(&id, &text, &updatedAt); err != nil {
		return HelpTopic{}, fmt.Errorf(
			"failed to delete help topic %q from database: %w",
			topicName,
			err,
		)
	}
	return HelpTopic{
		id:        id,
		Name:      topicName,
		Text:      text,
		UpdatedAt: updatedAt,
	}, nil
}
