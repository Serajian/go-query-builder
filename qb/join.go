package qb

// Join appends an INNER JOIN clause with the given ON condition.
func (qb *QueryBuilder) Join(table, condition string) *QueryBuilder {
	join := Join{
		Type:      INNER,
		Table:     table,
		Condition: condition,
	}
	qb.Joins = append(qb.Joins, join)
	return qb
}

// LeftJoin appends a LEFT JOIN clause with the given ON condition.
func (qb *QueryBuilder) LeftJoin(table, condition string) *QueryBuilder {
	join := Join{
		Type:      LEFT,
		Table:     table,
		Condition: condition,
	}
	qb.Joins = append(qb.Joins, join)
	return qb
}

// RightJoin appends a RIGHT JOIN clause with the given ON condition.
func (qb *QueryBuilder) RightJoin(table, condition string) *QueryBuilder {
	join := Join{
		Type:      RIGHT,
		Table:     table,
		Condition: condition,
	}
	qb.Joins = append(qb.Joins, join)
	return qb
}
