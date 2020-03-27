package sql

import (
	"context"
	gosql "database/sql"
	"net/http"

	"github.com/gildas/go-errors"
	"github.com/gildas/go-logger"
)

type DB struct {
	db     *gosql.DB
	Logger *logger.Logger
}

type key int
const dbContextKey key = iota * 31415

// Open opens a database specified by its database driver name and a driver-specific data source name,
// usually consisting of at least a database name and connection information.
//
// Most users will open a database via a driver-specific connection helper function that returns a *DB.
// No database drivers are included in the Go standard library.
// See https://golang.org/s/sqldrivers for a list of third-party drivers.
//
// Open may just validate its arguments without creating a connection to the database.
// To verify that the data source name is valid, call Ping.
//
// The returned DB is safe for concurrent use by multiple goroutines and maintains its own pool of idle connections.
// Thus, the Open function should be called just once. It is rarely necessary to close a DB.
func Open(drivername string, datasourceName string, l *logger.Logger) (db *DB, err error) {
	db = &DB{
		Logger: logger.CreateIfNil(l, "sql").Child("db", "db"),
	}

	db.db, err = gosql.Open(drivername, datasourceName)
	return db, errors.RuntimeError.Wrap(err)
}

// Ping verifies a connection to the database is still alive, establishing a connection if necessary
func (db DB) Ping() error {
	return db.db.Ping()
}

// Exec executes a query without returning any rows. The args are for any placeholder parameters in the query
func (db *DB) Exec(query string, args ...interface{}) (gosql.Result, error) {
	return db.db.Exec(query, args...)
}

// Exec executes a query without returning any rows. The args are for any placeholder parameters in the query
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (gosql.Result, error) {
	return db.db.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows, typically a SELECT. The args are for any placeholder parameters in the query
func (db *DB) Query(query string, args ...interface{}) (*gosql.Rows, error) {
	return db.db.Query(query, args...)
}

// Query executes a query that returns rows, typically a SELECT. The args are for any placeholder parameters in the query
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*gosql.Rows, error) {
	return db.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that is expected to return at most one row.
// QueryRow always returns a non-nil value. Errors are deferred until Row's Scan method is called.
// If the query selects no rows, the *Row's Scan will return ErrNoRows.
// Otherwise, the *Row's Scan scans the first selected row and discards the rest
func (db *DB) QueryRow(query string, args ...interface{}) *gosql.Row {
	return db.db.QueryRow(query, args...)
}

// QueryRowContext executes a query that is expected to return at most one row.
// QueryRowContext always returns a non-nil value. Errors are deferred until Row's Scan method is called.
// If the query selects no rows, the *Row's Scan will return ErrNoRows.
// Otherwise, the *Row's Scan scans the first selected row and discards the rest
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *gosql.Row {
	return db.db.QueryRowContext(ctx, query, args...)
}

// Close closes the database and prevents new queries from starting.
// Close then waits for all queries that have started processing on the server to finish.
//
// It is rare to Close a DB, as the DB handle is meant to be long-lived and shared between many goroutines.
func (db *DB) Close() error {
	db.Logger.Infof("Closing Database Connection")
	return db.db.Close()
}

// FromContext retrieves a DB stored in the given context
func FromContext(context context.Context) (*DB, error) {
	if db, ok := context.Value(dbContextKey).(*DB); ok {
		return db, nil
	}
	return nil, errors.ArgumentMissing.With("DB").WithStack()
}

// ToContext stores db to the given context
func (db *DB) ToContext(parent context.Context) context.Context {
	return context.WithValue(parent, dbContextKey, db)
}

// HttpHandler wraps a DB in an http middleware Handler
func (db *DB) HttpHandler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(db.ToContext(r.Context())))
		})
	}
}

// Must returns the given DB or panics upon error
func Must(db *DB, err error) *DB {
	if err != nil {
		panic(err)
	}
	return db
}