package sql

import (
	"fmt"
	"strings"

	"github.com/gildas/go-logger"
)

type UpdateStatement struct {
	DB     *DB
	Logger *logger.Logger
}

func (statement UpdateStatement) With(db *DB) Statement {
	return &UpdateStatement{
		DB:     db,
		Logger: logger.CreateIfNil(db.Logger, "sql").Child("statement", "statement"),
	}
}

// Build builds the statement to be executed by the DB
func (statement UpdateStatement) Build(table string, columns []string, queries Queries) (string, []interface{}) {
	where, parms := queries.WhereClause()
	assignments := []string{}

	if len(where) == 0 {
		return "", []interface{}{}
	}
	for key, values := range queries {
		if operator, ok := values[0].(QueryOperator); ok && operator.Operator == QuerySet.Operator {
			parms = append(parms, values[1])
			assignments = append(assignments, fmt.Sprintf("%s = $%d", strings.TrimPrefix(key,"="), len(parms)))
		}
	}
	return fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, strings.Join(assignments, ", "), where), parms
}