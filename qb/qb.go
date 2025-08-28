package qb

import (
	"fmt"
	"reflect"
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
		GuardWrites:      true, // Default
	}
}

// WithPlaceholders sets the placeholder style (DollarN or QuestionMark)
// and resets the internal placeholder counter. It returns qb for chaining.
func (qb *QueryBuilder) WithPlaceholders(style PlaceholderStyle) *QueryBuilder {
	qb.PhStyle = style
	qb.ParamIndex = 0
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
func (qb *QueryBuilder) Reset() *QueryBuilder {
	style := qb.PhStyle

	newQB := QueryBuilder{PhStyle: style, GuardWrites: true}
	*qb = newQB

	return qb
}

// Safe re-enables write guards for this query (default behavior).
func (qb *QueryBuilder) Safe() *QueryBuilder {
	qb.GuardWrites = true
	return qb
}

// Unsafe disables write guards for this query, allowing UPDATE/ DELETE
// without a WHERE clause. Use with caution.
func (qb *QueryBuilder) Unsafe() *QueryBuilder {
	qb.GuardWrites = false
	return qb
}

// Excluded returns a RawExpr like "excluded.<col>", handy for
// ON CONFLICT DO UPDATE SET col = excluded.col (PostgreSQL/SQLite).
func Excluded(col string) RawExpr { return RawExpr("excluded." + col) }

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
