package qb

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// NewQB creates a new QueryBuilder with default DollarN placeholder style.
// All per-query state is zeroed; placeholder counter starts from 0.
func NewQB() *QueryBuilder {
	return &QueryBuilder{
		Conditions:       []Condition{},
		Joins:            []Join{},
		GroupByColumns:   []string{},
		HavingConditions: []Condition{},
		OrderByArr:       []OrderBy{},
		Parameters:       []interface{}{},
		PhStyle:          DollarN, // Default
		ParamIndex:       0,
	}
}

// WithPlaceholders sets the placeholder style (DollarN or QuestionMark)
// and resets the internal placeholder counter. It returns qb for chaining.
func (qb *QueryBuilder) WithPlaceholders(style PlaceholderStyle) *QueryBuilder {
	qb.PhStyle = style
	qb.ParamIndex = 0
	return qb
}

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

// Delete starts a DELETE statement for the given table. Any prior per-query state
// is cleared by Reset during Build; conditions can be added via Where/ OrWhere.
func (qb *QueryBuilder) Delete(table string) *QueryBuilder {
	qb.Reset()

	qb.QueryType = DELETE
	qb.Table = table
	return qb
}

// Returning adds a RETURNING clause for INSERT/ UPDATE/ DELETE.
// If called with no columns, it defaults to RETURNING *.
// Note: MySQL generally does not support RETURNING.
func (qb *QueryBuilder) Returning(columns ...string) *QueryBuilder {
	if len(columns) == 0 {
		qb.ReturningColumns = []string{"*"}
	} else {
		qb.ReturningColumns = columns
	}
	return qb
}

// Build renders the SQL string and the ordered parameter slice.
// It resets the placeholder counter, collects args, and (via defer) clears
// per-query state after rendering. Special cases:
//   - INSERT with no values: renders "DEFAULT VALUES" for DollarN (PG/SQLite),
//     or "() VALUES ()" for QuestionMark (MySQL).
//   - IN([]) renders "(1=0)" and NOT IN([]) renders "(1=1)".
func (qb *QueryBuilder) Build() (string, []interface{}) {
	qb.Parameters = []interface{}{}
	qb.ParamIndex = 0 // reset placeholders
	defer func() { qb.Reset() }()

	switch qb.QueryType {
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

// Paginate is a convenience for LIMIT/OFFSET with 1-based page numbering.
// Paginate(page, perPage) == LIMIT perPage OFFSET (page-1)*perPage.
func (qb *QueryBuilder) Paginate(page, perPage int) *QueryBuilder {
	return qb.Limit(perPage).Offset((page - 1) * perPage)
}

// Reset clears the builder's per-query state in place while preserving
// the placeholder style. It returns qb for chaining or reuse.
func (qb *QueryBuilder) Reset() *QueryBuilder {
	style := qb.PhStyle

	newQB := QueryBuilder{PhStyle: style}
	*qb = newQB

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

func (qb *QueryBuilder) buildInsert() (string, []interface{}) {
	var query strings.Builder

	query.WriteString("INSERT INTO ")
	query.WriteString(qb.Table)

	// حالت بدون ستون/مقدار
	if len(qb.InsertData) == 0 {
		if qb.PhStyle == DollarN {
			// Postgres/SQLite
			query.WriteString(" DEFAULT VALUES")
			// RETURNING در PG/SQLite معتبر است
			if len(qb.ReturningColumns) > 0 {
				query.WriteString(" RETURNING ")
				query.WriteString(strings.Join(qb.ReturningColumns, ", "))
			}
		} else {
			// MySQL
			query.WriteString(" () VALUES ()")
			// توجه: MySQL به‌طور عمومی RETURNING ندارد؛ اگر ست شده باشد،
			// اجرای کوئری احتمالاً خطا می‌دهد. می‌توانی اینجا نادیده بگیری/لاگ کنی.
		}
		return query.String(), qb.Parameters
	}

	// حالت عادی با ستون‌ها
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

	if len(qb.ReturningColumns) > 0 {
		query.WriteString(" RETURNING ")
		query.WriteString(strings.Join(qb.ReturningColumns, ", "))
	}

	return query.String(), qb.Parameters
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
	}
	if len(qb.ReturningColumns) > 0 {
		query.WriteString(" RETURNING ")
		query.WriteString(strings.Join(qb.ReturningColumns, ", "))
	}

	return query.String(), qb.Parameters
}

func (qb *QueryBuilder) buildDelete() (string, []interface{}) {
	var query strings.Builder

	query.WriteString("DELETE FROM ")
	query.WriteString(qb.Table)

	// WHERE clause
	if len(qb.Conditions) > 0 {
		query.WriteString(" WHERE ")
		qb.buildConditions(&query, qb.Conditions)
	}
	if len(qb.ReturningColumns) > 0 {
		query.WriteString(" RETURNING ")
		query.WriteString(strings.Join(qb.ReturningColumns, ", "))
	}
	return query.String(), qb.Parameters
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
				qb.Parameters = append(qb.Parameters, v)
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
			qb.Parameters = append(qb.Parameters, condition.Value)
		}
	}
}

// placeholder returns the next placeholder according to the configured style.
func (qb *QueryBuilder) placeholder() string {
	switch qb.PhStyle {
	case DollarN:
		qb.ParamIndex++
		return fmt.Sprintf("$%d", qb.ParamIndex)
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
