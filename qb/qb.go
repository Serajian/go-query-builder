package qb

import (
	"fmt"
	"reflect"
	"sort"
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
		phStyle:    DollarN, // Default
		paramIndex: 0,
	}
}

// WithPlaceholders sets placeholder style: QuestionMark (?) or DollarN ($1..)
func (qb *QueryBuilder) WithPlaceholders(style PlaceholderStyle) *QueryBuilder {
	qb.phStyle = style
	qb.paramIndex = 0
	return qb
}

// Select is SELECT operations
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.Reset()

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
	qb.Reset()

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
	qb.Reset()

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
	qb.Reset()

	qb.queryType = DELETE
	qb.table = table
	return qb
}

func (qb *QueryBuilder) Build() (string, []interface{}) {
	qb.parameters = []interface{}{}
	qb.paramIndex = 0 // reset placeholders

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
	style := qb.phStyle

	newQB := QueryBuilder{phStyle: style}
	*qb = newQB

	return qb
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
				orderParts[i] = order.Column + " DESC"
			} else {
				orderParts[i] = order.Column + " ASC"
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
		for col := range qb.insertData {
			columns = append(columns, col)
		}
		sort.Strings(columns)
		placeholders := make([]string, 0, len(columns))
		for _, column := range columns {
			placeholders = append(placeholders, qb.placeholder())
			qb.parameters = append(qb.parameters, qb.insertData[column])
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

	// Stable order for update set clauses
	keys := make([]string, 0, len(qb.updateData))
	for k := range qb.updateData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	setParts := make([]string, 0, len(keys))
	for _, column := range keys {
		setParts = append(setParts, column+" = "+qb.placeholder())
		qb.parameters = append(qb.parameters, qb.updateData[column])
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
			query.WriteString(condition.Logic) // AND / OR
			query.WriteString(" ")
		}

		switch condition.Op {
		case NULL, NOTNULL:
			// col IS NULL / col IS NOT NULL
			query.WriteString(condition.Column)
			query.WriteString(" ")
			query.WriteString(string(condition.Op))

		case IN, NIN:
			values, ok := sliceToInterfaces(condition.Value)
			if !ok || len(values) == 0 {
				if condition.Op == IN {
					query.WriteString("(1=0)") // always false
				} else {
					query.WriteString("(1=1)") // always true
				}
				continue
			}

			query.WriteString(condition.Column)
			query.WriteString(" ")
			query.WriteString(string(condition.Op))
			query.WriteString(" (")

			phs := make([]string, len(values))
			for j, v := range values {
				phs[j] = qb.placeholder()
				qb.parameters = append(qb.parameters, v)
			}
			query.WriteString(strings.Join(phs, ", "))
			query.WriteString(")")

		default:
			//   (=, !=, >, >=, <, <=, LIKE, NOT LIKE, ...)
			query.WriteString(condition.Column)
			query.WriteString(" ")
			query.WriteString(string(condition.Op))
			query.WriteString(" ")
			query.WriteString(qb.placeholder())
			qb.parameters = append(qb.parameters, condition.Value)
		}
	}
}

// placeholder returns the next placeholder according to the configured style.
func (qb *QueryBuilder) placeholder() string {
	switch qb.phStyle {
	case DollarN:
		qb.paramIndex++
		return fmt.Sprintf("$%d", qb.paramIndex)
	default:
		return "?"
	}
}

// sliceToInterfaces converts any slice/array (except []byte) to []interface{}.
// Returns (nil, false) if the input is not a slice/array.
func sliceToInterfaces(v interface{}) ([]interface{}, bool) {
	val := reflect.ValueOf(v)
	k := val.Kind()
	if k != reflect.Slice && k != reflect.Array {
		return nil, false
	}
	// treat []byte separately to avoid exploding into bytes
	if val.Type().Elem().Kind() == reflect.Uint8 {
		// If user really meant []byte inside IN, it's unusual — return single element
		return []interface{}{v}, true
	}
	out := make([]interface{}, val.Len())
	for i := 0; i < val.Len(); i++ {
		out[i] = val.Index(i).Interface()
	}
	return out, true
}
