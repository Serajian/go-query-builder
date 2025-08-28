package qb

// GroupBy -> GROUP BY and HAVING
func (qb *QueryBuilder) GroupBy(columns ...string) *QueryBuilder {
	qb.groupBy = append(qb.groupBy, columns...)
	return qb
}

func (qb *QueryBuilder) Having(column string, op Operator, value interface{}) *QueryBuilder {
	condition := Condition{
		Column: column,
		Op:     op,
		Value:  value,
		Logic:  "AND",
	}
	qb.having = append(qb.having, condition)
	return qb
}
