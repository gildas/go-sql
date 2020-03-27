package sql

import (
	"fmt"

	"github.com/gildas/go-logger"
)

type DeleteStatement struct {
	DB     *DB
	Logger *logger.Logger
}

func (statement DeleteStatement) With(db *DB) Statement {
	return &DeleteStatement{
		DB:     db,
		Logger: logger.CreateIfNil(db.Logger, "sql").Child("statement", "statement"),
	}
}

// Build builds the statement to be executed by the DB
func (statement DeleteStatement) Build(table string, columns []string, queries Queries) (string, []interface{}) {
	where, parms := queries.WhereClause()
	if len(where) > 0 {
		return fmt.Sprintf("DELETE FROM %s WHERE %s", table, where), parms
	}
	return fmt.Sprintf("DELETE FROM %s", table), parms
}