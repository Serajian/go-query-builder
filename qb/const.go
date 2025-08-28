package qb

// QueryBuilder is a tiny, chainable SQL query builder that renders a SQL string
// plus its bound parameters. It supports SELECT/ INSERT/ UPDATE/ DELETE, WHERE/IN,
// JOINs, GROUP BY/HAVING, ORDER BY, LIMIT/OFFSET, and RETURNING.
type QueryBuilder struct {
	// QueryType is the kind of statement to build (SELECT/ INSERT/ UPDATE/ DELETE).
	QueryType QueryType
	// Table is the target table name (as written into SQL).
	Table string
	// Columns holds selected columns for SELECT or is used for rendering parts that list columns.
	Columns []string
	// Conditions are the WHERE conditions for SELECT/ UPDATE/ DELETE.
	Conditions []Condition
	// Joins lists JOIN clauses for SELECT queries.
	Joins []Join
	// GroupByColumns are the columns used in GROUP BY.
	GroupByColumns []string
	// HavingConditions are the HAVING conditions applied after GROUP BY.
	HavingConditions []Condition
	// OrderByArr is the ORDER BY clause specification.
	OrderByArr []OrderBy
	// LimitInt renders as LIMIT n when > 0.
	LimitInt int
	// OffsetInt renders as OFFSET n when > 0.
	OffsetInt int
	// InsertData holds column->value pairs for INSERT.
	InsertData map[string]interface{}
	// UpdateData holds column->value pairs for UPDATE SET.
	UpdateData map[string]interface{}
	// Parameters accumulates bound values in render order.
	Parameters []interface{}
	// PhStyle selects placeholder style (DollarN=$1,$2,... or QuestionMark=?).
	PhStyle PlaceholderStyle
	// ParamIndex tracks the next placeholder index for DollarN style.
	ParamIndex int
	// ReturningColumns lists columns for RETURNING (PostgreSQL/SQLite 3.35+).
	ReturningColumns []string
}

// PlaceholderStyle controls how placeholders are rendered.
//   - DollarN:    $1, $2, ... (PostgreSQL)
//   - QuestionMark: ?         (MySQL/SQLite)
type PlaceholderStyle int

const (
	// QuestionMark uses '?' placeholders (e.g., MySQL, SQLite).
	QuestionMark PlaceholderStyle = iota
	// DollarN uses '$1', '$2', ... placeholders (e.g., PostgreSQL).
	DollarN
)

// QueryType represents the statement being built.
type QueryType int

const (
	// SELECT builds a SELECT statement.
	SELECT QueryType = iota
	// INSERT builds an INSERT statement.
	INSERT
	// UPDATE builds an UPDATE statement.
	UPDATE
	// DELETE builds a DELETE statement.
	DELETE
)

// Operator enumerates supported comparison operators for WHERE/HAVING clauses.
//
//	EQ      = "="
//	NEQ     = "!="
//	GT      = ">"
//	GTE     = ">="
//	LT      = "<"
//	LTE     = "<="
//	IN      = "IN"
//	NIN     = "NOT IN"
//	NULL    = "IS NULL"
//	NOTNULL = "IS NOT NULL"
//	LIKE    = "LIKE"
//	NOTLIKE = "NOT LIKE"
type Operator string

const (
	EQ      Operator = "="
	NEQ     Operator = "!="
	GT      Operator = ">"
	GTE     Operator = ">="
	LT      Operator = "<"
	LTE     Operator = "<="
	IN      Operator = "IN"
	NIN     Operator = "NOT IN"
	NULL    Operator = "IS NULL"
	NOTNULL Operator = "IS NOT NULL"
	LIKE    Operator = "LIKE"
	NOTLIKE Operator = "NOT LIKE"
)

// JoinType declares supported SQL JOIN types.
//
//	INNER = "INNER JOIN"
//	LEFT  = "LEFT JOIN"
//	RIGHT = "RIGHT JOIN"
//	FULL  = "FULL OUTER JOIN"
type JoinType string

const (
	INNER JoinType = "INNER JOIN"
	LEFT  JoinType = "LEFT JOIN"
	RIGHT JoinType = "RIGHT JOIN"
	FULL  JoinType = "FULL OUTER JOIN"
)

// Condition represents a single boolean predicate (e.g., "age >= 18").
// Logic indicates how it combines with the previous condition ("AND" / "OR").
type Condition struct {
	Column string
	Op     Operator
	Value  interface{}
	Logic  string
}

// Join represents a table join: "Type Table ON Condition".
type Join struct {
	Type      JoinType
	Table     string
	Condition string
}

// OrderBy configures ORDER BY column and direction.
type OrderBy struct {
	Column string
	Desc   bool
}
