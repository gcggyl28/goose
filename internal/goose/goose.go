// Package goose provides database migration functionality.
package goose

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// ErrNoMigrations is returned when no migration files are found.
var ErrNoMigrations = errors.New("no migrations found")

// Direction represents the direction of a migration.
type Direction int

const (
	// DirectionUp applies migrations.
	DirectionUp Direction = iota
	// DirectionDown rolls back migrations.
	DirectionDown
)

// Migration represents a single database migration.
type Migration struct {
	// Version is the migration version number parsed from the filename.
	Version int64
	// Filename is the original migration filename.
	Filename string
	// Source is the SQL content of the migration.
	Source string
}

// Goose manages database migrations.
type Goose struct {
	db        *sql.DB
	fsys      fs.FS
	tableName string
}

// New creates a new Goose instance.
func New(db *sql.DB, fsys fs.FS, opts ...Option) (*Goose, error) {
	if db == nil {
		return nil, errors.New("db must not be nil")
	}
	if fsys == nil {
		return nil, errors.New("fsys must not be nil")
	}
	g := &Goose{
		db:        db,
		fsys:      fsys,
		tableName: "goose_db_version",
	}
	for _, opt := range opts {
		opt(g)
	}
	return g, nil
}

// Option is a functional option for configuring Goose.
type Option func(*Goose)

// WithTableName sets the migration tracking table name.
func WithTableName(name string) Option {
	return func(g *Goose) {
		g.tableName = name
	}
}

// CollectMigrations reads all .sql migration files from the filesystem.
func (g *Goose) CollectMigrations() ([]*Migration, error) {
	var migrations []*Migration

	err := fs.WalkDir(g.fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".sql" {
			return nil
		}
		version, err := parseVersion(filepath.Base(path))
		if err != nil {
			return fmt.Errorf("invalid migration filename %q: %w", path, err)
		}
		content, err := fs.ReadFile(g.fsys, path)
		if err != nil {
			return fmt.Errorf("reading migration %q: %w", path, err)
		}
		migrations = append(migrations, &Migration{
			Version:  version,
			Filename: path,
			Source:   string(content),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(migrations) == 0 {
		return nil, ErrNoMigrations
	}
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})
	return migrations, nil
}

// EnsureTable creates the migration tracking table if it does not exist.
func (g *Goose) EnsureTable(ctx context.Context) error {
	q := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id         SERIAL PRIMARY KEY,
		version_id BIGINT NOT NULL,
		is_applied BOOLEAN NOT NULL,
		tstamp     TIMESTAMP DEFAULT NOW()
	)`, g.tableName)
	_, err := g.db.ExecContext(ctx, q)
	return err
}

// parseVersion extracts the numeric version prefix from a migration filename.
// e.g. "00001_create_users.sql" -> 1
func parseVersion(filename string) (int64, error) {
	parts := strings.SplitN(filename, "_", 2)
	if len(parts) < 2 {
		return 0, fmt.Errorf("expected format: <version>_<name>.sql")
	}
	return strconv.ParseInt(parts[0], 10, 64)
}
