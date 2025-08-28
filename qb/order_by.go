package qb

// OrderBy -> ORDER BY

func (qb *QueryBuilder) OrderBy(column string) *QueryBuilder {
	order := OrderBy{
		Column: column,
		Desc:   false,
	}
	qb.orderBy = append(qb.orderBy, order)
	return qb
}

func (qb *QueryBuilder) OrderByDesc(column string) *QueryBuilder {
	order := OrderBy{
		Column: column,
		Desc:   true,
	}
	qb.orderBy = append(qb.orderBy, order)
	return qb
}

func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}
