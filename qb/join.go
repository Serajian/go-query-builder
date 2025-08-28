package qb

// Join -> JOIN operations
func (qb *QueryBuilder) Join(table string, condition string) *QueryBuilder {
	join := Join{
		Type:      INNER,
		Table:     table,
		Condition: condition,
	}
	qb.joins = append(qb.joins, join)
	return qb
}

func (qb *QueryBuilder) LeftJoin(table string, condition string) *QueryBuilder {
	join := Join{
		Type:      LEFT,
		Table:     table,
		Condition: condition,
	}
	qb.joins = append(qb.joins, join)
	return qb
}

func (qb *QueryBuilder) RightJoin(table string, condition string) *QueryBuilder {
	join := Join{
		Type:      RIGHT,
		Table:     table,
		Condition: condition,
	}
	qb.joins = append(qb.joins, join)
	return qb
}
