package sql

import (
	"fmt"
	"strings"

	"github.com/gildas/go-logger"
)

type SelectStatement struct {
	DB     *DB
	Logger *logger.Logger
}

func (statement SelectStatement) With(db *DB) Statement {
	return &SelectStatement{
		DB:     db,
		Logger: logger.CreateIfNil(db.Logger, "sql").Child("statement", "statement"),
	}
}

// Build builds the statement to be executed by the DB
func (statement SelectStatement) Build(table string, columns []string, queries Queries) (string, []interface{}) {
	where, parms := queries.WhereClause()
	if len(where) > 0 {
		return fmt.Sprintf("SELECT %s FROM %s WHERE %s", strings.Join(columns, ", "), table, where), parms
	}
	return fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ", "), table), parms
}