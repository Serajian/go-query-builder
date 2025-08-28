package qb

import (
	"fmt"
	"strings"
)

// NewQB creates a new QueryBuilder instance
func NewQB() *QueryBuilder {
	return &QueryBuilder{
		columns:    []string{},
		conditions: []Condition{},
		joins:      []Join{},
		groupBy:    []string{},
		having:     []Condition{},
		orderBy:    []OrderBy{},
		parameters: []interface{}{},
	}
}

// Select is SELECT operations
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.queryType = SELECT
	if len(columns) == 0 {
		qb.columns = []string{"*"}
	} else {
		qb.columns = columns
	}
	return qb
}

func (qb *QueryBuilder) From(table string) *QueryBuilder {
	qb.table = table
	return qb
}

// Insert is INSERT operations
func (qb *QueryBuilder) Insert(table string) *QueryBuilder {
	qb.queryType = INSERT
	qb.table = table
	qb.insertData = make(map[string]interface{})
	return qb
}

func (qb *QueryBuilder) Values(data map[string]interface{}) *QueryBuilder {
	qb.insertData = data
	return qb
}

func (qb *QueryBuilder) Set(column string, value interface{}) *QueryBuilder {
	if qb.insertData == nil {
		qb.insertData = make(map[string]interface{})
	}
	qb.insertData[column] = value
	return qb
}

// Update is UPDATE operations
func (qb *QueryBuilder) Update(table string) *QueryBuilder {
	qb.queryType = UPDATE
	qb.table = table
	qb.updateData = make(map[string]interface{})
	return qb
}

func (qb *QueryBuilder) SetUpdate(column string, value interface{}) *QueryBuilder {
	if qb.updateData == nil {
		qb.updateData = make(map[string]interface{})
	}
	qb.updateData[column] = value
	return qb
}

func (qb *QueryBuilder) Delete(table string) *QueryBuilder {
	qb.queryType = DELETE
	qb.table = table
	return qb
}

func (qb *QueryBuilder) Build() (string, []interface{}) {
	qb.parameters = []interface{}{}

	switch qb.queryType {
	case SELECT:
		return qb.buildSelect()
	case INSERT:
		return qb.buildInsert()
	case UPDATE:
		return qb.buildUpdate()
	case DELETE:
		return qb.buildDelete()
	default:
		return "", nil
	}
}

func (qb *QueryBuilder) Paginate(page, perPage int) *QueryBuilder {
	return qb.Limit(perPage).Offset((page - 1) * perPage)
}

func (qb *QueryBuilder) Reset() *QueryBuilder {
	return NewQB()
}

func (qb *QueryBuilder) buildSelect() (string, []interface{}) {
	var query strings.Builder

	// SELECT clause
	query.WriteString("SELECT ")
	query.WriteString(strings.Join(qb.columns, ", "))

	// FROM clause
	if qb.table != "" {
		query.WriteString(" FROM ")
		query.WriteString(qb.table)
	}

	// JOIN clause
	for _, join := range qb.joins {
		query.WriteString(" ")
		query.WriteString(string(join.Type))
		query.WriteString(" ")
		query.WriteString(join.Table)
		query.WriteString(" ON ")
		query.WriteString(join.Condition)
	}

	// WHERE clause
	if len(qb.conditions) > 0 {
		query.WriteString(" WHERE ")
		qb.buildConditions(&query, qb.conditions)
	}

	// GROUP BY clause
	if len(qb.groupBy) > 0 {
		query.WriteString(" GROUP BY ")
		query.WriteString(strings.Join(qb.groupBy, ", "))
	}

	// HAVING clause
	if len(qb.having) > 0 {
		query.WriteString(" HAVING ")
		qb.buildConditions(&query, qb.having)
	}

	// ORDER BY clause
	if len(qb.orderBy) > 0 {
		query.WriteString(" ORDER BY ")
		orderParts := make([]string, len(qb.orderBy))
		for i, order := range qb.orderBy {
			if order.Desc {
				orderParts[i] = order.Column + "DESC"
			} else {
				orderParts[i] = order.Column + "ASC"
			}
		}
		query.WriteString(strings.Join(orderParts, ", "))
	}

	// LIMIT clause
	if qb.limit > 0 {
		query.WriteString(fmt.Sprintf(" LIMIT %d", qb.limit))
	}

	// OFFSET clause
	if qb.offset > 0 {
		query.WriteString(fmt.Sprintf(" OFFSET %d", qb.offset))
	}

	return query.String(), qb.parameters
}

func (qb *QueryBuilder) buildInsert() (string, []interface{}) {
	var query strings.Builder

	query.WriteString("INSERT INTO ")
	query.WriteString(qb.table)

	if len(qb.insertData) > 0 {
		columns := make([]string, 0, len(qb.insertData))
		placeholders := make([]string, 0, len(qb.insertData))

		for column, value := range qb.insertData {
			columns = append(columns, column)
			placeholders = append(placeholders, "?")
			qb.parameters = append(qb.parameters, value)
		}

		query.WriteString(" (")
		query.WriteString(strings.Join(columns, ", "))
		query.WriteString(") VALUES (")
		query.WriteString(strings.Join(placeholders, ", "))
		query.WriteString(")")
	}

	return query.String(), qb.parameters
}

func (qb *QueryBuilder) buildUpdate() (string, []interface{}) {
	var query strings.Builder

	query.WriteString("UPDATE ")
	query.WriteString(qb.table)
	query.WriteString(" SET ")

	setParts := make([]string, 0, len(qb.updateData))
	for column, value := range qb.updateData {
		setParts = append(setParts, column+" = ?")
		qb.parameters = append(qb.parameters, value)
	}
	query.WriteString(strings.Join(setParts, ", "))

	// WHERE clause
	if len(qb.conditions) > 0 {
		query.WriteString(" WHERE ")
		qb.buildConditions(&query, qb.conditions)
	}

	return query.String(), qb.parameters
}

func (qb *QueryBuilder) buildDelete() (string, []interface{}) {
	var query strings.Builder

	query.WriteString("DELETE FROM ")
	query.WriteString(qb.table)

	// WHERE clause
	if len(qb.conditions) > 0 {
		query.WriteString(" WHERE ")
		qb.buildConditions(&query, qb.conditions)
	}

	return query.String(), qb.parameters
}

func (qb *QueryBuilder) buildConditions(query *strings.Builder, conditions []Condition) {
	for i, condition := range conditions {
		if i > 0 {
			query.WriteString(" ")
			query.WriteString(condition.Logic)
			query.WriteString(" ")
		}

		query.WriteString(condition.Column)
		query.WriteString(" ")
		query.WriteString(string(condition.Op))

		switch condition.Op {
		case NULL, NOTNULL:
			// No value needed
		case IN, NIN:
			if values, ok := condition.Value.([]interface{}); ok {
				placeholders := make([]string, len(values))
				for j, value := range values {
					placeholders[j] = "?"
					qb.parameters = append(qb.parameters, value)
				}
				query.WriteString(" (")
				query.WriteString(strings.Join(placeholders, ", "))
				query.WriteString(")")
			}
		default:
			query.WriteString(" ?")
			qb.parameters = append(qb.parameters, condition.Value)
		}
	}
}
