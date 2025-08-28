package qb

import (
	"fmt"
	"strings"
)

// Select starts a SELECT statement and sets the projected columns.
// When called with no columns, it defaults to SELECT *.
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.QueryType = SELECT
	if len(columns) == 0 {
		if qb.Columns == nil {
			qb.Columns = []string{"*"}
		}
	} else {
		qb.Columns = columns
	}
	return qb
}

// From sets the source table for SELECT/ DELETE and returns qb.
func (qb *QueryBuilder) From(table string) *QueryBuilder {
	qb.Table = table
	return qb
}

func (qb *QueryBuilder) buildSelect() (string, []interface{}) {
	var query strings.Builder

	// SELECT clause
	query.WriteString("SELECT ")
	query.WriteString(strings.Join(qb.Columns, ", "))

	// FROM clause
	if qb.Table != "" {
		query.WriteString(" FROM ")
		query.WriteString(qb.Table)
	}

	// JOIN clause
	for _, join := range qb.Joins {
		query.WriteString(" ")
		query.WriteString(string(join.Type))
		query.WriteString(" ")
		query.WriteString(join.Table)
		query.WriteString(" ON ")
		query.WriteString(join.Condition)
	}

	// WHERE clause
	if len(qb.Conditions) > 0 {
		query.WriteString(" WHERE ")
		qb.buildConditions(&query, qb.Conditions)
	}

	// GROUP BY clause
	if len(qb.GroupByColumns) > 0 {
		query.WriteString(" GROUP BY ")
		query.WriteString(strings.Join(qb.GroupByColumns, ", "))
	}

	// HAVING clause
	if len(qb.HavingConditions) > 0 {
		query.WriteString(" HAVING ")
		qb.buildConditions(&query, qb.HavingConditions)
	}

	// ORDER BY clause
	if len(qb.OrderByArr) > 0 {
		query.WriteString(" ORDER BY ")
		orderParts := make([]string, len(qb.OrderByArr))
		for i, order := range qb.OrderByArr {
			if order.Desc {
				orderParts[i] = order.Column + " DESC"
			} else {
				orderParts[i] = order.Column + " ASC"
			}
		}
		query.WriteString(strings.Join(orderParts, ", "))
	}

	// LIMIT clause
	if qb.LimitInt > 0 {
		query.WriteString(fmt.Sprintf(" LIMIT %d", qb.LimitInt))
	}

	// OFFSET clause
	if qb.OffsetInt > 0 {
		query.WriteString(fmt.Sprintf(" OFFSET %d", qb.OffsetInt))
	}

	return query.String(), qb.Parameters
}
