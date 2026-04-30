// SPDX-FileCopyrightText: 2026 Pagefault Games
// SPDX-FileContributor: Bertie690
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log/slog"
)

// A Migration represents a single migration to be applied to the database.
type Migration struct {
	// The 0-indexed version of the database to apply to.
	Version int
	// The SQL query to execute when increasing the version.
	Up string
	// The SQL query to execute when decreasing the version.
	Down string
}

//go:embed migrations/00001_add_updated_at_and_versions.up.sql
var migration00001Up string

//go:embed migrations/00001_add_updated_at_and_versions.down.sql
var migration00001Down string

var migrations = []Migration{
	{
		Version: 1,
		Up:      migration00001Up,
		Down:    migration00001Down,
	},
}

// runMigrations runs all pending migrations on the given database store, returning the first error encountered.
// A non-nil error will abort the transaction and result in no changes being applied to the database.
func runMigrations(ctx context.Context, db *sql.DB) error {
	var schemaVersion int
	if err := db.QueryRow(`SELECT version FROM schema_version LIMIT 1`).
		Scan(&schemaVersion); err != nil {
		return fmt.Errorf("failed to get schema version: %w", err)
	}
	initialVersion := schemaVersion

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && rollbackErr != sql.ErrTxDone {
			err = fmt.Errorf("failure to rollback transaction: %w", rollbackErr)
		}
	}()

	for _, migration := range migrations {
		if migration.Version <= schemaVersion {
			continue
		}

		if _, err := tx.ExecContext(ctx, migration.Up); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}

		_, err := tx.ExecContext(ctx, `UPDATE schema_version SET version = ?`, migration.Version)
		if err != nil {
			return fmt.Errorf("failed to update schema version to %d: %w", migration.Version, err)
		}

		schemaVersion = migration.Version
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	if initialVersion != schemaVersion {
		slog.Info(fmt.Sprintf(
			"Successfully migrated database from version %d to %d\n",
			initialVersion,
			schemaVersion,
		))
	}
	return nil
}
