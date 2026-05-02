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
	//  `SELECT id, name, text, updated_at FROM topics WHERE LOWER(name) = LOWER(?)`
	getHelpTopic statementName = "getHelpTopic"
	// Get all help topics.
	//  `SELECT id, name, text, updated_at FROM topics`
	getAllTopics statementName = "getAllTopics"
	// Add a new help topic to the database.
	//  `INSERT INTO topics (name, text, updated_at) VALUES (?, ?, ?)`
	addHelpTopic statementName = "addHelpTopic"
	// Modify an existing help topic in the database, using the provided timestamp to verify correctness.
	//  `UPDATE topics
	//    SET text = ?, updated_at = CURRENT_TIMESTAMP(6)
	//    WHERE LOWER(name) = LOWER(?) AND updated_at = ?`
	updateHelpTopic statementName = "updateHelpTopic"
	// Delete a help topic.
	//  `DELETE FROM topics WHERE LOWER(name) = LOWER(?) RETURNING id, name, text, updated_at`
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
			CHECK (length(name) > 0 AND length(text) > 0)
		);`); err != nil {
		return err
	}

	if _, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_version (
  			id 			TINYINT NOT NULL PRIMARY KEY DEFAULT 1,
  			version 	INT NOT NULL,
			CONSTRAINT one_row CHECK (id = 1)
		);`); err != nil {
		return err
	}

	// populate the version starting at 0 if not already present
	if _, err := s.db.ExecContext(ctx,
		`INSERT IGNORE INTO schema_version SET version = 0;`,
	); err != nil {
		return err
	}

	if err := runMigrations(ctx, s.db); err != nil {
		return fmt.Errorf("failed to run database migrations: %w", err)
	}

	return s.prepareStatements(ctx)
}

// prepareStatements prepares commonly used SQL statements.
func (s *Store) prepareStatements(ctx context.Context) (err error) {
	s.statements = make(map[statementName]*sql.Stmt, 3)
	s.statements[getHelpTopic], err = s.db.PrepareContext(ctx, `
		SELECT id, name, text, updated_at FROM topics WHERE LOWER(name) = LOWER(?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare getHelpTopic statement: %w", err)
	}

	s.statements[addHelpTopic], err = s.db.PrepareContext(ctx, `
		INSERT INTO topics (name, text) VALUES (?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare addHelpTopic statement: %w", err)
	}

	s.statements[updateHelpTopic], err = s.db.PrepareContext(ctx, `
		  UPDATE topics
		  SET text = ?, updated_at = CURRENT_TIMESTAMP(6)
		  WHERE LOWER(name) = LOWER(?) AND updated_at = ?
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare updateHelpTopic statement: %w", err)
	}

	s.statements[getAllTopics], err = s.db.PrepareContext(ctx, `
		SELECT id, name, text, updated_at FROM topics
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare getAllTopics statement: %w", err)
	}

	s.statements[deleteTopic], err = s.db.PrepareContext(ctx, `
		DELETE FROM topics WHERE LOWER(name) = LOWER(?) RETURNING id, name, text, updated_at
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
	err = s.statements[getHelpTopic].QueryRow(name).Scan(&id, &name, &text, &updatedAt)
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

// addHelpTopic adds a new help topic to the database.
// If a topic with the same name already exists, it will NOT be modified.
func (s *Store) addHelpTopic(name, text string) error {
	_, err := s.statements[addHelpTopic].Exec(name, text)
	if err != nil {
		return fmt.Errorf("failed to add help topic %q to database: %w", name, err)
	}

	return nil
}

var ErrStaleUpdate = errors.New("updatedAt time was greater")

// updateHelpTopic updates the contents of an existing help topic, using the provided timestamp to verify that the topic has not been modified since it was last retrieved.
// If the topic was modified since the provided timestamp, no update will be performed and an error implementing [errStaleUpdate] will be returned.
//
// An error implementing [sql.ErrNoRows] will be returned if the topic with the given name does not exist at all.
func (s *Store) updateHelpTopic(name, text string, expectedUpdatedAt time.Time) (err error) {
	result, err := s.statements[updateHelpTopic].Exec(text, name, expectedUpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update help topic %q: %w", name, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for help topic update %q: %w", name, err)
	}

	if rowsAffected == 0 {
		return ErrStaleUpdate
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

	if err := row.Scan(&id, &topicName, &text, &updatedAt); err != nil {
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
