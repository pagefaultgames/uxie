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

	_ "github.com/go-sql-driver/mysql"
)

// helpTopics is the global help topic database.
var helpTopics *Store

// Open creates (or opens) a new MySQL database at the given path, initializing it if necessary.
// It returns any error encountered.
func Open(ctx context.Context, dbPath string) error {
	db, err := sql.Open("mysql", dbPath)
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

// getStore is an tnternal helper function to get the global DB store.
func getStore() (*Store, error) {
	if helpTopics == nil || helpTopics.db == nil {
		return nil, errNoDB
	}

	return helpTopics, nil
}

// GetHelpTopic retrieves the stored [HelpTopic] with the given name.
// If no such topic exists, an error implementing [sql.ErrNoRows] will be returned.
func GetHelpTopic(name string) (HelpTopic, error) {
	store, err := getStore()
	if err != nil {
		return HelpTopic{}, err
	}

	return store.getHelpTopic(name)
}

// GetAllTopics retrieves all help topics stored in the database.
func GetAllTopics() ([]HelpTopic, error) {
	store, err := getStore()
	if err != nil {
		return nil, err
	}

	return store.getAllTopics()
}

// AddHelpTopic adds a new help topic to the database.
// If a topic with the same name already exists, it will be overwritten.
// (This technically makes it an UPSERT operation.)
func AddHelpTopic(name, text string) error {
	store, err := getStore()
	if err != nil {
		return err
	}

	return store.addHelpTopic(name, text)
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
