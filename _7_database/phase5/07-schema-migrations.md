# 5.7 — Schema Migration & Evolution

> Every application evolves. So does its schema.  
> The question isn't WHETHER you'll change your schema,  
> but whether you can do it WITHOUT downtime.

---

## 1. Migration Tools

```
Tool         Language    Format    State tracking
─────────── ─────────── ───────── ─────────────────
Flyway       Java/CLI    SQL/Java  Version table
Liquibase    Java/CLI    XML/YAML  DATABASECHANGELOG
Alembic      Python      Python    alembic_version table
golang-migrate Go/CLI    SQL       schema_migrations table
Atlas        Go/CLI      HCL/SQL   atlas_schema_revisions
Sqitch       Perl/CLI    SQL       sqitch.changes
Prisma       TypeScript  Prisma    _prisma_migrations
Knex         JavaScript  JS        knex_migrations

# Common pattern:
migrations/
  001_create_users.sql
  002_add_email_index.sql
  003_add_orders_table.sql
  004_add_status_to_orders.sql

# Each migration runs once, in order, tracked in a metadata table.
# Rollback migrations: paired down migrations (not always reliable).
# Prefer forward-only migrations with expand-and-contract.
```

---

## 2. Zero-Downtime Migration Patterns

### Safe Operations (No Lock Issues)

```sql
-- Adding a column (without default — instant in PG 11+):
ALTER TABLE orders ADD COLUMN notes TEXT;
-- PG 11+: no table rewrite, no lock (just metadata change)

-- Adding a column WITH volatile default (PG 11+):
ALTER TABLE orders ADD COLUMN status TEXT DEFAULT 'pending';
-- PG 11+ stores default in catalog → instant, no rewrite
-- PG < 11: REWRITES ENTIRE TABLE → downtime!

-- Creating an index CONCURRENTLY:
CREATE INDEX CONCURRENTLY idx_orders_status ON orders (status);
-- Does NOT lock the table for writes
-- Takes longer but doesn't block production traffic
-- If it fails: DROP INDEX CONCURRENTLY idx_orders_status; and retry

-- Adding a CHECK constraint (NOT VALID first):
ALTER TABLE orders ADD CONSTRAINT chk_positive_total CHECK (total > 0) NOT VALID;
-- NOT VALID: skips checking existing rows → instant, no scan
-- Then validate in background:
ALTER TABLE orders VALIDATE CONSTRAINT chk_positive_total;
-- Validates existing rows with Share Update Exclusive lock (doesn't block writes)
```

### Dangerous Operations (Need Special Handling)

```sql
-- ✗ ADDING a NOT NULL constraint (scans entire table, brief lock):
-- Safe pattern:
ALTER TABLE orders ADD CONSTRAINT chk_nn CHECK (email IS NOT NULL) NOT VALID;
ALTER TABLE orders VALIDATE CONSTRAINT chk_nn;
-- Then: ALTER TABLE orders ALTER COLUMN email SET NOT NULL;
-- PG 12+: if a CHECK constraint proves NOT NULL, SET NOT NULL is instant

-- ✗ CHANGING column type (rewrites table):
-- Instead of: ALTER TABLE orders ALTER COLUMN amount TYPE NUMERIC(12,2);
-- Do:
--   1. Add new column: ALTER TABLE orders ADD COLUMN amount_new NUMERIC(12,2);
--   2. Backfill:       UPDATE orders SET amount_new = amount WHERE amount_new IS NULL;
--                       (do in batches to avoid long transactions)
--   3. Application: write to BOTH columns
--   4. Verify data matches
--   5. Switch reads to new column
--   6. Drop old column (future migration)

-- ✗ RENAMING a column:
-- Never rename directly (breaks running application code)
-- Instead: add new column → copy data → switch app → drop old column
-- Or: use a VIEW as an adapter

-- ✗ DROPPING a column:
-- Application must stop using the column FIRST
-- Then: ALTER TABLE orders DROP COLUMN old_status;
-- PostgreSQL: DROP COLUMN is instant (marks column as dropped, no rewrite)
-- But accessing table before vacuum rewrite still sees old data

-- ✗ ADDING a FOREIGN KEY:
ALTER TABLE orders ADD CONSTRAINT fk_customer 
    FOREIGN KEY (customer_id) REFERENCES customers(id) NOT VALID;
ALTER TABLE orders VALIDATE CONSTRAINT fk_customer;
-- NOT VALID → instant add. VALIDATE → scans but doesn't block writes.
```

### The Expand-and-Contract Pattern

```
The universal pattern for safe schema migration:

Phase 1 — EXPAND (backward compatible):
  Add new column/table/index
  Application writes to BOTH old and new
  Backfill new column with existing data

Phase 2 — MIGRATE (switch reads):
  Application reads from new, writes to both
  Verify data consistency

Phase 3 — CONTRACT (clean up):
  Application stops writing to old
  Drop old column/table/index

Each phase is a separate deployment.
At any point, you can roll back to the previous phase.

Example: rename column "username" to "display_name":
  Phase 1: ADD COLUMN display_name; trigger copies username→display_name
  Phase 2: App reads from display_name, writes to both
  Phase 3: App only uses display_name
  Phase 4: DROP COLUMN username (next release)
```

---

## 3. Online DDL Tools

```bash
# pg_repack (PostgreSQL — no VACUUM FULL locking):
pg_repack -d mydb -t bloated_table
# Rewrites table in background, then atomic swap
# Doesn't block reads or writes (uses triggers to capture changes)

# gh-ost (MySQL — GitHub Online Schema Change):
gh-ost --database=mydb --table=orders --alter="ADD COLUMN notes TEXT" --execute
# Creates ghost table, copies data, captures binlog changes
# Atomic rename at the end
# Throttles based on replica lag

# pt-online-schema-change (Percona Toolkit — MySQL):
pt-online-schema-change --alter="ADD COLUMN notes TEXT" D=mydb,t=orders --execute
# Similar approach: triggers + copy + rename
```

---

## Key Takeaways

1. **CONCURRENTLY is your friend** in PostgreSQL. Use `CREATE INDEX CONCURRENTLY` and `NOT VALID` constraints to avoid locks.
2. **Expand-and-contract** is the universal zero-downtime migration pattern. Never break backward compatibility in a single step.
3. **Column type changes are the most dangerous.** They rewrite the entire table. Use the add-copy-switch-drop pattern.
4. **Forward-only migrations.** Don't rely on rollback scripts. Instead, make each migration self-contained and backward compatible.
5. **Batch your backfills.** `UPDATE ... WHERE id BETWEEN x AND y` in loops of 10K-100K rows. Never one massive UPDATE.

---

Next: [08-data-warehousing.md](08-data-warehousing.md) →
