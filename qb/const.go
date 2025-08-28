package qb

// QueryBuilder is the main builder struct
type QueryBuilder struct {
	queryType  QueryType
	table      string
	columns    []string
	conditions []Condition
	joins      []Join
	groupBy    []string
	having     []Condition
	orderBy    []OrderBy
	limit      int
	offset     int
	insertData map[string]interface{}
	updateData map[string]interface{}
	parameters []interface{}
}

// QueryType represents different types of queries
type QueryType int

const (
	SELECT QueryType = iota
	INSERT
	UPDATE
	DELETE
)

// Operator represents comparison operators
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

// JoinType represents SQL join types
type JoinType string

const (
	INNER JoinType = "INNER JOIN"
	LEFT  JoinType = "LEFT JOIN"
	RIGHT JoinType = "RIGHT JOIN"
	FULL  JoinType = "FULL OUTER JOIN"
)

// Condition represents a where condition
type Condition struct {
	Column string
	Op     Operator
	Value  interface{}
	Logic  string // AND, OR
}

// Join represent a table join
type Join struct {
	Type      JoinType
	Table     string
	Condition string
}

// OrderBy represents ordering clause
type OrderBy struct {
	Column string
	Desc   bool
}
