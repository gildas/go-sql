package sql

// Statement describes stuff that can be Built into a statement string with its parameters
type Statement interface {
	With(db *DB) Statement
	Build(table string, columns []string, queries Queries) (string, []interface{})
}