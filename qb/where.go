package qb

// Where adds a WHERE predicate combined with AND.
func (qb *QueryBuilder) Where(column string, op Operator, value interface{}) *QueryBuilder {
	condition := Condition{
		Column: column,
		Op:     op,
		Value:  value,
		Logic:  "AND",
	}
	qb.Conditions = append(qb.Conditions, condition)
	return qb
}

// OrWhere adds a WHERE predicate combined with OR.
func (qb *QueryBuilder) OrWhere(column string, op Operator, value interface{}) *QueryBuilder {
	condition := Condition{
		Column: column,
		Op:     op,
		Value:  value,
		Logic:  "OR",
	}
	qb.Conditions = append(qb.Conditions, condition)
	return qb
}

// WhereIn adds an IN (...) predicate; accepts any slice/array as value.
func (qb *QueryBuilder) WhereIn(column string, value interface{}) *QueryBuilder {
	return qb.Where(column, IN, value)
}

// WhereNotIn adds a NOT IN (...) predicate; accepts any slice/array as value.
func (qb *QueryBuilder) WhereNotIn(column string, value interface{}) *QueryBuilder {
	return qb.Where(column, NIN, value)
}

// WhereLike adds a LIKE predicate (value should include wildcards, e.g. %foo%).
func (qb *QueryBuilder) WhereLike(column, pattern string) *QueryBuilder {
	return qb.Where(column, LIKE, pattern)
}

// WhereNotLike adds a NOT LIKE predicate.
func (qb *QueryBuilder) WhereNotLike(column, pattern string) *QueryBuilder {
	return qb.Where(column, NOTLIKE, pattern)
}

// WhereNull adds an IS NULL predicate.
func (qb *QueryBuilder) WhereNull(column string) *QueryBuilder {
	return qb.Where(column, NULL, nil)
}

// WhereNotNull adds an IS NOT NULL predicate.
func (qb *QueryBuilder) WhereNotNull(column string) *QueryBuilder {
	return qb.Where(column, NOTNULL, nil)
}
