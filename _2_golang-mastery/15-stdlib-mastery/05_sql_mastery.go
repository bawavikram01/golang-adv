//go:build ignore

// =============================================================================
// LESSON 15.5: database/sql — The Right Way to Talk to Databases in Go
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - database/sql architecture: DB, Conn, Tx, Stmt, Rows
// - Connection pooling: how it works, tuning parameters
// - Query patterns: QueryRow, Query, Exec, prepared statements
// - Transactions: Begin/Commit/Rollback, isolation levels
// - Scanning: Scan, StructScan, NullString/NullInt64
// - Context-aware queries (cancellation, timeouts)
// - Production patterns: health checks, graceful shutdown, migrations
// - Common pitfalls: connection leaks, N+1, SQL injection
//
// THE KEY INSIGHT:
// database/sql is NOT an ORM. It's a thin abstraction over database drivers
// that manages connection pooling for you. The pool is goroutine-safe.
// Every Query/Exec may use a different connection — that's by design.
// Understanding this prevents the #1 bug: leaked connections.
//
// NOTE: This file demonstrates patterns without a live database.
// Comments show the exact code you'd use with a real driver (pgx, mysql, etc.)
//
// RUN: go run 05_sql_mastery.go
// =============================================================================

package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func main() {
	fmt.Println("=== DATABASE/SQL MASTERY ===")
	fmt.Println()

	architectureOverview()
	connectionPooling()
	queryPatterns()
	transactionPatterns()
	scanningPatterns()
	contextPatterns()
	productionPatterns()
	commonPitfalls()
}

// =============================================================================
// PART 1: Architecture Overview
// =============================================================================
func architectureOverview() {
	fmt.Println("--- ARCHITECTURE OVERVIEW ---")

	// database/sql is an INTERFACE layer. You need a DRIVER to talk to a DB.
	//
	// Popular drivers:
	//   PostgreSQL: github.com/jackc/pgx/v5/stdlib (pgx) ← recommended
	//               github.com/lib/pq (legacy, unmaintained)
	//   MySQL:      github.com/go-sql-driver/mysql
	//   SQLite:     modernc.org/sqlite (pure Go, no CGo!)
	//               github.com/mattn/go-sqlite3 (CGo-based)
	//
	// THE OBJECT HIERARCHY:
	//
	// sql.DB ───────────────────────── Connection pool (safe for concurrent use)
	//   │
	//   ├── sql.Conn ─────────────── Single connection (for session-level state)
	//   │
	//   ├── sql.Tx ───────────────── Transaction (bound to ONE connection)
	//   │     │
	//   │     ├── tx.QueryRow()   ── Query within transaction
	//   │     ├── tx.Exec()       ── Execute within transaction
	//   │     └── tx.Prepare()    ── Prepared statement within transaction
	//   │
	//   ├── sql.Stmt ─────────────── Prepared statement (can span connections)
	//   │
	//   └── sql.Rows ─────────────── Result set (MUST be closed!)
	//
	// IMPORTANT:
	// - sql.DB is NOT a single connection. It's a POOL.
	// - sql.DB is safe for concurrent use from multiple goroutines.
	// - You typically create ONE sql.DB for the entire application lifetime.
	// - sql.DB opens/closes connections lazily.

	// ─── Opening a database ───
	// import _ "github.com/jackc/pgx/v5/stdlib"  // register the driver
	//
	// db, err := sql.Open("pgx", "postgres://user:pass@localhost:5432/mydb?sslmode=disable")
	// if err != nil {
	//     log.Fatal(err)  // only checks DSN format, NOT connectivity!
	// }
	// defer db.Close()
	//
	// // VERIFY connectivity:
	// if err := db.Ping(); err != nil {  // or PingContext with timeout
	//     log.Fatal("cannot reach database:", err)
	// }
	//
	// NOTE: sql.Open does NOT establish a connection!
	// It only validates the driver name and DSN format.
	// The first actual connection happens on first query or Ping().

	fmt.Println("  sql.DB = connection pool (one per app, goroutine-safe)")
	fmt.Println("  sql.Open validates DSN format only — Ping() tests connectivity")
	fmt.Println("  Drivers: pgx (Postgres), go-sql-driver (MySQL), modernc/sqlite")
	fmt.Println()
}

// =============================================================================
// PART 2: Connection Pooling
// =============================================================================
func connectionPooling() {
	fmt.Println("--- CONNECTION POOLING ---")

	// sql.DB manages a pool of connections automatically.
	//
	// CONFIGURATION:
	// ──────────────
	// db.SetMaxOpenConns(25)         // Max connections to the database (default: unlimited!)
	// db.SetMaxIdleConns(10)         // Max idle connections kept in pool (default: 2)
	// db.SetConnMaxLifetime(5*time.Minute)  // Max time a conn can be reused (default: forever)
	// db.SetConnMaxIdleTime(1*time.Minute)  // Max time a conn can be idle (default: forever)
	//
	// RULES OF THUMB:
	// ───────────────
	// MaxOpenConns:
	//   - Start with 25 for most apps
	//   - Too many → database overload, more memory, lock contention
	//   - Too few → goroutines block waiting for a connection
	//   - PostgreSQL default max_connections: 100 (shared across ALL clients!)
	//
	// MaxIdleConns:
	//   - Set to ~50% of MaxOpenConns
	//   - Higher = faster (reuses existing connections)
	//   - Lower = less memory but more connection setup overhead
	//   - MUST be <= MaxOpenConns (or connections are opened/closed constantly)
	//
	// ConnMaxLifetime:
	//   - 5 minutes is a good default
	//   - Must be < database server's wait_timeout
	//   - Prevents using stale connections after DB failover
	//   - Set shorter for cloud databases (PgBouncer, RDS Proxy)
	//
	// ConnMaxIdleTime:
	//   - 1-5 minutes
	//   - Frees idle connections during low traffic
	//   - Reduces connection count to database
	//
	// MONITORING:
	// ───────────
	// stats := db.Stats()
	// stats.OpenConnections  — currently open connections
	// stats.InUse            — connections currently in use
	// stats.Idle             — idle connections in pool
	// stats.WaitCount        — total number of waits for a connection
	// stats.WaitDuration     — total wait time
	// stats.MaxIdleClosed    — connections closed due to MaxIdleConns
	// stats.MaxLifetimeClosed — connections closed due to ConnMaxLifetime
	//
	// PRODUCTION SETUP EXAMPLE:
	// db.SetMaxOpenConns(25)
	// db.SetMaxIdleConns(10)
	// db.SetConnMaxLifetime(5 * time.Minute)
	// db.SetConnMaxIdleTime(1 * time.Minute)

	// Demonstrate Stats struct fields
	stats := sql.DBStats{}
	fmt.Printf("  Pool stats: Open=%d, InUse=%d, Idle=%d, WaitCount=%d\n",
		stats.OpenConnections, stats.InUse, stats.Idle, stats.WaitCount)

	fmt.Println()
}

// =============================================================================
// PART 3: Query Patterns
// =============================================================================
func queryPatterns() {
	fmt.Println("--- QUERY PATTERNS ---")

	// ─── QueryRow: expect exactly ONE row ───
	// var name string
	// var age int
	// err := db.QueryRowContext(ctx,
	//     "SELECT name, age FROM users WHERE id = $1", userID,
	// ).Scan(&name, &age)
	//
	// switch {
	// case errors.Is(err, sql.ErrNoRows):
	//     // Not found — NOT an error in most cases
	//     return nil, ErrNotFound
	// case err != nil:
	//     return nil, fmt.Errorf("query user: %w", err)
	// }
	fmt.Println("  QueryRow: single row, check sql.ErrNoRows for not-found")

	// ─── Query: expect multiple rows ───
	// rows, err := db.QueryContext(ctx,
	//     "SELECT id, name FROM users WHERE active = $1 ORDER BY name", true,
	// )
	// if err != nil {
	//     return nil, err
	// }
	// defer rows.Close()  // ⚠️ ALWAYS close! Leaks connection if not closed.
	//
	// var users []User
	// for rows.Next() {
	//     var u User
	//     if err := rows.Scan(&u.ID, &u.Name); err != nil {
	//         return nil, err
	//     }
	//     users = append(users, u)
	// }
	// // Check for errors from iteration
	// if err := rows.Err(); err != nil {
	//     return nil, err
	// }
	// return users, nil
	fmt.Println("  Query: multiple rows, MUST defer rows.Close(), check rows.Err()")

	// ─── Exec: INSERT, UPDATE, DELETE (no rows returned) ───
	// result, err := db.ExecContext(ctx,
	//     "INSERT INTO users (name, email) VALUES ($1, $2)", name, email,
	// )
	// if err != nil {
	//     return err
	// }
	// id, _ := result.LastInsertId()    // MySQL only (Postgres: use RETURNING)
	// affected, _ := result.RowsAffected()
	fmt.Println("  Exec: INSERT/UPDATE/DELETE, returns RowsAffected")

	// ─── Postgres RETURNING clause (use QueryRow, not Exec) ───
	// var id int64
	// err := db.QueryRowContext(ctx,
	//     "INSERT INTO users (name) VALUES ($1) RETURNING id", name,
	// ).Scan(&id)
	fmt.Println("  Postgres RETURNING: use QueryRow to get inserted ID")

	// ─── Prepared Statements ───
	// stmt, err := db.PrepareContext(ctx, "SELECT name FROM users WHERE id = $1")
	// if err != nil { return err }
	// defer stmt.Close()
	//
	// // Reuse for multiple queries (saves query parsing on the DB side)
	// for _, id := range userIDs {
	//     var name string
	//     stmt.QueryRowContext(ctx, id).Scan(&name)
	// }
	//
	// NOTE: Prepared statements pin to a connection. In a pool, Go handles
	// re-preparing on different connections transparently, but there's overhead.
	// Only worth it for high-frequency queries.
	fmt.Println("  Prepared statements: worth it for repeated high-frequency queries")

	// ─── PLACEHOLDER SYNTAX BY DRIVER ───
	// PostgreSQL: $1, $2, $3  (positional)
	// MySQL:      ?, ?, ?     (positional)
	// SQLite:     ?, ?, ?     (or $1, $2, :name)
	//
	// NEVER build SQL with string concatenation!
	// BAD:  db.Query("SELECT * FROM users WHERE name = '" + name + "'")  // SQL INJECTION!
	// GOOD: db.Query("SELECT * FROM users WHERE name = $1", name)

	fmt.Println()
}

// =============================================================================
// PART 4: Transactions
// =============================================================================
func transactionPatterns() {
	fmt.Println("--- TRANSACTION PATTERNS ---")

	// ─── Basic transaction ───
	// tx, err := db.BeginTx(ctx, nil)  // nil = default isolation level
	// if err != nil { return err }
	// // defer Rollback is safe even after Commit (becomes no-op)
	// defer tx.Rollback()
	//
	// _, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, fromID)
	// if err != nil { return err }
	//
	// _, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, toID)
	// if err != nil { return err }
	//
	// return tx.Commit()
	fmt.Println("  Basic: BeginTx → operations → Commit (defer Rollback for safety)")

	// ─── PRODUCTION PATTERN: Transaction helper function ───
	// Eliminates the "forgot to rollback/commit" class of bugs.
	//
	// func withTx(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
	//     tx, err := db.BeginTx(ctx, nil)
	//     if err != nil {
	//         return fmt.Errorf("begin tx: %w", err)
	//     }
	//     defer tx.Rollback() // no-op after commit
	//
	//     if err := fn(tx); err != nil {
	//         return err  // Rollback called by defer
	//     }
	//     return tx.Commit()
	// }
	//
	// // Usage:
	// err := withTx(ctx, db, func(tx *sql.Tx) error {
	//     _, err := tx.ExecContext(ctx, "INSERT ...", args...)
	//     if err != nil { return err }
	//     _, err = tx.ExecContext(ctx, "UPDATE ...", args...)
	//     return err  // if nil, Commit; if error, Rollback
	// })
	fmt.Println("  Pattern: withTx helper prevents forgot-to-commit/rollback bugs")

	// ─── Isolation Levels ───
	// tx, err := db.BeginTx(ctx, &sql.TxOptions{
	//     Isolation: sql.LevelSerializable,  // strictest
	//     ReadOnly:  true,                    // hint to DB (may optimize)
	// })
	//
	// Levels (weakest → strictest):
	//   sql.LevelDefault           — database default
	//   sql.LevelReadUncommitted   — can see uncommitted data (dirty reads)
	//   sql.LevelReadCommitted     — only sees committed data (most Postgres default)
	//   sql.LevelRepeatableRead    — snapshot at tx start (MySQL InnoDB default)
	//   sql.LevelSerializable      — full serial execution (slowest, safest)
	fmt.Println("  Isolation: ReadCommitted (Postgres default), RepeatableRead (MySQL)")

	// ─── GOTCHA: Don't use db.Query inside a tx! ───
	// USE tx.QueryContext, tx.ExecContext inside transactions.
	// db.QueryContext may use a DIFFERENT connection than your transaction!

	fmt.Println()
}

// =============================================================================
// PART 5: Scanning Patterns
// =============================================================================
func scanningPatterns() {
	fmt.Println("--- SCANNING PATTERNS ---")

	// ─── Basic Scan: positional mapping ───
	// var id int64
	// var name string
	// var email sql.NullString  // nullable column
	// err := row.Scan(&id, &name, &email)
	fmt.Println("  Basic Scan: positional, order must match SELECT columns")

	// ─── Nullable columns: sql.NullXxx types ───
	// sql.NullString{String: "", Valid: false}  — NULL
	// sql.NullString{String: "hello", Valid: true}  — "hello"
	// sql.NullInt64, sql.NullFloat64, sql.NullBool, sql.NullTime
	//
	// Or in Go 1.22+: use *T (pointer) directly
	// var name *string
	// row.Scan(&name)  // name == nil for NULL
	//
	// WHICH TO USE?
	// - sql.NullString: when you need to distinguish "" from NULL
	// - *string: simpler, NULL → nil, but "" and nil behave differently in JSON

	var ns sql.NullString
	ns = sql.NullString{String: "", Valid: false}
	fmt.Printf("  NullString (NULL): Value=%q, Valid=%v\n", ns.String, ns.Valid)
	ns = sql.NullString{String: "hello", Valid: true}
	fmt.Printf("  NullString (set):  Value=%q, Valid=%v\n", ns.String, ns.Valid)

	// ─── sql.Null[T] generic wrapper (Go 1.22+) ───
	// var n sql.Null[int64]
	// row.Scan(&n)
	// if n.Valid { use(n.V) }
	fmt.Println("  Go 1.22+: sql.Null[T] generic wrapper for any type")

	// ─── Scanning into a struct (manual) ───
	// type User struct {
	//     ID    int64
	//     Name  string
	//     Email sql.NullString
	// }
	//
	// func scanUser(row *sql.Row) (User, error) {
	//     var u User
	//     err := row.Scan(&u.ID, &u.Name, &u.Email)
	//     return u, err
	// }
	//
	// For automatic struct scanning: use github.com/jmoiron/sqlx
	// or github.com/blockloop/scan

	// ─── Scanning dynamic columns ───
	// cols, _ := rows.Columns()
	// vals := make([]interface{}, len(cols))
	// ptrs := make([]interface{}, len(cols))
	// for i := range vals {
	//     ptrs[i] = &vals[i]
	// }
	// rows.Scan(ptrs...)
	fmt.Println("  Dynamic columns: rows.Columns() + interface{} slice scan")

	fmt.Println()
}

// =============================================================================
// PART 6: Context-Aware Patterns
// =============================================================================
func contextPatterns() {
	fmt.Println("--- CONTEXT-AWARE PATTERNS ---")

	// ALWAYS use the Context variants:
	//   db.QueryContext(ctx, ...)
	//   db.QueryRowContext(ctx, ...)
	//   db.ExecContext(ctx, ...)
	//   db.BeginTx(ctx, opts)
	//   db.PingContext(ctx)
	//
	// The non-Context versions (db.Query, db.Exec) use context.Background().
	// In production: NEVER use the non-Context versions in HTTP handlers!

	// ─── Query timeout ───
	// ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	// defer cancel()
	// rows, err := db.QueryContext(ctx, "SELECT * FROM big_table")
	// // If query takes >5s: context is canceled, query is killed, err returned
	fmt.Println("  Always use Context variants (QueryContext, ExecContext)")
	fmt.Println("  context.WithTimeout for query deadlines")

	// ─── Request-scoped context ───
	// func handler(w http.ResponseWriter, r *http.Request) {
	//     ctx := r.Context()  // canceled when client disconnects
	//     user, err := db.QueryRowContext(ctx, "SELECT ...", id)
	//     // If client disconnects: query is canceled automatically!
	// }
	fmt.Println("  r.Context() propagates client disconnect to DB queries")

	// ─── Long-running operations ───
	// For background jobs, DON'T use the HTTP request context!
	// Create a separate context with its own timeout:
	// jobCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// defer cancel()
	fmt.Println("  Background jobs: use context.Background() with own timeout")

	fmt.Println()
}

// =============================================================================
// PART 7: Production Patterns
// =============================================================================
func productionPatterns() {
	fmt.Println("--- PRODUCTION PATTERNS ---")

	// ─── PATTERN 1: Repository pattern ───
	// type UserRepository struct {
	//     db *sql.DB
	// }
	//
	// func NewUserRepository(db *sql.DB) *UserRepository {
	//     return &UserRepository{db: db}
	// }
	//
	// func (r *UserRepository) GetByID(ctx context.Context, id int64) (*User, error) {
	//     var u User
	//     err := r.db.QueryRowContext(ctx,
	//         "SELECT id, name, email FROM users WHERE id = $1", id,
	//     ).Scan(&u.ID, &u.Name, &u.Email)
	//     if errors.Is(err, sql.ErrNoRows) {
	//         return nil, ErrNotFound
	//     }
	//     return &u, err
	// }
	fmt.Println("  Pattern 1: Repository wraps sql.DB, translates sql.ErrNoRows")

	// ─── PATTERN 2: Health check ───
	// func (r *UserRepository) Health(ctx context.Context) error {
	//     ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	//     defer cancel()
	//     return r.db.PingContext(ctx)
	// }
	fmt.Println("  Pattern 2: PingContext with 2s timeout for health checks")

	// ─── PATTERN 3: Graceful shutdown ───
	// func main() {
	//     db, _ := sql.Open(...)
	//     defer db.Close()  // waits for in-flight queries, then closes pool
	//     // ... start server ...
	//     // On SIGTERM: server.Shutdown() first, then db.Close()
	// }
	fmt.Println("  Pattern 3: db.Close() on shutdown (waits for in-flight queries)")

	// ─── PATTERN 4: Batch inserts ───
	// For bulk inserts, build a multi-value INSERT:
	//
	// func batchInsert(ctx context.Context, db *sql.DB, users []User) error {
	//     if len(users) == 0 { return nil }
	//
	//     var b strings.Builder
	//     args := make([]interface{}, 0, len(users)*2)
	//     b.WriteString("INSERT INTO users (name, email) VALUES ")
	//
	//     for i, u := range users {
	//         if i > 0 { b.WriteString(", ") }
	//         b.WriteString(fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
	//         args = append(args, u.Name, u.Email)
	//     }
	//
	//     _, err := db.ExecContext(ctx, b.String(), args...)
	//     return err
	// }
	//
	// Or use PostgreSQL COPY for maximum throughput (via pgx directly).
	fmt.Println("  Pattern 4: Multi-value INSERT for batch operations")

	// ─── PATTERN 5: Connection pool monitoring ───
	// Export these as Prometheus metrics:
	//
	// func collectDBMetrics(db *sql.DB) {
	//     ticker := time.NewTicker(10 * time.Second)
	//     for range ticker.C {
	//         stats := db.Stats()
	//         dbOpenConns.Set(float64(stats.OpenConnections))
	//         dbInUse.Set(float64(stats.InUse))
	//         dbIdle.Set(float64(stats.Idle))
	//         dbWaitCount.Add(float64(stats.WaitCount))
	//         dbWaitDuration.Observe(stats.WaitDuration.Seconds())
	//     }
	// }
	fmt.Println("  Pattern 5: Export db.Stats() as Prometheus metrics")

	// ─── PATTERN 6: Migrations ───
	// Use a migration tool — don't manage schema manually!
	//
	// Popular tools:
	//   golang-migrate/migrate: SQL file-based, simple, reliable
	//   pressly/goose: Go code or SQL, good for complex migrations
	//   atlas: declarative + versioned, HCL or SQL
	//
	// Convention: migrations/001_create_users.up.sql
	//             migrations/001_create_users.down.sql
	fmt.Println("  Pattern 6: golang-migrate or goose for schema migrations")

	_ = time.Second          // use time
	_ = context.Background() // use context

	fmt.Println()
}

// =============================================================================
// PART 8: Common Pitfalls
// =============================================================================
func commonPitfalls() {
	fmt.Println("--- COMMON PITFALLS ---")

	// ─── PITFALL 1: Forgetting to close Rows ───
	// rows, err := db.QueryContext(ctx, "SELECT ...")
	// if err != nil { return err }
	// // MISSING: defer rows.Close()
	// // → Connection is NEVER returned to pool → pool exhaustion → deadlock!
	//
	// RULE: ALWAYS defer rows.Close() immediately after checking err.
	fmt.Println("  Pitfall 1: Always defer rows.Close() — leaks connections!")

	// ─── PITFALL 2: Using db.Query for non-SELECT ───
	// db.Query("DELETE FROM users WHERE id = $1", id)
	// → Returns (*Rows, error). If you don't close Rows, connection leaks!
	// → Use db.Exec for INSERT/UPDATE/DELETE (returns Result, no Rows to close)
	fmt.Println("  Pitfall 2: Use Exec for INSERT/UPDATE/DELETE (not Query)")

	// ─── PITFALL 3: SQL injection via string concatenation ───
	// NEVER: db.Query("SELECT * FROM users WHERE id = " + id)
	// ALWAYS: db.Query("SELECT * FROM users WHERE id = $1", id)
	// The driver automatically escapes parameters.
	fmt.Println("  Pitfall 3: NEVER concatenate SQL — use parameterized queries")

	// ─── PITFALL 4: Not setting MaxOpenConns ───
	// Default MaxOpenConns is 0 (unlimited).
	// Under load: Go opens hundreds of connections → database overwhelmed.
	// ALWAYS set MaxOpenConns in production!
	fmt.Println("  Pitfall 4: Set MaxOpenConns (default unlimited → DB overload)")

	// ─── PITFALL 5: Scanning NULL into non-pointer type ───
	// var name string
	// row.Scan(&name)  // ERROR if column is NULL!
	// → Use sql.NullString or *string for nullable columns.
	fmt.Println("  Pitfall 5: Use sql.NullString or *string for nullable columns")

	// ─── PITFALL 6: Using db.Query inside a transaction ───
	// tx.Begin(...)
	// db.Query(...)  // WRONG: uses a different connection!
	// tx.Commit()
	// → Always use tx.QueryContext, tx.ExecContext inside transactions.
	fmt.Println("  Pitfall 6: Inside tx, use tx.QueryContext (not db.QueryContext)")

	// ─── PITFALL 7: N+1 query problem ───
	// // Fetching users, then querying orders for EACH user
	// users, _ := db.Query("SELECT id FROM users")
	// for users.Next() {
	//     var id int
	//     users.Scan(&id)
	//     orders, _ := db.Query("SELECT * FROM orders WHERE user_id = $1", id)
	//     // → N+1 queries! Use JOIN or IN clause instead.
	// }
	//
	// FIX: "SELECT u.*, o.* FROM users u JOIN orders o ON u.id = o.user_id"
	// Or: "SELECT * FROM orders WHERE user_id = ANY($1)" with all IDs
	fmt.Println("  Pitfall 7: N+1 queries — use JOIN or IN/ANY clause")

	// ─── PITFALL 8: Not checking rows.Err() ───
	// for rows.Next() { ... }
	// // MISSING: if err := rows.Err(); err != nil { ... }
	// rows.Next() can stop due to an error (not just EOF).
	// Always check rows.Err() after the loop!
	fmt.Println("  Pitfall 8: Check rows.Err() after iteration loop")

	fmt.Println()
}
