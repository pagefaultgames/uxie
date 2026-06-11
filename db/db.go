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
	"time"

	mysql "github.com/go-sql-driver/mysql"
)

// helpTopics is the global help topic database.
var helpTopics *Store

// Open creates (or opens) a new MySQL database using the given DSN, initializing it if necessary.
// It returns any error encountered.
//
// Note that the provided DSN is expected to set `parseTime=true` and `loc=UTC` to allow for proper timestamp parsing.
func Open(ctx context.Context, dsn string) error {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	helpTopics = &Store{db: db}
	if err := helpTopics.init(ctx); err != nil {
		err = errors.Join(fmt.Errorf("failed to initialize database: %w", err), db.Close())
		return err
	}

	return nil
}

// Close releases the underlying database connection.
func Close() error {
	if helpTopics == nil || helpTopics.db == nil {
		return nil
	}

	err := helpTopics.close()
	helpTopics = nil
	return err
}

var errNoDB = errors.New("database not initialized")

// getStore is an internal helper function to get the global DB store.
func getStore() (*Store, error) {
	if helpTopics == nil || helpTopics.db == nil {
		return nil, errNoDB
	}

	return helpTopics, nil
}

// GetHelpTopic retrieves the stored [HelpTopic] with the given name or alias.
//
// It returns the topic and any error produced during retrieval.
// Callers _must_ check the error before using the returned topic.
//
// If no topic with the given name exists, an error implementing [sql.ErrNoRows] will be returned.
func GetHelpTopic(name string) (topic HelpTopic, err error) {
	store, err := getStore()
	if err != nil {
		return HelpTopic{}, err
	}

	// check directly first
	if topic, err := store.getHelpTopic(name); err == nil || !errors.Is(err, sql.ErrNoRows) {
		return topic, err
	}

	// no rows matched; check aliases
	return store.getTopicByAlias(name)
}

// GetAllTopics retrieves all help topics stored in the database.
//
// It returns the list of topics and any error produced during retrieval.
func GetAllTopics() (topics []HelpTopic, err error) {
	store, err := getStore()
	if err != nil {
		return nil, err
	}

	return store.getAllTopics()
}

// GetAllAliases retrieves all help topic aliases stored in the database.
func GetAllAliases() ([]TopicAlias, error) {
	store, err := getStore()
	if err != nil {
		return nil, err
	}

	return store.getAllAliases()
}

// ErrStaleTopic represents an error produced by inserting a topic that was modified since last retrieved.
//
// It does not contain the expected timestamp value as callers are assumed to already have access to said value.
type ErrStaleTopic struct {
	// The name of the database topic for which an update was attempted.
	DBTopicName string
	// The time of the help topic's last modification within the database.
	LastUpdatedAt time.Time
}

// Error returns a string representation of the error to implement the error interface.
// Callers are encouraged to perform custom handling based on LastUpdatedAt as needed.
func (e ErrStaleTopic) Error() string {
	return fmt.Sprintf(
		"help topic %q was modified at %s",
		e.DBTopicName,
		e.LastUpdatedAt.Format(time.RFC1123),
	)
}

// UpsertHelpTopic inserts or updates a help topic in the database, using the provided timestamp for optimistic update locking.
// It returns whether the topic was inserted or updated, and any error produced during either operation.
//
// If an existing topic was modified since lastUpdatedAt, an [ErrStaleTopic] (which can be matched with [errors.AsType]) will be returned.
func UpsertHelpTopic(
	topicName, text string,
	lastUpdatedAt time.Time,
	omitTitle bool,
) (inserted bool, err error) {
	store, err := getStore()
	if err != nil {
		return false, err
	}

	// try insert
	err = store.addHelpTopic(topicName, text, omitTitle)
	if err == nil {
		return true, nil
	} else if !errors.Is(err, &mysql.MySQLError{Number: 1062}) { // duplicate key error code
		return false, fmt.Errorf("failed to add help topic to database: %w", err)
	}

	// update if insert failed due to duplicate key
	return false, store.updateHelpTopic(topicName, text, lastUpdatedAt, omitTitle)
}

// RenameTopic renames an existing help topic, returning any error produced during the operation.
// If no topic with the given name exists, an error implementing [sql.ErrNoRows] will be returned.
// If the new name conflicts with an existing topic, an error implementing [ErrDuplicateTopicName] will be returned.
func RenameTopic(oldName, newName string) error {
	store, err := getStore()
	if err != nil {
		return err
	}

	return store.renameTopic(oldName, newName)
}

// DeleteTopic deletes the help topic with the given name, returning the deleted topic and any error produced.
// If no topic with the given name exists, an error implementing [sql.ErrNoRows] will be returned.
func DeleteTopic(topicName string) (HelpTopic, error) {
	store, err := getStore()
	if err != nil {
		return HelpTopic{}, err
	}

	return store.deleteTopic(topicName)
}

// AddAlias adds a new alias for a given topic.
// If the alias conflicts with an existing topic name, an [ErrDuplicateAlias] will be returnedA
func AddAlias(topicName, aliasName string) error {
	store, err := getStore()
	if err != nil {
		return err
	}

	return store.addAlias(topicName, aliasName)
}

// DeleteAlias deletes a help topic alias by name.
// If no alias with the given name exists, an error implementing [sql.ErrNoRows] will be returned.
func DeleteAlias(aliasName string) error {
	store, err := getStore()
	if err != nil {
		return err
	}

	return store.deleteAlias(aliasName)
}
