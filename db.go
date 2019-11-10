package main

import (
	"context"
	"database/sql"
	"fmt"

	_ "rsc.io/sqlite"
)

// DbConfig contains necessary for database configuration information
type DbConfig struct {
	Conn   string `yaml:"connectionString"`
	Driver string `yaml:"driver"`
}

type executor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// due simplicity of test task i just make all migrations here
func migrate(db *sql.DB) error {

	// check if migration needed
	_, err := db.Query("SELECT id FROM users LIMIT 1")
	if err == nil {
		// to do looks weird a bit, but i will fix that if i have enough time
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS users (id VARCHAR(36) NOT NULL, balance REAL NOT NULL DEFAULT 0, PRIMARY KEY (id))")
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS operations (id VARCHAR(36) NOT NULL, amount REAL NOT NULL DEFAULT 0, prev_balance REAL NOT NULL DEFAULT 0, game_state INT NOT NULL, created_at BIGINT NOT NULL, deleted_at BIGINT, PRIMARY KEY (id))")
	if err != nil {
		tx.Rollback()
		return err
	}

	insertUserQuery := "INSERT INTO users (id) VALUES (?)"
	_, err = tx.Exec(insertUserQuery, defaultUserUUID)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// NewDB build sql.DB instance from db config file
func NewDB(cfg DbConfig) (*sql.DB, error) {
	Log("open db %v", cfg)
	db, err := sql.Open(cfg.Driver, cfg.Conn)
	if err != nil {
		return nil, fmt.Errorf("Failed to open database: %w", err)
	}
	// TODO: move to config
	db.SetMaxOpenConns(1)
	return db, nil
}
