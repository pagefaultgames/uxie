package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// helpTopics is the global help topic database.
var helpTopics *Store

const dbPath = "help.db"

// Open creates (or opens) a new SQLite database at the default path, initializing it if necessary.
// It returns any error encountered.
func Open(ctx context.Context) error {
	db, err := sql.Open("sqlite3", dbPath)
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

// Internal helper function to get the global DB store.
func getStore() (*Store, error) {
	if helpTopics == nil || helpTopics.db == nil {
		return nil, errNoDB
	}

	return helpTopics, nil
}

// GetHelpTopic retrieves the stored [HelpTopic] with the given name.
// If no such topic exists, an error implementing [sql.ErrNoRows] will be returned.
func GetHelpTopic(name string) (*HelpTopic, error) {
	store, err := getStore()
	if err != nil {
		return nil, err
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

// DeleteTopic deletes the help topic with the given name, returning any error produced.
// If no such topic exists, an error implementing [sql.ErrNoRows] will be returned.
func DeleteTopic(name string) error {
	store, err := getStore()
	if err != nil {
		return err
	}

	return store.deleteTopic(name)
}
