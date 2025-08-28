package qb

import (
	"sort"
	"strings"
)

// Update starts an UPDATE statement for the given table and initializes UpdateData.
// Use SetUpdate to add assignments. Supports RETURNING on dialects that allow it.
func (qb *QueryBuilder) Update(table string) *QueryBuilder {
	qb.QueryType = UPDATE
	qb.Table = table
	qb.UpdateData = make(map[string]interface{})
	return qb
}

// SetUpdate adds or replaces a single column assignment for UPDATE SET.
func (qb *QueryBuilder) SetUpdate(column string, value interface{}) *QueryBuilder {
	if qb.UpdateData == nil {
		qb.UpdateData = make(map[string]interface{})
	}
	qb.UpdateData[column] = value
	return qb
}

func (qb *QueryBuilder) buildUpdate() (string, []interface{}) {
	var query strings.Builder

	query.WriteString("UPDATE ")
	query.WriteString(qb.Table)
	query.WriteString(" SET ")

	// Stable order for update set clauses
	keys := make([]string, 0, len(qb.UpdateData))
	for k := range qb.UpdateData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	setParts := make([]string, 0, len(keys))
	for _, column := range keys {
		setParts = append(setParts, column+" = "+qb.placeholder())
		qb.Parameters = append(qb.Parameters, qb.UpdateData[column])
	}
	query.WriteString(strings.Join(setParts, ", "))

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
