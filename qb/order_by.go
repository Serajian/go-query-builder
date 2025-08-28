package qb

// OrderBy appends an ascending ORDER BY on the given column.
func (qb *QueryBuilder) OrderBy(column string) *QueryBuilder {
	order := OrderBy{
		Column: column,
		Desc:   false,
	}
	qb.OrderByArr = append(qb.OrderByArr, order)
	return qb
}

// OrderByDesc appends a descending ORDER BY on the given column.
func (qb *QueryBuilder) OrderByDesc(column string) *QueryBuilder {
	order := OrderBy{
		Column: column,
		Desc:   true,
	}
	qb.OrderByArr = append(qb.OrderByArr, order)
	return qb
}

// Limit sets the LIMIT value (rendered inline, not as a parameter).
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.LimitInt = limit
	return qb
}

// Offset sets the OFFSET value (rendered inline, not as a parameter).
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.OffsetInt = offset
	return qb
}
