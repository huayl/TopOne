package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"sandswind/marble/log"
	"strings"
)

const (
	dbName = "db.sqlite"
)

// The SQL database.
type DB struct {
	dbConn *sql.DB
}

// Query result types
type RowResult map[string]string
type RowResults []map[string]string

// New creates a new database. Deletes any existing database.
func New(dbPath string) *DB {
	os.Remove(dbPath)
	return Open(dbPath)
}

// Open an existing database, creating it if it does not exist.
func Open(dbPath string) *DB {
	log.Info("SQLite database path is %s", dbPath)
	dbc, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Error("sqlite open %v", err)
		return nil
	}
	return &DB{
		dbConn: dbc,
	}
}

// Close closes the underlying database connection.
func (db *DB) Close() error {
	return db.dbConn.Close()
}

// Query runs the supplied query against the sqlite database. It returns a slice of
// RowResults.
func (db *DB) Query(query string) (RowResults, error) {
	if !strings.HasPrefix(strings.ToUpper(query), "SELECT ") {
		log.Warn("Query \"%s\" may modify the database", query)
	}
	rows, err := db.dbConn.Query(query)
	if err != nil {
		log.Error("failed to execute SQLite query", err.Error())
		return nil, err
	}
	defer rows.Close()

	results := make(RowResults, 0)

	columns, _ := rows.Columns()
	rawResult := make([][]byte, len(columns))
	dest := make([]interface{}, len(columns))
	for i, _ := range rawResult {
		dest[i] = &rawResult[i] // Pointers to each string in the interface slice
	}

	for rows.Next() {
		err = rows.Scan(dest...)
		if err != nil {
			log.Error("failed to scan SQLite row", err.Error())
			return nil, err
		}

		r := make(RowResult)
		for i, raw := range rawResult {
			if raw == nil {
				r[columns[i]] = "null"
			} else {
				r[columns[i]] = string(raw)
			}
		}
		results = append(results, r)
	}
	log.Debug("Executed query successfully:%v", query)
	return results, nil
}

// Execute executes the given sqlite statement, of a type that doesn't return rows.
func (db *DB) Execute(stmt string) error {
	_, err := db.dbConn.Exec(stmt)
	if err != nil {
		log.Error("Error executing \"%s\", error: %v", stmt, err)
	}
	return err
}

// StartTransaction starts an explicit transaction.
func (db *DB) Begin() error {
	_, err := db.dbConn.Exec("BEGIN")
	if err != nil {
		log.Error("Error starting transaction")
	}
	return err
}

// CommitTransaction commits all changes made since StartTraction was called.
func (db *DB) Commit() error {
	_, err := db.dbConn.Exec("END")
	if err != nil {
		log.Error("Error ending transaction")
	}
	return err
}

// RollbackTransaction aborts the transaction. No statement issued since
// StartTransaction was called will take effect.
func (db *DB) Rollback() error {
	_, err := db.dbConn.Exec("ROLLBACK")
	if err != nil {
		log.Error("Error rolling back transaction")
	}
	return err
}
