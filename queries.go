package sql

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Queries describes a map of Query objects
type Queries map[string]Query

// Query describes a query in a Statement Where Clause
type Query []interface{}

// QueriesFromRequest creates a Queries from an HTTP Request
func QueriesFromRequest(r *http.Request) Queries {
	return QueriesFromURL(r.URL)
}

// QueriesFromURL creates Queries from a URL (from its query part)
func QueriesFromURL(u *url.URL) Queries {
	queries := Queries{}
	for key, values := range u.Query() {
		qvalues := make([]interface{}, len(values))
		for i, value := range values {
			qvalues[i] = value
		}
		queries.Add(key, qvalues...)
	}
	return queries
}

// Add adds a new Query
//
// If the key is already present, values are added
// If no values are given, the Queries is unchanged
func (queries Queries) Add(key string, values ...interface{}) Queries {
	if len(values) == 0 {
		return queries
	}
	if set, ok := values[0].(QueryOperator); ok && set.Operator == QuerySet.Operator {
		key = "=" + key
	}
	if current, found := queries[key]; found {
		if len(current) == 2 {
			current[0] = QueryIn
		}
	} else {
		if _, ok := values[0].(QueryOperator); !ok {
			switch len(values) {
			case 1:
				queries[key] = []interface{}{QueryEqual}
			default:
				queries[key] = []interface{}{QueryIn}
			}
		}
	}
	queries[key] = append(queries[key], values...)
	return queries
}

// WhereClause builds the SQL Where Clause for a Statement
func (queries Queries) WhereClause() (string, []interface{}) {
	clause := strings.Builder{}
	parms  := []interface{}{}
	for column, values := range queries {
		operator, _ := values[0].(QueryOperator)
		if operator.Operator == QueryIn.Operator {
			args := []string{}
			for _, value := range values[1:] {
				parms = append(parms, value)
				args  = append(args, fmt.Sprintf("$%d", len(parms)))
			}
			clause.WriteString(fmt.Sprintf(" AND %s %s (%s)", column, operator, strings.Join(args, ", ")))
		} else {
			if len(values) != operator.Arity || operator.Operator == QuerySet.Operator {
				// ignore wrong # of arguments or SET Operator (used by UpdateStatement)
				continue
			}
			parms = append(parms, values[1])
			clause.WriteString(fmt.Sprintf(" AND %s %s $%d", column, operator, len(parms)))
		}
	}
	return strings.TrimPrefix(clause.String(), " AND "), parms
}