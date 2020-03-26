package sql

import "math"

// QueryOperator describes the operator used to build
type QueryOperator struct {
	Operator string
	Arity    int
}

var (
	QueryBetween        = QueryOperator{"BETWEEN", 3}
	QueryDifferent      = QueryOperator{"<>", 2}
	QueryEqual          = QueryOperator{"=", 2}
	QueryGreater        = QueryOperator{">", 2}
	QueryGreaterOrEqual = QueryOperator{">=", 2}
	QueryIn             = QueryOperator{"IN", math.MaxInt32}
	QueryLesser         = QueryOperator{"<", 2}
	QueryLesserOrEqual  = QueryOperator{"<=", 2}
	QueryLike           = QueryOperator{"LIKE", 2}
)

// String returns a string representation of the operator
func (operator QueryOperator) String() string {
	return operator.Operator
}