package qb

// OnConflict sets the ON CONFLICT target columns (PostgreSQL/SQLite).
// Example: OnConflict("id", "email")
func (qb *QueryBuilder) OnConflict(columns ...string) *QueryBuilder {
	qb.ConflictColumns = columns
	qb.ConflictConstraint = ""
	return qb
}

// OnConflictConstraint sets ON CONSTRAINT <name> as the conflict target.
func (qb *QueryBuilder) OnConflictConstraint(name string) *QueryBuilder {
	qb.ConflictConstraint = name
	qb.ConflictColumns = nil
	return qb
}

// OnConflictDoNothing emits ON CONFLICT ... DO NOTHING.
func (qb *QueryBuilder) OnConflictDoNothing() *QueryBuilder {
	qb.ConflictDoNothing = true
	return qb
}

// OnConflictSet adds a single assignment for ON CONFLICT ... DO UPDATE SET.
// Value may be a regular value (bound via placeholder) or a RawExpr.
// Example: OnConflictSet("name", Excluded("name"))
func (qb *QueryBuilder) OnConflictSet(column string, value interface{}) *QueryBuilder {
	if qb.ConflictUpdateSet == nil {
		qb.ConflictUpdateSet = make(map[string]interface{})
	}
	qb.ConflictDoNothing = false
	qb.ConflictUpdateSet[column] = value
	return qb
}

// OnConflictSetMap adds multiple assignments for ON CONFLICT ... DO UPDATE SET.
func (qb *QueryBuilder) OnConflictSetMap(m map[string]interface{}) *QueryBuilder {
	for k, v := range m {
		qb.OnConflictSet(k, v)
	}
	return qb
}
