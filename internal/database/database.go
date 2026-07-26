package database

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// Database is the main structure for interacting with SQLite database.
type Database struct {
	connection *sqlx.DB
}

func Connect(name string) (*Database, error) {
	connection, err := sqlx.Connect("sqlite", name)
	if err != nil {
		return nil, err
	}

	db := &Database{
		connection: connection,
	}

	// Perform migrations.
	if err := Migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

// Exec executes a SQL query.
func (db *Database) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.connection.Exec(query, args...)
}

// BeginTransaction begins a SQL transaction.
func (db *Database) BeginTransaction() (*sqlx.Tx, error) {
	return db.connection.Beginx()
}
