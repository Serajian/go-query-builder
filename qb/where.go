package qb

// Where -> WHERE conditions by AND logic
func (qb *QueryBuilder) Where(column string, op Operator, value interface{}) *QueryBuilder {
	condition := Condition{
		Column: column,
		Op:     op,
		Value:  value,
		Logic:  "AND",
	}
	qb.condition = append(qb.condition, condition)
	return qb
}

// OrWhere -> WHERE conditions by OR logic
func (qb *QueryBuilder) OrWhere(column string, op Operator, value interface{}) *QueryBuilder {
	condition := Condition{
		Column: column,
		Op:     op,
		Value:  value,
		Logic:  "OR",
	}
	qb.condition = append(qb.condition, condition)
	return qb
}

func (qb *QueryBuilder) WhereIn(column string, value interface{}) *QueryBuilder {
	return qb.Where(column, IN, value)
}

func (qb *QueryBuilder) WhereNotIn(column string, value interface{}) *QueryBuilder {
	return qb.Where(column, NIN, value)
}

func (qb *QueryBuilder) WhereLike(column string, pattern string) *QueryBuilder {
	return qb.Where(column, LIKE, pattern)
}

func (qb *QueryBuilder) WhereNotLike(column string, pattern string) *QueryBuilder {
	return qb.Where(column, NOTLIKE, pattern)
}

func (qb *QueryBuilder) WhereNull(column string) *QueryBuilder {
	return qb.Where(column, NULL, nil)
}

func (qb *QueryBuilder) WhereNotNull(column string) *QueryBuilder {
	return qb.Where(column, NOTNULL, nil)
}
