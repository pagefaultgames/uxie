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

// Store wraps a MariaDB database that tracks registered help topics and their relevant data.
// TODO: Figure out whether to make names case insensitive/enforce a particular case
type Store struct {
	// The underlying database connection.
	db *sql.DB

	// TODO: Pass in top level context from main function (for cancellation on shutdown)
}

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

	return nil
}

func (s *Store) close() error {
	if err := s.db.Close(); err != nil {
		utils.ErrorAttrs("failed to close database", slog.Any("error", err))
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}

func (s *Store) getHelpTopic(name string) (topic HelpTopic, err error) {
	var (
		id        topicId
		text      string
		updatedAt time.Time
	)
	err = s.db.QueryRow(`
		SELECT id, name, text, updated_at FROM topics WHERE name = ?
	`, name).Scan(&id, &name, &text, &updatedAt)
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
	rows, err := s.db.Query(`
		SELECT id, name, text, updated_at FROM topics
	`)
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

// addHelpTopic adds a new help topic with the given name and text to the database.
// If a topic with the same name already exists, it will remain unchanged and the underlying [mysql.MySQLError] will be returned.
func (s *Store) addHelpTopic(name, text string) error {
	_, err := s.db.Exec(`
		INSERT INTO topics (name, text) VALUES (?, ?)
	`, name, text)
	if err != nil {
		return fmt.Errorf("failed to add help topic %q to database: %w", name, err)
	}

	return nil
}

// updateHelpTopic updates the contents of an existing help topic, using the provided timestamp to verify that the topic has not been modified since it was last retrieved.
// It returns any error produced.
//
// If the topic does not exist in the database, an error implementing [sql.ErrNoRows] will be returned.
// If the topic was modified since the provided timestamp, an [ErrStaleTopic] will be returned.
func (s *Store) updateHelpTopic(
	name, text string,
	expectedUpdatedAt time.Time,
) error {
	result, err := s.db.Exec(`
		  UPDATE topics
		  SET text = ?, updated_at = CURRENT_TIMESTAMP(6)
		  WHERE name = ? AND updated_at = ?
	`, text, name, expectedUpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update help topic %q in database: %w", name, err)
	}

	if rowsAffected, err := result.RowsAffected(); err != nil {
		return fmt.Errorf(
			"failed to get rows affected for help topic update of %q: %w",
			name,
			err,
		)
	} else if rowsAffected > 0 {
		// successful update
		return nil
	}

	// run a quick SELECT on update failure to narrow down nonexistent vs stale (and retrieve updated time for error handling)
	// This WOULD be a great place to use `UPDATE... RETURNING`, but that's only supported for MariaDB 13.0 (which is in preview)

	var lastUpdatedAt time.Time
	if serr := s.db.QueryRow(`
		SELECT updated_at FROM topics WHERE name = ?
	`, name).Scan(&lastUpdatedAt); serr != nil {
		return fmt.Errorf("failed to query for help topic %q last update time: %w", name, serr)
	}

	return ErrStaleTopic{
		DBTopicName:   name,
		LastUpdatedAt: lastUpdatedAt,
	}
}

// deleteTopic deletes the help topic with the given name, returning the deleted topic and any error produced.
// If no topic with the given name exists, an error implementing [sql.ErrNoRows] will be returned.
func (s *Store) deleteTopic(topicName string) (HelpTopic, error) {
	row := s.db.QueryRow(`
		DELETE FROM topics WHERE name = ? RETURNING id, name, text, updated_at
	`, topicName)
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
