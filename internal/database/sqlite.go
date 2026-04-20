package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// DB represents a wrapper around the sql.DB for common operations.
type DB struct {
	sqlDB *sql.DB
}

// New initializes the SQLite connection and runs necessary migrations.
func New(dsn string) (*DB, error) {
	// Enable WAL (Write-Ahead Logging) and foreign keys for better performance and safety.
	connectionString := fmt.Sprintf("file:%s?cache=shared&mode=rwc&_journal_mode=WAL&_fk=1", dsn)
	
	db, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Connection pooling configuration best suited for SQLite
	db.SetMaxOpenConns(1) // Keep to 1 for SQLite writers to prevent busy locks
	db.SetMaxIdleConns(1)

	portfolioDB := &DB{sqlDB: db}

	// Run schemas/migrations
	if err := portfolioDB.executeMigrations(); err != nil {
		return nil, fmt.Errorf("failed to execute migrations: %w", err)
	}

	return portfolioDB, nil
}

// Close gracefully closes the database connection.
func (db *DB) Close() error {
	if db.sqlDB != nil {
		return db.sqlDB.Close()
	}
	return nil
}

// Ping verifies the database connection.
func (db *DB) Ping() error {
	if db.sqlDB != nil {
		return db.sqlDB.Ping()
	}
	return nil
}

// ExecContext provides access to standard ExecContext, useful for building services.
func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.sqlDB.ExecContext(ctx, query, args...)
}

// QueryRowContext wrapper.
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return db.sqlDB.QueryRowContext(ctx, query, args...)
}

// QueryContext wrapper.
func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.sqlDB.QueryContext(ctx, query, args...)
}
