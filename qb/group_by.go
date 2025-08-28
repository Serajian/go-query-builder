package qb

// GroupBy appends columns to GROUP BY.
func (qb *QueryBuilder) GroupBy(columns ...string) *QueryBuilder {
	qb.GroupByColumns = append(qb.GroupByColumns, columns...)
	return qb
}

// Having adds a HAVING predicate (combined with AND by default).
func (qb *QueryBuilder) Having(column string, op Operator, value interface{}) *QueryBuilder {
	condition := Condition{
		Column: column,
		Op:     op,
		Value:  value,
		Logic:  "AND",
	}
	qb.HavingConditions = append(qb.HavingConditions, condition)
	return qb
}
