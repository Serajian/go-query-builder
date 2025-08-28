package qb

import (
	"sort"
	"strings"
)

// Insert starts an INSERT statement for the given table and initializes InsertData.
// Use Set/ Values to add column values. Supports RETURNING on dialects that allow it.
func (qb *QueryBuilder) Insert(table string) *QueryBuilder {
	qb.QueryType = INSERT
	qb.Table = table
	qb.InsertData = make(map[string]interface{})
	return qb
}

// Values replaces the current InsertData map with the provided one.
// Keys are sorted at render time to make placeholder order deterministic.
func (qb *QueryBuilder) Values(data map[string]interface{}) *QueryBuilder {
	qb.InsertData = data
	return qb
}

// Set adds or replaces a single column/value pair for INSERT.
func (qb *QueryBuilder) Set(column string, value interface{}) *QueryBuilder {
	if qb.InsertData == nil {
		qb.InsertData = make(map[string]interface{})
	}
	qb.InsertData[column] = value
	return qb
}

func (qb *QueryBuilder) buildInsert() (string, []interface{}) {
	var query strings.Builder

	query.WriteString("INSERT INTO ")
	query.WriteString(qb.Table)

	if len(qb.InsertData) == 0 {
		if qb.PhStyle == DollarN {
			// Postgres / (SQLite 3.35+)
			query.WriteString(" DEFAULT VALUES")
			// ON CONFLICT (just PG/SQLite)
			qb.renderOnConflict(&query)
			// RETURNING (just PG/SQLite)
			if len(qb.ReturningColumns) > 0 {
				query.WriteString(" RETURNING ")
				query.WriteString(strings.Join(qb.ReturningColumns, ", "))
			}
		} else {
			// MySQL
			query.WriteString(" () VALUES ()")
		}
		return query.String(), qb.Parameters
	}

	columns := make([]string, 0, len(qb.InsertData))
	for col := range qb.InsertData {
		columns = append(columns, col)
	}
	sort.Strings(columns)

	placeholders := make([]string, 0, len(columns))
	for _, column := range columns {
		placeholders = append(placeholders, qb.placeholder())
		qb.Parameters = append(qb.Parameters, qb.InsertData[column])
	}

	query.WriteString(" (")
	query.WriteString(strings.Join(columns, ", "))
	query.WriteString(") VALUES (")
	query.WriteString(strings.Join(placeholders, ", "))
	query.WriteString(")")

	// ON CONFLICT (just in case: DollarN ⇒ PG/SQLite)
	qb.renderOnConflict(&query)

	// RETURNING (just PG/SQLite)
	if qb.PhStyle == DollarN && len(qb.ReturningColumns) > 0 {
		query.WriteString(" RETURNING ")
		query.WriteString(strings.Join(qb.ReturningColumns, ", "))
	}

	return query.String(), qb.Parameters
}

func (qb *QueryBuilder) renderOnConflict(query *strings.Builder) {
	if qb.PhStyle != DollarN {
		return
	}
	if len(qb.ConflictColumns) == 0 && qb.ConflictConstraint == "" &&
		!qb.ConflictDoNothing && len(qb.ConflictUpdateSet) == 0 {
		return
	}

	query.WriteString(" ON CONFLICT ")
	if qb.ConflictConstraint != "" {
		query.WriteString("ON CONSTRAINT ")
		query.WriteString(qb.ConflictConstraint)
	} else if len(qb.ConflictColumns) > 0 {
		query.WriteString("(")
		query.WriteString(strings.Join(qb.ConflictColumns, ", "))
		query.WriteString(")")
	}

	if qb.ConflictDoNothing {
		query.WriteString(" DO NOTHING")
		return
	}

	if len(qb.ConflictUpdateSet) > 0 {
		query.WriteString(" DO UPDATE SET ")

		keys := make([]string, 0, len(qb.ConflictUpdateSet))
		for k := range qb.ConflictUpdateSet {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		parts := make([]string, 0, len(keys))
		for _, col := range keys {
			val := qb.ConflictUpdateSet[col]
			if raw, ok := val.(RawExpr); ok {
				parts = append(parts, col+" = "+string(raw))
			} else {
				ph := qb.placeholder()
				qb.Parameters = append(qb.Parameters, val)
				parts = append(parts, col+" = "+ph)
			}
		}
		query.WriteString(strings.Join(parts, ", "))
	}
}
