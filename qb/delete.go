package qb

import "strings"

// Delete starts a DELETE statement for the given table. Any prior per-query state
// is cleared by Reset during Build; conditions can be added via Where/ OrWhere.
func (qb *QueryBuilder) Delete(table string) *QueryBuilder {
	qb.Reset()

	qb.QueryType = DELETE
	qb.Table = table
	return qb
}

func (qb *QueryBuilder) buildDelete() (string, []interface{}) {
	var query strings.Builder

	query.WriteString("DELETE FROM ")
	query.WriteString(qb.Table)

	// WHERE clause
	if len(qb.Conditions) > 0 {
		query.WriteString(" WHERE ")
		qb.buildConditions(&query, qb.Conditions)
	} else if qb.GuardWrites {
		query.WriteString(" WHERE 1=0 /*guarded: mising WHERE */")
	}

	// RETURNING
	if len(qb.ReturningColumns) > 0 {
		query.WriteString(" RETURNING ")
		query.WriteString(strings.Join(qb.ReturningColumns, ", "))
	}
	return query.String(), qb.Parameters
}
