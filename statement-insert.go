package sql

import (
	"fmt"
	"strings"

	"github.com/gildas/go-logger"
)

type InsertStatement struct {
	DB     *DB
	Logger *logger.Logger
}

func (statement InsertStatement) With(db *DB) Statement {
	return &InsertStatement{
		DB:     db,
		Logger: logger.CreateIfNil(db.Logger, "sql").Child("statement", "statement"),
	}
}

// Build builds the statement to be executed by the DB
func (statement InsertStatement) Build(table string, columns []string, queries Queries) (string, []interface{}) {
	cols   := []string{}
	values := []string{}
	parms  := []interface{}{}

	for key, query := range queries {
		cols   = append(cols, strings.TrimPrefix(key, "="))
		parms  = append(parms, query[1])
		values = append(values, fmt.Sprintf("$%d", len(parms)))
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(cols, ", "), strings.Join(values, ", ")), parms
}