package qb

import (
	"reflect"
	"strings"
	"testing"
)

func TestSelectBasic(t *testing.T) {
	sql, args := NewQB().
		WithPlaceholders(DollarN).
		Select("id", "name").
		From("users").
		Where("age", GTE, 18).
		OrderBy("created_at").
		Limit(10).
		Build()

	wantSQL := "SELECT id, name FROM users WHERE age >= $1 ORDER BY created_at ASC LIMIT 10"
	if sql != wantSQL {
		t.Fatalf("sql mismatch:\n got: %s\nwant: %s", sql, wantSQL)
	}
	wantArgs := []interface{}{18}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Fatalf("args mismatch:\n got: %#v\nwant: %#v", args, wantArgs)
	}
}

func TestSelectWithInAndEmptyPolicies(t *testing.T) {
	t.Run("IN non-empty", func(t *testing.T) {
		sql, args := NewQB().
			WithPlaceholders(DollarN).
			Select("id", "name").
			From("users").
			Where("age", GTE, 18).
			WhereIn("status", []string{"active", "trial"}).
			OrderBy("created_at").
			Limit(10).
			Build()

		wantSQL := "SELECT id, name FROM users WHERE age >= $1 AND status IN ($2, $3) ORDER BY created_at ASC LIMIT 10"
		if sql != wantSQL {
			t.Fatalf("sql mismatch:\n got: %s\nwant: %s", sql, wantSQL)
		}
		wantArgs := []interface{}{18, "active", "trial"}
		if !reflect.DeepEqual(args, wantArgs) {
			t.Fatalf("args mismatch:\n got: %#v\nwant: %#v", args, wantArgs)
		}
	})

	t.Run("IN empty -> (1=0)", func(t *testing.T) {
		sql, args := NewQB().
			WithPlaceholders(DollarN).
			Select("id").
			From("users").
			WhereIn("status", []string{}).
			Build()

		if !strings.Contains(sql, "(1=0)") {
			t.Fatalf("expected (1=0) for empty IN, got: %s", sql)
		}
		if len(args) != 0 {
			t.Fatalf("expected no args, got: %#v", args)
		}
	})

	t.Run("NOT IN empty -> (1=1)", func(t *testing.T) {
		sql, args := NewQB().
			WithPlaceholders(DollarN).
			Select("id").
			From("users").
			WhereNotIn("status", []string{}).
			Build()

		if !strings.Contains(sql, "(1=1)") {
			t.Fatalf("expected (1=1) for empty NOT IN, got: %s", sql)
		}
		if len(args) != 0 {
			t.Fatalf("expected no args, got: %#v", args)
		}
	})
}

func TestSelectNullAndJoinsAndOffset(t *testing.T) {
	sql, args := NewQB().
		WithPlaceholders(DollarN).
		Select("u.name", "p.title").
		From("users u").
		LeftJoin("posts p", "u.id = p.user_id").
		Where("u.active", EQ, true).
		WhereNotNull("p.published_at").
		OrderByDesc("p.created_at").
		Limit(5).
		Offset(5).
		Build()

	want := "SELECT u.name, p.title FROM users u LEFT JOIN posts p ON u.id = p.user_id " +
		"WHERE u.active = $1 AND p.published_at IS NOT NULL ORDER BY p.created_at DESC LIMIT 5 OFFSET 5"
	if sql != want {
		t.Fatalf("sql mismatch:\n got: %s\nwant: %s", sql, want)
	}
	wantArgs := []interface{}{true}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Fatalf("args mismatch:\n got: %#v\nwant: %#v", args, wantArgs)
	}
}

func TestInsertSortedKeysAndPlaceholders(t *testing.T) {
	sql, args := NewQB().
		WithPlaceholders(DollarN).
		Insert("users").
		Values(map[string]interface{}{
			"name": "Alice",
			"age":  30,
		}).
		Build()

	wantSQL := "INSERT INTO users (age, name) VALUES ($1, $2)"
	if sql != wantSQL {
		t.Fatalf("sql mismatch:\n got: %s\nwant: %s", sql, wantSQL)
	}
	wantArgs := []interface{}{30, "Alice"}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Fatalf("args mismatch:\n got: %#v\nwant: %#v", args, wantArgs)
	}
}

func TestUpdateSortedKeysAndWhere(t *testing.T) {
	sql, args := NewQB().
		WithPlaceholders(DollarN).
		Update("users").
		SetUpdate("name", "Bob").
		SetUpdate("age", 41).
		Where("id", EQ, 9).
		Build()

	wantSQL := "UPDATE users SET age = $1, name = $2 WHERE id = $3"
	if sql != wantSQL {
		t.Fatalf("sql mismatch:\n got: %s\nwant: %s", sql, wantSQL)
	}
	wantArgs := []interface{}{41, "Bob", 9}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Fatalf("args mismatch:\n got: %#v\nwant: %#v", args, wantArgs)
	}
}

func TestDeleteWithWhere(t *testing.T) {
	sql, args := NewQB().
		WithPlaceholders(DollarN).
		Delete("users").
		Where("id", EQ, 5).
		Build()

	wantSQL := "DELETE FROM users WHERE id = $1"
	if sql != wantSQL {
		t.Fatalf("sql mismatch:\n got: %s\nwant: %s", sql, wantSQL)
	}
	wantArgs := []interface{}{5}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Fatalf("args mismatch:\n got: %#v\nwant: %#v", args, wantArgs)
	}
}

func TestBuildResetsParamIndex(t *testing.T) {
	b := NewQB().WithPlaceholders(DollarN)

	sql1, args1 := b.
		Select("id").
		From("users").
		Where("age", GT, 18).
		Build()
	if !strings.Contains(sql1, "$1") {
		t.Fatalf("first build should start with $1, got: %s", sql1)
	}
	if len(args1) != 1 || args1[0] != 18 {
		t.Fatalf("args mismatch on first build: %#v", args1)
	}

	sql2, args2 := b.
		Select("id").
		From("users").
		Where("age", GT, 21).
		Build()

	if !strings.Contains(sql2, "$1") {
		t.Fatalf("second build should start with $1, got: %s", sql2)
	}
	if len(args2) != 1 || args2[0] != 21 {
		t.Fatalf("args mismatch on second build: %#v", args2)
	}
}

func TestInsertReturning(t *testing.T) {
	sql, args := NewQB().
		WithPlaceholders(DollarN).
		Insert("users").
		Values(map[string]any{"name": "A", "age": 1}).
		Returning("id").
		Build()

	wantPrefix := "INSERT INTO users"
	if !strings.HasPrefix(sql, wantPrefix) || !strings.Contains(sql, " RETURNING id") {
		t.Fatalf("unexpected sql: %s", sql)
	}
	if len(args) != 2 {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestSelectStarDefault(t *testing.T) {
	sql, _ := NewQB().Select().From("users").Build()
	if !strings.HasPrefix(sql, "SELECT * FROM users") {
		t.Fatalf("expected SELECT * by default, got: %s", sql)
	}
}

func TestInsertDefaultValues_Postgres(t *testing.T) {
	sql, args := NewQB().
		WithPlaceholders(DollarN).
		Insert("users").
		Build()
	want := "INSERT INTO users DEFAULT VALUES"
	if sql != want {
		t.Fatalf("got: %s, want: %s", sql, want)
	}
	if len(args) != 0 {
		t.Fatalf("want no args, got: %#v", args)
	}
}

func TestQuestionMarkPlaceholdersSelect(t *testing.T) {
	sql, args := NewQB().
		WithPlaceholders(QuestionMark).
		Select("id").
		From("t").
		Where("x", EQ, 1).
		Build()

	if strings.Contains(sql, "$") {
		t.Fatalf("expected question mark placeholders, got: %s", sql)
	}
	if !strings.Contains(sql, "WHERE x = ?") {
		t.Fatalf("expected WHERE x = ?, got: %s", sql)
	}
	if len(args) != 1 || args[0] != 1 {
		t.Fatalf("args mismatch: %#v", args)
	}
}

func TestOrderByMultiple(t *testing.T) {
	sql, _ := NewQB().
		WithPlaceholders(DollarN).
		Select("id").
		From("users").
		OrderBy("name").
		OrderByDesc("created_at").
		Build()

	wantFrag := "ORDER BY name ASC, created_at DESC"
	if !strings.Contains(sql, wantFrag) {
		t.Fatalf("expected %q, got: %s", wantFrag, sql)
	}
}

func TestPaginateHelper(t *testing.T) {
	sql, _ := NewQB().
		Select("*").
		From("users").
		Paginate(3, 25). // page 3 -> LIMIT 25 OFFSET 50
		Build()

	if !strings.Contains(sql, "LIMIT 25") || !strings.Contains(sql, "OFFSET 50") {
		t.Fatalf("expected LIMIT 25 OFFSET 50, got: %s", sql)
	}
}

func TestGroupByHaving(t *testing.T) {
	sql, args := NewQB().
		WithPlaceholders(DollarN).
		Select("status", "COUNT(*) AS cnt").
		From("users").
		GroupBy("status").
		Having("COUNT(*)", GT, 10).
		Build()

	wantFrag := "GROUP BY status HAVING COUNT(*) > $1"
	if !strings.Contains(sql, wantFrag) {
		t.Fatalf("expected %q in sql, got: %s", wantFrag, sql)
	}
	if len(args) != 1 || args[0] != 10 {
		t.Fatalf("args mismatch: %#v", args)
	}
}

func TestJoinsVariants(t *testing.T) {
	sql, _ := NewQB().
		Select("u.id", "o.id").
		From("users u").
		Join("orders o", "o.user_id = u.id").
		RightJoin("payments p", "p.user_id = u.id").
		Build()

	want := "SELECT u.id, o.id FROM users u INNER JOIN orders o ON o.user_id = u.id RIGHT JOIN payments p ON p.user_id = u.id"
	if sql != want {
		t.Fatalf("sql mismatch:\n got: %s\nwant: %s", sql, want)
	}
}

func TestResetKeepsPlaceholderStyle(t *testing.T) {
	b := NewQB().WithPlaceholders(QuestionMark)
	b.Reset()
	sql, _ := b.Select("id").From("t").Where("x", EQ, 1).Build()
	if !strings.Contains(sql, "WHERE x = ?") {
		t.Fatalf("expected '?' placeholders after Reset, got: %s", sql)
	}
}

func TestLikeAndNotLike(t *testing.T) {
	sql, args := NewQB().
		WithPlaceholders(DollarN).
		Select("id").
		From("users").
		WhereLike("name", "%ali%").
		WhereNotLike("email", "%@spam.com").
		Build()

	wantA := "name LIKE $1"
	wantB := "email NOT LIKE $2"
	if !strings.Contains(sql, wantA) || !strings.Contains(sql, wantB) {
		t.Fatalf("expected %q and %q in sql, got: %s", wantA, wantB, sql)
	}
	if len(args) != 2 || args[0] != "%ali%" || args[1] != "%@spam.com" {
		t.Fatalf("args mismatch: %#v", args)
	}
}

func TestDeleteReturningPostgres(t *testing.T) {
	sql, args := NewQB().
		WithPlaceholders(DollarN).
		Delete("users").
		Where("id", EQ, 7).
		Returning("*").
		Build()

	want := "DELETE FROM users WHERE id = $1 RETURNING *"
	if sql != want {
		t.Fatalf("sql mismatch:\n got: %s\nwant: %s", sql, want)
	}
	if len(args) != 1 || args[0] != 7 {
		t.Fatalf("args mismatch: %#v", args)
	}
}

func TestOffsetOnly(t *testing.T) {
	sql, _ := NewQB().
		WithPlaceholders(DollarN).
		Select("id").
		From("t").
		Offset(15).
		Build()

	if strings.Contains(sql, "LIMIT ") || !strings.Contains(sql, "OFFSET 15") {
		t.Fatalf("expected only OFFSET 15, got: %s", sql)
	}
}

func TestWhereInBytesAsSingleValue(t *testing.T) {
	val := []byte{1, 2, 3}
	sql, args := NewQB().
		WithPlaceholders(DollarN).
		Select("id").
		From("t").
		WhereIn("blob", val).
		Build()

	want := "blob IN ($1)"
	if !strings.Contains(sql, want) {
		t.Fatalf("expected %q in sql, got: %s", want, sql)
	}
	if len(args) != 1 || !reflect.DeepEqual(args[0], val) {
		t.Fatalf("args mismatch: %#v", args)
	}
}

func TestWhereInScalarTreatedAsEmpty(t *testing.T) {
	sql, args := NewQB().
		Select("id").
		From("t").
		WhereIn("id", 5). // not a slice → treated as empty
		Build()

	if !strings.Contains(sql, "(1=0)") {
		t.Fatalf("expected (1=0) for scalar IN, got: %s", sql)
	}
	if len(args) != 0 {
		t.Fatalf("expected no args, got: %#v", args)
	}
}

func TestANDORCombination(t *testing.T) {
	sql, args := NewQB().
		WithPlaceholders(DollarN).
		Select("*").
		From("t").
		Where("a", EQ, 1).   // a = $1
		OrWhere("b", EQ, 2). // OR b = $2
		Where("c", EQ, 3).   // AND c = $3
		Build()

	want := "SELECT * FROM t WHERE a = $1 OR b = $2 AND c = $3"
	if sql != want {
		t.Fatalf("sql mismatch:\n got: %s\nwant: %s", sql, want)
	}
	wantArgs := []interface{}{1, 2, 3}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Fatalf("args mismatch:\n got: %#v\nwant: %#v", args, wantArgs)
	}
}

func TestInsertWithSetOnly(t *testing.T) {
	sql, args := NewQB().
		WithPlaceholders(DollarN).
		Insert("users").
		Set("name", "A").
		Set("age", 32).
		Build()

	// keys sorted: age, name
	want := "INSERT INTO users (age, name) VALUES ($1, $2)"
	if sql != want {
		t.Fatalf("sql mismatch:\n got: %s\nwant: %s", sql, want)
	}
	wantArgs := []interface{}{32, "A"}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Fatalf("args mismatch:\n got: %#v\nwant: %#v", args, wantArgs)
	}
}

func TestUpdateReturning(t *testing.T) {
	sql, args := NewQB().
		WithPlaceholders(DollarN).
		Update("users").
		SetUpdate("name", "B").
		Returning("id").
		Build()

	want := "UPDATE users SET name = $1 WHERE 1=0 /*guarded: mising WHERE */ RETURNING id"
	if sql != want {
		t.Fatalf("sql mismatch:\n got: %s\nwant: %s", sql, want)
	}
	if len(args) != 1 || args[0] != "B" {
		t.Fatalf("args mismatch: %#v", args)
	}
}

func TestUpdateGuardedWithoutWhere_Default(t *testing.T) {
	sql, args := NewQB().
		WithPlaceholders(DollarN).
		Update("users").
		SetUpdate("role", "admin").
		Build()

	if !strings.Contains(sql, "WHERE 1=0") {
		t.Fatalf("expected guard to add WHERE 1=0, got: %s", sql)
	}
	if len(args) != 1 || args[0] != "admin" {
		t.Fatalf("args mismatch: %#v", args)
	}
}

func TestUpdateUnsafeWithoutWhere(t *testing.T) {
	sql, args := NewQB().
		WithPlaceholders(DollarN).
		Update("users").
		SetUpdate("role", "admin").
		Unsafe().
		Build()

	if strings.Contains(sql, "WHERE 1=0") {
		t.Fatalf("did not expect guard, got: %s", sql)
	}
	if !strings.HasPrefix(sql, "UPDATE users SET role = $1") {
		t.Fatalf("unexpected sql: %s", sql)
	}
	if len(args) != 1 || args[0] != "admin" {
		t.Fatalf("args mismatch: %#v", args)
	}
}

func TestDeleteGuardedWithoutWhere(t *testing.T) {
	sql, _ := NewQB().
		WithPlaceholders(DollarN).
		Delete("users").
		Build()

	if !strings.Contains(sql, "DELETE FROM users WHERE 1=0") {
		t.Fatalf("expected guarded delete, got: %s", sql)
	}
}

func TestGuardResetsToSafeAfterBuild(t *testing.T) {
	b := NewQB().WithPlaceholders(DollarN)

	// First, disable guard for this build
	sql1, _ := b.Update("users").SetUpdate("x", 1).Unsafe().Build()
	if strings.Contains(sql1, "WHERE 1=0") {
		t.Fatalf("unexpected guard on unsafe build: %s", sql1)
	}

	// Next build should be safe again (guard ON by default after Reset)
	sql2, _ := b.Update("users").SetUpdate("x", 2).Build()
	if !strings.Contains(sql2, "WHERE 1=0") {
		t.Fatalf("expected guard to be ON after reset, got: %s", sql2)
	}
}

func TestInsertOnConflictDoNothing_PG(t *testing.T) {
	sql, args := NewQB().
		WithPlaceholders(DollarN).
		Insert("users").
		Values(map[string]any{"id": 1, "name": "A"}).
		OnConflict("id").
		OnConflictDoNothing().
		Build()

	wantFrag := "ON CONFLICT (id) DO NOTHING"
	if !strings.Contains(sql, wantFrag) {
		t.Fatalf("expected %q in sql, got: %s", wantFrag, sql)
	}
	if len(args) != 2 || args[0] != 1 || args[1] != "A" {
		t.Fatalf("args mismatch: %#v", args)
	}
}

func TestInsertOnConflictDoUpdate_PG_WithValuesAndExcluded(t *testing.T) {
	sql, args := NewQB().
		WithPlaceholders(DollarN).
		Insert("users").
		Values(map[string]any{"id": 1, "name": "A"}).
		OnConflict("id").
		OnConflictSet("age", 30).                // bind param → $3
		OnConflictSet("name", Excluded("name")). // raw expr → excluded.name
		Build()

	// keys in update set are sorted: age, name
	wantA := "ON CONFLICT (id) DO UPDATE SET age = $3, name = excluded.name"
	if !strings.Contains(sql, wantA) {
		t.Fatalf("expected %q in sql, got: %s", wantA, sql)
	}

	// params: [$1=id, $2=name, $3=age]
	wantArgs := []any{1, "A", 30}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Fatalf("args mismatch:\n got: %#v\nwant: %#v", args, wantArgs)
	}
}

func TestInsertOnConflictConstraint_PG(t *testing.T) {
	sql, _ := NewQB().
		WithPlaceholders(DollarN).
		Insert("users").
		Values(map[string]any{"id": 1}).
		OnConflictConstraint("users_pkey").
		OnConflictDoNothing().
		Build()

	want := "ON CONFLICT ON CONSTRAINT users_pkey DO NOTHING"
	if !strings.Contains(sql, want) {
		t.Fatalf("expected %q in sql, got: %s", want, sql)
	}
}

func TestInsertDefaultValuesReturning_PG(t *testing.T) {
	sql, args := NewQB().
		WithPlaceholders(DollarN).
		Insert("users").
		Returning("id").
		Build()

	want := "INSERT INTO users DEFAULT VALUES RETURNING id"
	if sql != want {
		t.Fatalf("sql mismatch:\n got: %s\nwant: %s", sql, want)
	}
	if len(args) != 0 {
		t.Fatalf("expected no args, got: %#v", args)
	}
}

func TestInsertNoConflict_MySQL_NoReturning(t *testing.T) {
	sql, args := NewQB().
		WithPlaceholders(QuestionMark).
		Insert("users").
		Values(map[string]any{"id": 1, "name": "A"}).
		OnConflict("id").      // should be ignored on MySQL path
		OnConflictDoNothing(). // should be ignored on MySQL path
		Returning("id").       // MySQL generally doesn't support RETURNING
		Build()

	if strings.Contains(sql, "ON CONFLICT") || strings.Contains(sql, "RETURNING") {
		t.Fatalf("did not expect ON CONFLICT/RETURNING for MySQL, got: %s", sql)
	}
	if !strings.Contains(sql, "VALUES (?, ?)") {
		t.Fatalf("expected question mark placeholders, got: %s", sql)
	}
	wantArgs := []any{1, "A"}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Fatalf("args mismatch:\n got: %#v\nwant: %#v", args, wantArgs)
	}
}

func TestInsertDefaultValues_MySQL(t *testing.T) {
	sql, args := NewQB().
		WithPlaceholders(QuestionMark).
		Insert("users").
		Build()

	want := "INSERT INTO users () VALUES ()"
	if sql != want {
		t.Fatalf("sql mismatch:\n got: %s\nwant: %s", sql, want)
	}
	if len(args) != 0 {
		t.Fatalf("expected no args, got: %#v", args)
	}
}
