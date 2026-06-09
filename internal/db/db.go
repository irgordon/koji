package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func Open(ctx context.Context, path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}
	conn.SetMaxOpenConns(1)

	if err := initializeConnection(ctx, conn); err != nil {
		_ = conn.Close()
		return nil, err
	}

	if err := RunMigrations(ctx, conn, InitialMigrations()); err != nil {
		_ = conn.Close()
		return nil, err
	}

	return conn, nil
}

func initializeConnection(ctx context.Context, conn *sql.DB) error {
	if _, err := conn.ExecContext(ctx, "PRAGMA busy_timeout = 5000"); err != nil {
		return fmt.Errorf("set sqlite busy timeout: %w", err)
	}
	if _, err := conn.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("enable sqlite foreign keys: %w", err)
	}
	if _, err := conn.ExecContext(ctx, "PRAGMA journal_mode = WAL"); err != nil {
		return fmt.Errorf("enable sqlite wal: %w", err)
	}
	return nil
}
