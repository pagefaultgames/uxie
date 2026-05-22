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

	mysql "github.com/go-sql-driver/mysql"
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
	// The topic's name; is guaranteed to be nonempty.
	Name string
	// The topic's text.
	Text string
	// The time at which the topic was last updated.
	UpdatedAt time.Time
	// Whether the topic should be displayed without its title.
	OmitTitle bool

	// The topic's internal ID.
	id topicId
}

// TopicAlias represents a user-facing alias for a help topic.
type TopicAlias struct {
	AliasName string
	TopicName string
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

// getHelpTopic retrieves the help topic with the given name, if it exists.
// If no topic exists with the given name, an error implementing [sql.ErrNoRows] will be returned.
func (s *Store) getHelpTopic(name string) (topic HelpTopic, err error) {
	err = s.db.QueryRow(`
		SELECT id, name, text, updated_at, omit_title FROM topics WHERE name = ?
	`, name).Scan(&topic.id, &topic.Name, &topic.Text, &topic.UpdatedAt, &topic.OmitTitle)
	if err != nil {
		return HelpTopic{}, fmt.Errorf("failed to get help topic %q: %w", name, err)
	}

	return topic, nil
}

func (s *Store) getAllTopics() (topics []HelpTopic, err error) {
	rows, err := s.db.Query(`
		SELECT id, name, text, updated_at, omit_title FROM topics
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
		var topic HelpTopic
		if serr := rows.Scan(
			&topic.id,
			&topic.Name,
			&topic.Text,
			&topic.UpdatedAt,
			&topic.OmitTitle,
		); serr != nil {
			return nil, fmt.Errorf("failed to scan help topic row: %w", serr)
		}
		topics = append(topics, topic)
	}

	return topics, rows.Err()
}

func (s *Store) getAllAliases() (aliases []TopicAlias, err error) {
	rows, err := s.db.Query(`
		SELECT a.alias_name, t.name
		FROM topic_aliases a
		JOIN topics t ON t.id = a.topic_id
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query topic aliases: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			utils.ErrorAttrs("failed to close db rows query", slog.Any("error", cerr))
			err = errors.Join(err, fmt.Errorf("failed to close db rows query: %w", cerr))
		}
	}()

	for rows.Next() {
		var alias TopicAlias
		if serr := rows.Scan(&alias.AliasName, &alias.TopicName); serr != nil {
			return nil, fmt.Errorf("failed to scan topic alias row: %w", serr)
		}
		aliases = append(aliases, alias)
	}

	return aliases, rows.Err()
}

// addHelpTopic adds a new help topic with the given name and text to the database.
// If a topic with the same name already exists, it will remain unchanged and the underlying [mysql.MySQLError] will be returned.
func (s *Store) addHelpTopic(name, text string, omitTitle bool) error {
	_, err := s.db.Exec(`
		INSERT INTO topics (name, text, omit_title) VALUES (?, ?, ?)
	`, name, text, omitTitle)
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
	omitTitle bool,
) error {
	result, err := s.db.Exec(`
		  UPDATE topics
		  SET text = ?, omit_title = ?, updated_at = CURRENT_TIMESTAMP(6)
		  WHERE name = ? AND updated_at = ?
	`, text, omitTitle, name, expectedUpdatedAt)
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
func (s *Store) deleteTopic(topicName string) (deleted HelpTopic, err error) {
	err = s.db.QueryRow(`
		DELETE FROM topics WHERE name = ? RETURNING id, name, text, updated_at, omit_title
	`, topicName).Scan(&deleted.id, &deleted.Name, &deleted.Text, &deleted.UpdatedAt, &deleted.OmitTitle)
	if err != nil {
		return HelpTopic{}, fmt.Errorf(
			"failed to delete help topic %q from database: %w",
			topicName,
			err,
		)
	}
	return deleted, nil
}

// ErrDuplicateAlias represents the error produced when attempting to create a help topic alias whose name conflicts with an existing alias or topic.
type ErrDuplicateAlias struct {
	// The name of the alias that caused the conflict.
	AliasName string
	// The topic to which the existing conflicting alias points, if applicable.
	// Will be empty if the conflict is with an existing topic name rather than alias.
	OtherAliasTarget string
}

// Error implements the error interface.
func (e ErrDuplicateAlias) Error() string {
	if e.OtherAliasTarget == "" {
		return fmt.Sprintf(
			"a help topic with name %q already exists",
			e.AliasName,
		)
	}
	return fmt.Sprintf(
		"an alias with name %q already exists for topic %q",
		e.AliasName,
		e.OtherAliasTarget,
	)
}

// addAlias adds a new alias for a given topic.
// If the alias conflicts with an existing topic name, an [ErrDuplicateAlias] will be returned.
func (s *Store) addAlias(topicName, aliasName string) error {
	// Check that alias doesn't conflict with existing topics
	var topicExists bool
	err := s.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM topics WHERE name = ?)
	`, aliasName).Scan(&topicExists)
	if err != nil {
		return fmt.Errorf("failed to check for conflicting topic name %q: %w", aliasName, err)
	}

	if topicExists {
		return ErrDuplicateAlias{AliasName: aliasName}
	}

	existing, err := s.getTopicByAlias(aliasName)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to check for existing alias %q: %w", aliasName, err)
	}
	if existing.Name != "" {
		return ErrDuplicateAlias{AliasName: aliasName, OtherAliasTarget: existing.Name}
	}

	result, err := s.db.Exec(`
		INSERT INTO topic_aliases (topic_id, alias_name)
			SELECT id, ?
			FROM topics
			WHERE name = ?;
	`, aliasName, topicName)
	if errors.Is(err, &mysql.MySQLError{Number: 1062}) {
		return ErrDuplicateAlias{AliasName: aliasName}
	} else if err != nil {
		return fmt.Errorf("failed to add alias %q for topic %q: %w", aliasName, topicName, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"failed to get rows affected for alias %q insertion: %w",
			aliasName,
			err,
		)
	}
	if rowsAffected == 0 {
		return fmt.Errorf(
			"failed to add alias %q for topic %q: %w",
			aliasName,
			topicName,
			sql.ErrNoRows,
		)
	}

	return nil
}

// deleteAlias deletes the alias with the given name.
// If no alias with the given name exists, an error implementing [sql.ErrNoRows] will be returned.
func (s *Store) deleteAlias(aliasName string) error {
	result, err := s.db.Exec(`
		DELETE FROM topic_aliases WHERE alias_name = ?
	`, aliasName)
	if err != nil {
		return fmt.Errorf("failed to delete alias %q: %w", aliasName, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for alias deletion %q: %w", aliasName, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("failed to delete alias %q: %w", aliasName, sql.ErrNoRows)
	}

	return nil
}

// getTopicByAlias retrieves a help topic by one of its aliases.
// If no topic with the given alias exists, an error implementing [sql.ErrNoRows] will be returned.
func (s *Store) getTopicByAlias(aliasName string) (topic HelpTopic, err error) {
	err = s.db.QueryRow(`
		SELECT t.id, t.name, t.text, t.updated_at, t.omit_title
		FROM topics t
		JOIN topic_aliases a ON t.id = a.topic_id
		WHERE a.alias_name = ?
	`, aliasName).Scan(&topic.id, &topic.Name, &topic.Text, &topic.UpdatedAt, &topic.OmitTitle)
	if err != nil {
		return HelpTopic{}, fmt.Errorf("failed to get help topic for alias %q: %w", aliasName, err)
	}

	return topic, nil
}
