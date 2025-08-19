package database

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

type Database struct {
    DB *sql.DB
}

func NewDatabase(connectionString string) (*Database, error) {
    // Open database connection
    db, err := sql.Open("postgres", connectionString)
    if err != nil {
        return nil, err
    }

    // Configure connection pool
    db.SetMaxOpenConns(25) // Limit maximum simultaneous connections
    db.SetMaxIdleConns(5)  // Keep some connections ready
    db.SetConnMaxLifetime(5 * time.Minute) // Refresh connections periodically

    // Verify connection is working
    if err := db.Ping(); err != nil {
        return nil, err
    }

    return &Database{DB: db}, nil
}

func (d *Database) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.DB.Query(query, args...)
}

func (d *Database) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.DB.QueryRow(query, args...)
}

func (d *Database) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.DB.Exec(query, args...)
}
