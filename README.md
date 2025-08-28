# qb — Tiny SQL Query Builder for Go

[![Go CI](https://github.com/Serajian/query-builder-GO/actions/workflows/go.yml/badge.svg)](https://github.com/Serajian/query-builder-GO/actions/workflows/go.yml)
![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go)
[![Go Reference](https://pkg.go.dev/badge/github.com/Serajian/query-builder-GO.svg)](https://pkg.go.dev/github.com/Serajian/query-builder-GO)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](#license)
[![Go Report Card](https://goreportcard.com/badge/github.com/Serajian/query-builder-GO)](https://goreportcard.com/report/github.com/Serajian/query-builder-GO)

A lightweight, ergonomic Query Builder that produces **safe placeholder SQL** with **type-safe parameters**.  
Works with PostgreSQL (`$1,$2,...`) and MySQL/SQLite (`?`). You get a plain `sql string` plus `[]any` params—perfect for `database/sql`, `sqlx`, `pgx`, etc.

---

## ✨ Features

- 🔁 **Pluggable placeholders:** `DollarN` (PostgreSQL, default) or `QuestionMark` (MySQL/SQLite).
- 🧱 **Core statements:** `SELECT`, `INSERT`, `UPDATE`, `DELETE`.
- 🔍 **Filters:** `WHERE`, `OR WHERE`, `IN/NOT IN`, `LIKE`, `IS NULL/IS NOT NULL`.
- 🔗 **Joins:** `INNER`, `LEFT`, `RIGHT`.
- 📊 **Grouping:** `GROUP BY` + `HAVING`.
- 🧭 **Ordering & Paging:** `ORDER BY`, `LIMIT`, `OFFSET`, `Paginate(page, perPage)`.
- 🧷 **Stable params:** deterministic arg order for `INSERT/UPDATE` (sorted keys).
- 🛡️ **Explicit empty-list policy:** `IN([]) → (1=0)`, `NOT IN([]) → (1=1)`—no 3-valued NULL surprises.
- ♻️ **Reusable builder:** starting a new query resets per-query state; placeholder counter restarts at each `Build()`.

---

## 📦 Installation

```bash
go get github.com/Serajian/query-builder-GO
```

---

## 🚀 Quick Start

```go
package main

import (
	"fmt"
	qb "github.com/Serajian/query-builder-GO"
)

func main() {
	sqlStr, args := qb.NewQB().
		WithPlaceholders(qb.DollarN). // or qb.QuestionMark for MySQL/SQLite
		Select("id", "name").
		From("users").
		Where("age", qb.GTE, 18).
		OrderBy("created_at").
		Limit(10).
		Build()

	fmt.Println(sqlStr)
	// SELECT id, name FROM users WHERE age >= $1 ORDER BY created_at ASC LIMIT 10

	fmt.Println(args) // [18]
}
```

---

## 🧪 Filters (IN / NOT IN, NULL, LIKE)

```go
// IN with items
sqlStr, args := qb.NewQB().
	WithPlaceholders(qb.DollarN).
	Select("*").From("users").
	WhereIn("status", []string{"active", "trial"}).
	Build()
// ... WHERE status IN ($1, $2)
// args: ["active", "trial"]

// IN with an empty list → always false
sqlEmpty, argsEmpty := qb.NewQB().
	Select("id").From("users").
	WhereIn("status", []string{}).
	Build()
// ... WHERE (1=0)
// args: []

// NOT IN with an empty list → always true
sqlNotInEmpty, _ := qb.NewQB().
	Select("id").From("users").
	WhereNotIn("role", []int{}).
	Build()
// ... WHERE (1=1)

// NULL checks
sqlNull, _ := qb.NewQB().
	Select("id").From("users").
	WhereNull("deleted_at").
	Build()
// ... WHERE deleted_at IS NULL
_ = sqlNull
```

---

## 🔗 Joins & Composed Queries

```go
sqlStr, args := qb.NewQB().
	WithPlaceholders(qb.DollarN).
	Select("u.name", "p.title", "c.name AS category").
	From("users u").
	LeftJoin("posts p", "u.id = p.user_id").
	LeftJoin("categories c", "p.category_id = c.id").
	Where("u.active", qb.EQ, true).
	WhereNotNull("p.published_at").
	OrderByDesc("p.created_at").
	Limit(5).Offset(5).
	Build()

// SELECT u.name, p.title, c.name AS category FROM users u
// LEFT JOIN posts p ON u.id = p.user_id
// LEFT JOIN categories c ON p.category_id = c.id
// WHERE u.active = $1 AND p.published_at IS NOT NULL
// ORDER BY p.created_at DESC LIMIT 5 OFFSET 5
// args: [true]
```

---

## ✍️ INSERT / UPDATE with Stable Arg Order

```go
// INSERT (keys are sorted → age, name)
sqlStr, args := qb.NewQB().
	WithPlaceholders(qb.DollarN).
	Insert("users").
	Values(map[string]any{"name": "Alice", "age": 30}).
	Build()
// INSERT INTO users (age, name) VALUES ($1, $2)
// args: [30, "Alice"]

// UPDATE (keys are sorted → age, name)
sqlStr, args = qb.NewQB().
	WithPlaceholders(qb.DollarN).
	Update("users").
	SetUpdate("name", "Bob").
	SetUpdate("age", 41).
	Where("id", qb.EQ, 9).
	Build()
// UPDATE users SET age = $1, name = $2 WHERE id = $3
// args: [41, "Bob", 9]
```

---

## 📖 API Cheatsheet

- **Config**
  - `NewQB()`
  - `WithPlaceholders(qb.DollarN | qb.QuestionMark)`
  - `Reset()` *(in-place; keeps placeholder style)*

- **Statements**
  - `Select(cols...)`, `From(table)`
  - `Insert(table)`, `Values(map[string]any)`, `Set(col, val)`
  - `Update(table)`, `SetUpdate(col, val)`
  - `Delete(table)`
  - `Build() (sql string, args []any)`

- **Filters**
  - `Where(col, op, val)`, `OrWhere(col, op, val)`
  - `WhereIn(col, slice)`, `WhereNotIn(col, slice)`
  - `WhereLike(col, pattern)`, `WhereNotLike(col, pattern)`
  - `WhereNull(col)`, `WhereNotNull(col)`
  - `GroupBy(cols...)`, `Having(col, op, val)`

- **Joins**
  - `Join(table, on)`, `LeftJoin(table, on)`, `RightJoin(table, on)`

- **Ordering & Paging**
  - `OrderBy(col)`, `OrderByDesc(col)`
  - `Limit(n)`, `Offset(n)`, `Paginate(page, perPage)`

---

## ⚙️ Placeholder Policy & Important Behaviors

- **Placeholder style**
  - Default: `DollarN` (PostgreSQL). Switch to `QuestionMark` for MySQL/SQLite.
- **LIMIT/OFFSET**
  - Always rendered **inline** in SQL (not as parameters), e.g. `LIMIT 10 OFFSET 20`.
- **Empty list semantics**
  - `IN([])` → `(1=0)` (always false)  
  - `NOT IN([])` → `(1=1)` (always true)
- **Deterministic params**
  - `INSERT/UPDATE` keys are sorted, so placeholder order always matches `args` order.
- **Reusable builder**
  - Starting a new query clears previous state; each `Build()` restarts the `$n` counter from 1.

---

## 🧰 Testing

```bash
go test ./... -v -race -covermode=atomic -coverprofile=coverage.out
go tool cover -func=coverage.out
```

### CI (GitHub Actions)

Create the file **`.github/workflows/go.yml`**:

```yaml
name: Go CI
on:
  push:
  pull_request:
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25.x'
      - run: go test ./... -v -race -covermode=atomic -coverprofile=coverage.out
      - run: go tool cover -func=coverage.out
```

Add the CI badge at the top of this README:
```md
[![Go CI](https://github.com/Serajian/query-builder-GO/actions/workflows/go.yml/badge.svg)](https://github.com/Serajian/query-builder-GO/actions/workflows/go.yml)
```

---

## 🤝 Contributing

PRs are welcome!  
Open an issue to discuss features like `FullJoin`, `RETURNING`, dialect helpers, etc.

---

## 📄 License

[MIT License](LICENSE)
