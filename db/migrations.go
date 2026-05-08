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

// A migration represents a single migration to be applied to the database.
type migration struct {
	// The name of the migration; used for historical purposes and logging, but is otherwise unused.
	name string
	// The 0-indexed version of the database to apply to.
	version int
	// The SQL query to execute when increasing the version.
	up string
	// The SQL query to execute when decreasing the version.
	down string
}

//go:embed migrations/00001_add_updated_at_and_versions.up.sql
var migration00001Up string

//go:embed migrations/00001_add_updated_at_and_versions.down.sql
var migration00001Down string

//go:embed migrations/00002_add_aliases.up.sql
var migration00002Up string

//go:embed migrations/00002_add_aliases.down.sql
var migration00002Down string

var migrations = []migration{
	{
		name:    "Add updated_at column and schema versioning",
		version: 1,
		up:      migration00001Up,
		down:    migration00001Down,
	},
	{
		name:    "Add aliases table",
		version: 2,
		up:      migration00002Up,
		down:    migration00002Down,
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
		if migration.version <= schemaVersion {
			continue
		}

		if _, err := tx.ExecContext(ctx, migration.up); err != nil {
			return fmt.Errorf(
				"failed to apply migration %q (v%d): %w",
				migration.name,
				migration.version,
				err,
			)
		}

		_, err := tx.ExecContext(ctx, `UPDATE schema_version SET version = ?`, migration.version)
		if err != nil {
			return fmt.Errorf("failed to update schema version to %d: %w", migration.version, err)
		}

		schemaVersion = migration.version
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
