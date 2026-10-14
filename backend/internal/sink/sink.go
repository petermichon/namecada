// Package sink persists engine results to durable stores. The SQLite sink
// writes one row per provider:label into a checks table and exposes an
// available view over the free ones.
package sink

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/petermichon/domain-browser/backend/internal/engine"

	_ "modernc.org/sqlite"
)

// schema creates the checks table and the available view. name + provider is
// the primary key, so re-checking a label updates its rows in place.
const schema = `
CREATE TABLE IF NOT EXISTS checks (
	name       TEXT NOT NULL,
	provider   TEXT NOT NULL,
	platform   TEXT NOT NULL,
	status     TEXT NOT NULL,
	detail     TEXT NOT NULL DEFAULT '',
	url        TEXT NOT NULL,
	checked_at TEXT NOT NULL,
	PRIMARY KEY (name, provider)
);

CREATE VIEW IF NOT EXISTS available AS
	SELECT name, provider, platform, url, checked_at
	FROM checks
	WHERE status = 'available';
`

// upsert records a result, replacing the row with the same name + provider.
const upsert = `
INSERT INTO checks (name, provider, platform, status, detail, url, checked_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(name, provider) DO UPDATE SET
	platform   = excluded.platform,
	status     = excluded.status,
	detail     = excluded.detail,
	url        = excluded.url,
	checked_at = excluded.checked_at;
`

// SQLite is a Store that writes results to a SQLite database file.
type SQLite struct {
	db  *sql.DB
	now func() time.Time
}

// OpenSQLite opens (creating if needed) the SQLite database at path and
// ensures the checks table and available view exist.
func OpenSQLite(path string) (*SQLite, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("sink: open %s: %w", path, err)
	}
	// A single connection avoids SQLite write contention and keeps :memory:
	// databases consistent across queries.
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("sink: init %s: %w", path, err)
	}
	return &SQLite{db: db, now: time.Now}, nil
}

// Write upserts every result, stamping the batch with the current UTC time.
func (s *SQLite) Write(ctx context.Context, results []engine.Result) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("sink: begin: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, upsert)
	if err != nil {
		return fmt.Errorf("sink: prepare: %w", err)
	}
	defer stmt.Close()

	at := s.now().UTC().Format(time.RFC3339)
	for _, r := range results {
		if _, err := stmt.ExecContext(ctx, r.Name, r.ID, r.Platform, string(r.Status), r.Detail, r.URL, at); err != nil {
			return fmt.Errorf("sink: write %s/%s: %w", r.Name, r.ID, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("sink: commit: %w", err)
	}
	return nil
}

// Close releases the database handle.
func (s *SQLite) Close() error {
	return s.db.Close()
}
