# 1.5 — Normalization & Schema Design

> Normalization is the **science** of eliminating redundancy.
> Schema design is the **art** of modeling reality in tables.
> Masters know both — and know when to break the rules.

---

## 1. Why Normalize?

Without normalization, you get **anomalies**:

```
┌────┬───────┬─────────────┬──────────┬───────────┐
│ id │ name  │ department  │ dept_loc │ course    │
├────┼───────┼─────────────┼──────────┼───────────┤
│  1 │ Alice │ Engineering │ Floor 3  │ Databases │
│  1 │ Alice │ Engineering │ Floor 3  │ Networks  │
│  2 │ Bob   │ Engineering │ Floor 3  │ Databases │
└────┴───────┴─────────────┴──────────┴───────────┘
```

**Update anomaly:** Change Engineering's location → must update EVERY row. Miss one = inconsistency.

**Insert anomaly:** Can't add a new department with no students (no id to use as key).

**Delete anomaly:** Delete all students from a department → lose the department's location data.

Normalization splits this into separate tables to eliminate these problems.

---

## 2. The Normal Forms

### 1NF — First Normal Form

**Rule:** All attribute values must be **atomic** (no repeating groups, no arrays, no nested tables).

```
VIOLATES 1NF:
┌────┬───────┬──────────────────────┐
│ id │ name  │ phones               │
├────┼───────┼──────────────────────┤
│  1 │ Alice │ 555-1234, 555-5678   │  ← multiple values in one cell
└────┴───────┴──────────────────────┘

1NF SOLUTION:
┌────┬───────┬──────────┐     ┌────┬──────────┐
│ id │ name  │          │     │ id │ phone    │
├────┼───────┤          │     ├────┼──────────┤
│  1 │ Alice │          │     │  1 │ 555-1234 │
└────┴───────┘          │     │  1 │ 555-5678 │
                              └────┴──────────┘
```

**Reality check:** PostgreSQL has `TEXT[]` arrays and `JSONB`. Does that violate 1NF? Technically yes in Codd's original definition. Practically, it's fine for tags, metadata, etc. The question is: do you need to JOIN on or independently query those values? If yes → separate table.

### 2NF — Second Normal Form

**Prerequisites:** Must be in 1NF.
**Rule:** No **partial dependencies** — every non-key attribute must depend on the **entire** primary key, not just part of it.

Only relevant when you have a **composite primary key**.

```
VIOLATES 2NF:
Table: StudentCourse(student_id, course_id, student_name, grade)
PK: (student_id, course_id)

FDs:
  (student_id, course_id) → grade          ✓ depends on full PK
  student_id → student_name                 ✗ depends on PART of PK (partial dependency!)

2NF SOLUTION: Split into two tables
  Student(student_id, student_name)         -- student_id → student_name
  Enrollment(student_id, course_id, grade)  -- (student_id, course_id) → grade
```

### 3NF — Third Normal Form

**Prerequisites:** Must be in 2NF.
**Rule:** No **transitive dependencies** — no non-key attribute depends on another non-key attribute.

**"Every non-key attribute must depend on the key, the whole key, and nothing but the key, so help me Codd."**

```
VIOLATES 3NF:
Employee(id, name, dept_id, dept_name, dept_location)
PK: id

FDs:
  id → dept_id             ✓ (depends on key)
  id → dept_name           but via dept_id → dept_name (transitive!)
  dept_id → dept_name      ✗ non-key → non-key (transitive dependency)
  dept_id → dept_location  ✗ non-key → non-key

3NF SOLUTION:
  Employee(id, name, dept_id)
  Department(dept_id, dept_name, dept_location)
```

### BCNF — Boyce-Codd Normal Form

**Prerequisites:** Must be in 3NF.
**Rule:** For every non-trivial FD X → Y, X must be a **superkey**.

BCNF is stricter than 3NF. They differ only when:
- A table has multiple overlapping candidate keys

```
VIOLATES BCNF (but is in 3NF):
StudentAdvisor(student_id, subject, advisor)
Candidate keys: (student_id, subject)

FDs:
  advisor → subject    ← advisor determines subject
                         but advisor is NOT a superkey!

Example data:
  (Alice, Databases, Prof. Smith)
  (Bob, Databases, Prof. Smith)
  (Alice, Networks, Prof. Jones)

Problem: If Prof. Smith switches from Databases to AI, must update multiple rows.

BCNF SOLUTION:
  Advisors(advisor, subject)              -- advisor → subject
  StudentAdvisors(student_id, advisor)    -- student chooses advisor
```

**Trade-off:** BCNF decomposition can make it impossible to enforce certain dependencies without joins. 3NF is always **dependency-preserving**; BCNF is not.

### 4NF — Fourth Normal Form

**Rule:** No **multivalued dependencies** (MVDs) other than those implied by candidate keys.

An MVD X →→ Y means: for a given X, the set of Y values is independent of other attributes.

```
VIOLATES 4NF:
PersonSkillLanguage(person, skill, language)

If a person's skills and languages are independent:
  (Alice, Python, English)
  (Alice, Python, Spanish)
  (Alice, Java, English)
  (Alice, Java, Spanish)    ← must have ALL combinations! Redundancy.

MVDs: person →→ skill, person →→ language

4NF SOLUTION:
  PersonSkill(person, skill)        -- (Alice, Python), (Alice, Java)
  PersonLanguage(person, language)   -- (Alice, English), (Alice, Spanish)
```

### 5NF (PJNF) — Fifth Normal Form (Project-Join Normal Form)

**Rule:** No **join dependencies** that aren't implied by candidate keys.

A relation R has a join dependency *(R1, R2, ..., Rn) if R can be losslessly decomposed into R1, R2, ..., Rn.

5NF handles cases where three or more attributes have complex interrelationships.

```
VIOLATES 5NF:
SupplierPartProject(supplier, part, project)

The relationship is: a supplier supplies a part to a project, BUT only if:
  - The supplier supplies that part (to some project)
  - The supplier supplies to that project (some part)
  - That part is used in that project (by some supplier)

This three-way constraint means we need ALL three two-way tables:
  SupplierPart(supplier, part)
  SupplierProject(supplier, project)
  PartProject(part, project)

And the original table = JOIN of all three.
```

5NF is rare in practice. If you encounter it, you'll know.

### 6NF — Sixth Normal Form

**Rule:** A relation is in 6NF if every non-trivial join dependency is trivial (every row has the primary key + at most one non-key attribute).

```
6NF (used in temporal databases):
  Employee_Name(emp_id, valid_from, valid_to, name)
  Employee_Salary(emp_id, valid_from, valid_to, salary)
  Employee_Dept(emp_id, valid_from, valid_to, dept_id)

Each attribute changes independently with its own time range.
```

6NF is mainly theoretical — used in data vault modeling and temporal databases.

---

## 3. Normalization Summary

```
  1NF: Atomic values             (no arrays, no repeating groups)
  2NF: No partial dependencies   (attributes depend on WHOLE key)
  3NF: No transitive dependencies (non-key attributes don't depend on each other)
 BCNF: Every determinant is a superkey
  4NF: No multivalued dependencies
  5NF: No join dependencies
  6NF: One attribute per relation (temporal use)
```

**In practice:** Most databases are designed to 3NF or BCNF. Going further is rarely needed.

---

## 4. Denormalization — When to Break the Rules

Normalization minimizes redundancy but maximizes JOINs. At scale, JOINs are expensive.

### When to Denormalize

| Scenario | Denormalization Technique |
|----------|--------------------------|
| Read-heavy workload | Pre-join tables, materialized views |
| Frequent aggregation | Store pre-computed totals |
| Too many JOINs per query | Embed related data |
| Distributed databases | Denormalize to avoid cross-node joins |
| Reporting/analytics | Star schema (inherently denormalized) |

### Example: Order with Customer Name

```sql
-- Normalized: join every time
SELECT o.id, c.name, o.total
FROM orders o JOIN customers c ON o.customer_id = c.id;

-- Denormalized: store customer_name in orders
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER REFERENCES customers(id),
    customer_name TEXT,      -- ← denormalized!
    total DECIMAL(10,2)
);
-- Trade-off: faster reads, but if customer changes name, orders table is stale
-- Solution: update on change, or accept staleness (it's the name AT TIME OF ORDER anyway)
```

---

## 5. Entity-Relationship (ER) Modeling

### Core Concepts

```
ENTITY         = a thing (noun): Student, Course, Department
ATTRIBUTE      = a property: name, age, salary
RELATIONSHIP   = association between entities: enrolls_in, works_for

ENTITY TYPES:
  Strong entity    — exists independently, has its own PK
  Weak entity      — depends on a strong entity (e.g., Room depends on Building)

RELATIONSHIP CARDINALITIES:
  1:1   — One person has one passport
  1:N   — One department has many employees
  M:N   — Many students enroll in many courses (needs junction table)
```

### ER to Tables

```
1:1 relationship → FK in either table (or merge into one table)
  Person(id, name)
  Passport(id, person_id UNIQUE, number)    -- UNIQUE FK enforces 1:1

1:N relationship → FK in the "many" side
  Department(id, name)
  Employee(id, name, dept_id REFERENCES department(id))

M:N relationship → Junction table
  Student(id, name)
  Course(id, title)
  Enrollment(student_id, course_id, grade, PRIMARY KEY(student_id, course_id))

Multivalued attribute → Separate table
  Employee(id, name)
  EmployeePhone(employee_id, phone, PRIMARY KEY(employee_id, phone))

Derived attribute → Computed column or view (don't store)
  age = current_date - birth_date (don't store age!)

Weak entity → Composite PK including owner's PK
  Building(building_id, name)
  Room(building_id, room_number, capacity, PRIMARY KEY(building_id, room_number))
  -- Room doesn't exist without a Building
```

### Common Design Patterns

#### Pattern 1: Polymorphic Associations

"A comment can belong to a Post, Photo, or Video"

```sql
-- APPROACH 1: Separate FK columns (simple, but sparse NULLs)
CREATE TABLE comment (
    id SERIAL PRIMARY KEY,
    body TEXT NOT NULL,
    post_id INTEGER REFERENCES post(id),
    photo_id INTEGER REFERENCES photo(id),
    video_id INTEGER REFERENCES video(id),
    CHECK (
        (post_id IS NOT NULL)::INT +
        (photo_id IS NOT NULL)::INT +
        (video_id IS NOT NULL)::INT = 1  -- exactly one must be set
    )
);

-- APPROACH 2: Generic FK (flexible, but no referential integrity)
CREATE TABLE comment (
    id SERIAL PRIMARY KEY,
    body TEXT NOT NULL,
    commentable_type TEXT NOT NULL,  -- 'post', 'photo', 'video'
    commentable_id INTEGER NOT NULL
    -- Can't create a FOREIGN KEY because target table varies
);
CREATE INDEX idx_comment_target ON comment(commentable_type, commentable_id);

-- APPROACH 3: Separate tables (most normalized, more JOINs)
CREATE TABLE post_comment (id SERIAL PRIMARY KEY, post_id REFERENCES post(id), body TEXT);
CREATE TABLE photo_comment (id SERIAL PRIMARY KEY, photo_id REFERENCES photo(id), body TEXT);
```

#### Pattern 2: Self-Referential (Trees/Hierarchies)

```sql
-- Adjacency list (simplest)
CREATE TABLE category (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    parent_id INTEGER REFERENCES category(id)
);
-- Pro: simple, easy to update
-- Con: recursive queries needed for tree traversal

-- Materialized path
CREATE TABLE category (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    path TEXT NOT NULL  -- '/1/5/12/45'
);
CREATE INDEX idx_category_path ON category USING BTREE (path text_pattern_ops);
-- Find all descendants: WHERE path LIKE '/1/5/%'
-- Pro: fast subtree queries
-- Con: moving a subtree requires updating all descendant paths

-- Nested sets
CREATE TABLE category (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    lft INTEGER NOT NULL,
    rgt INTEGER NOT NULL
);
-- Pro: fast subtree queries (WHERE lft BETWEEN parent.lft AND parent.rgt)
-- Con: inserts/deletes require renumbering

-- Closure table (best for complex operations)
CREATE TABLE category (id SERIAL PRIMARY KEY, name TEXT NOT NULL);
CREATE TABLE category_closure (
    ancestor_id INTEGER REFERENCES category(id),
    descendant_id INTEGER REFERENCES category(id),
    depth INTEGER NOT NULL,
    PRIMARY KEY (ancestor_id, descendant_id)
);
-- Every node has a row to itself (depth=0) and to every ancestor
-- Pro: fast for any tree query (ancestors, descendants, depth)
-- Con: more storage, must maintain closure rows on insert/delete
```

#### Pattern 3: Temporal Data (History Tracking)

```sql
-- Approach 1: Slowly Changing Dimension Type 2 (SCD2)
CREATE TABLE employee_history (
    id SERIAL PRIMARY KEY,
    employee_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    salary DECIMAL(10,2) NOT NULL,
    dept_id INTEGER,
    valid_from TIMESTAMPTZ NOT NULL,
    valid_to TIMESTAMPTZ,  -- NULL = current
    is_current BOOLEAN DEFAULT TRUE
);

-- Current record:
SELECT * FROM employee_history WHERE employee_id = 1 AND is_current = TRUE;

-- Record at a point in time:
SELECT * FROM employee_history
WHERE employee_id = 1
  AND valid_from <= '2025-06-01'
  AND (valid_to IS NULL OR valid_to > '2025-06-01');

-- PostgreSQL temporal tables (system versioning)
-- Not yet native in PG, but temporal_tables extension exists
-- SQL:2011 defines this natively (supported in MariaDB, SQL Server)
```

#### Pattern 4: Tagging / Many-to-Many with Attributes

```sql
-- Simple M:N
CREATE TABLE post_tag (
    post_id INTEGER REFERENCES post(id) ON DELETE CASCADE,
    tag_id INTEGER REFERENCES tag(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, tag_id)
);

-- M:N with attributes
CREATE TABLE enrollment (
    student_id INTEGER REFERENCES student(id),
    course_id INTEGER REFERENCES course(id),
    grade CHAR(2),
    enrolled_at DATE DEFAULT CURRENT_DATE,
    PRIMARY KEY (student_id, course_id)
);
```

#### Pattern 5: Multi-Tenant Design

```sql
-- APPROACH 1: Shared table with tenant_id (most common)
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL,
    total DECIMAL(10,2),
    created_at TIMESTAMPTZ
);
CREATE INDEX idx_orders_tenant ON orders(tenant_id);
-- EVERY query must include WHERE tenant_id = ?
-- Row-Level Security (RLS) can enforce this:
ALTER TABLE orders ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON orders
    USING (tenant_id = current_setting('app.tenant_id')::INT);

-- APPROACH 2: Schema per tenant
CREATE SCHEMA tenant_acme;
CREATE TABLE tenant_acme.orders (...);
-- Pro: true isolation, easy to drop a tenant
-- Con: schema management complexity, migration on N schemas

-- APPROACH 3: Database per tenant
-- Maximum isolation, most operational overhead
```

---

## 6. Star Schema and Data Warehouse Design

For analytics/OLAP workloads, normalized schemas are the WRONG choice.

### Star Schema

```
                    ┌──────────────┐
                    │  dim_date    │
                    │──────────────│
                    │ date_key (PK)│
                    │ date         │
                    │ year         │
                    │ quarter      │
                    │ month        │
                    │ day_of_week  │
                    │ is_weekend   │
                    └──────┬───────┘
                           │
┌──────────────┐   ┌───────┴────────┐   ┌──────────────┐
│ dim_product  │   │  fact_sales    │   │ dim_customer │
│──────────────│   │────────────────│   │──────────────│
│ product_key  │──→│ date_key (FK)  │←──│ customer_key │
│ name         │   │ product_key(FK)│   │ name         │
│ category     │   │ customer_key(FK│   │ email        │
│ brand        │   │ store_key (FK) │   │ city         │
│ price        │   │ quantity       │   │ segment      │
└──────────────┘   │ revenue        │   └──────────────┘
                   │ discount       │
                   │ cost           │   ┌──────────────┐
                   └───────┬────────┘   │ dim_store    │
                           │            │──────────────│
                           └───────────→│ store_key    │
                                        │ name         │
                                        │ city, state  │
                                        │ region       │
                                        └──────────────┘
```

```sql
-- Fact table: measures (numbers you aggregate)
CREATE TABLE fact_sales (
    date_key INTEGER REFERENCES dim_date(date_key),
    product_key INTEGER REFERENCES dim_product(product_key),
    customer_key INTEGER REFERENCES dim_customer(customer_key),
    store_key INTEGER REFERENCES dim_store(store_key),
    quantity INTEGER NOT NULL,
    revenue DECIMAL(12,2) NOT NULL,
    discount DECIMAL(10,2) DEFAULT 0,
    cost DECIMAL(12,2) NOT NULL
);

-- Dimension table: descriptive attributes (things you filter/group by)
CREATE TABLE dim_date (
    date_key INTEGER PRIMARY KEY,  -- surrogate key: 20260417
    full_date DATE NOT NULL,
    year SMALLINT NOT NULL,
    quarter SMALLINT NOT NULL,
    month SMALLINT NOT NULL,
    month_name VARCHAR(20) NOT NULL,
    day_of_week SMALLINT NOT NULL,
    day_name VARCHAR(20) NOT NULL,
    is_weekend BOOLEAN NOT NULL,
    is_holiday BOOLEAN DEFAULT FALSE
);

-- Analytics query on star schema (simple, fast):
SELECT
    d.year,
    d.quarter,
    p.category,
    s.region,
    SUM(f.revenue) AS total_revenue,
    SUM(f.quantity) AS total_units,
    AVG(f.revenue / NULLIF(f.quantity, 0)) AS avg_price
FROM fact_sales f
JOIN dim_date d ON f.date_key = d.date_key
JOIN dim_product p ON f.product_key = p.product_key
JOIN dim_store s ON f.store_key = s.store_key
WHERE d.year = 2026
GROUP BY d.year, d.quarter, p.category, s.region
ORDER BY total_revenue DESC;
```

### Slowly Changing Dimensions (SCD)

When dimension data changes (customer moves cities, product is reclassified):

| Type | Strategy | Example |
|------|----------|---------|
| **SCD1** | Overwrite | Just UPDATE the old value. Lose history. |
| **SCD2** | New row | Add new row with valid_from/valid_to. Keep history. |
| **SCD3** | New column | Add previous_value column. Limited history (only one change). |
| **SCD4** | Separate history table | Current in main table, all history in separate table. |
| **SCD6** | Hybrid 1+2+3 | SCD2 with current value columns for convenience. |

```sql
-- SCD2 example:
CREATE TABLE dim_customer (
    customer_key SERIAL PRIMARY KEY,   -- surrogate key (NEW for each version)
    customer_id INTEGER NOT NULL,       -- natural/business key
    name TEXT NOT NULL,
    city TEXT NOT NULL,
    valid_from DATE NOT NULL,
    valid_to DATE,                      -- NULL = current
    is_current BOOLEAN DEFAULT TRUE
);

-- Customer moves from NYC to LA:
-- 1. Close old record
UPDATE dim_customer SET valid_to = CURRENT_DATE, is_current = FALSE
WHERE customer_id = 42 AND is_current = TRUE;

-- 2. Insert new version
INSERT INTO dim_customer (customer_id, name, city, valid_from, is_current)
VALUES (42, 'Alice Chen', 'Los Angeles', CURRENT_DATE, TRUE);

-- Fact table references customer_KEY (not customer_id!)
-- So old sales still point to the NYC version, new sales point to LA version.
```

---

## 7. Schema Design Checklist

```
□ Every table has a primary key
□ Use surrogate keys (SERIAL/BIGSERIAL/UUID) for PKs, keep natural keys as UNIQUE
□ All foreign keys are explicitly declared
□ Use appropriate data types (no storing numbers as strings!)
□ Use NOT NULL unless NULL has genuine meaning
□ Use CHECK constraints for domain validation
□ Timestamps always use TIMESTAMPTZ
□ Table and column names are snake_case, singular (employee not employees)
□ Junction tables named: entity1_entity2 or with a meaningful verb (enrollment, subscription)
□ Indexes on all foreign keys
□ Indexes on columns used in WHERE and ORDER BY
□ No redundant indexes (index on (A, B) covers queries on just (A))
□ Consider partial indexes for hot queries with fixed predicates
□ created_at and updated_at on every entity table
□ Soft delete (deleted_at) vs hard delete — decide per table
□ At least 3NF for OLTP, star schema for OLAP
```

---

## 8. Practice Exercises

**E1.** Given this unnormalized table, normalize to 3NF:
```
Invoice(invoice_id, customer_name, customer_email, customer_city,
        item1_name, item1_qty, item1_price,
        item2_name, item2_qty, item2_price,
        invoice_date, total)
```

<details><summary>Solution</summary>

```sql
-- Step 1: 1NF — eliminate repeating groups
-- Step 2: Separate entities

CREATE TABLE customer (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    city TEXT
);

CREATE TABLE invoice (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER REFERENCES customer(id),
    invoice_date DATE NOT NULL
    -- total is derived (SUM of line items), don't store
);

CREATE TABLE invoice_line (
    id SERIAL PRIMARY KEY,
    invoice_id INTEGER REFERENCES invoice(id),
    item_name TEXT NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(10,2) NOT NULL CHECK (unit_price >= 0)
);

-- Total can be computed:
-- SELECT SUM(quantity * unit_price) FROM invoice_line WHERE invoice_id = ?
-- Or stored as a denormalized column if performance requires it.
```
</details>

**E2.** Design a schema for a blog platform with: users, posts, comments (nested), tags, likes (on posts and comments).

<details><summary>Solution</summary>

```sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE post (
    id BIGSERIAL PRIMARY KEY,
    author_id BIGINT NOT NULL REFERENCES users(id),
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE tag (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL
);

CREATE TABLE post_tag (
    post_id BIGINT REFERENCES post(id) ON DELETE CASCADE,
    tag_id INTEGER REFERENCES tag(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, tag_id)
);

-- Nested comments using adjacency list
CREATE TABLE comment (
    id BIGSERIAL PRIMARY KEY,
    post_id BIGINT NOT NULL REFERENCES post(id) ON DELETE CASCADE,
    author_id BIGINT NOT NULL REFERENCES users(id),
    parent_comment_id BIGINT REFERENCES comment(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Polymorphic likes using separate tables (preserves FK integrity)
CREATE TABLE post_like (
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    post_id BIGINT REFERENCES post(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, post_id)
);

CREATE TABLE comment_like (
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    comment_id BIGINT REFERENCES comment(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, comment_id)
);

-- Key indexes
CREATE INDEX idx_post_author ON post(author_id);
CREATE INDEX idx_post_published ON post(published_at DESC) WHERE published_at IS NOT NULL;
CREATE INDEX idx_comment_post ON comment(post_id);
CREATE INDEX idx_comment_parent ON comment(parent_comment_id);
```
</details>

**E3.** Design a star schema for an e-commerce analytics dashboard tracking: sales by product, category, customer segment, geography, and time.

<details><summary>Solution</summary>

```sql
CREATE TABLE dim_date (
    date_key INTEGER PRIMARY KEY,
    full_date DATE UNIQUE NOT NULL,
    year SMALLINT, quarter SMALLINT, month SMALLINT,
    week SMALLINT, day_of_week SMALLINT, day_name TEXT,
    is_weekend BOOLEAN, is_holiday BOOLEAN
);

CREATE TABLE dim_product (
    product_key SERIAL PRIMARY KEY,
    product_id TEXT NOT NULL,
    name TEXT NOT NULL,
    category TEXT, subcategory TEXT, brand TEXT,
    valid_from DATE, valid_to DATE, is_current BOOLEAN  -- SCD2
);

CREATE TABLE dim_customer (
    customer_key SERIAL PRIMARY KEY,
    customer_id TEXT NOT NULL,
    name TEXT, email TEXT, segment TEXT,
    city TEXT, state TEXT, country TEXT,
    valid_from DATE, valid_to DATE, is_current BOOLEAN  -- SCD2
);

CREATE TABLE dim_geography (
    geo_key SERIAL PRIMARY KEY,
    city TEXT, state TEXT, country TEXT, region TEXT,
    latitude DECIMAL(9,6), longitude DECIMAL(9,6)
);

CREATE TABLE fact_sales (
    date_key INTEGER REFERENCES dim_date(date_key),
    product_key INTEGER REFERENCES dim_product(product_key),
    customer_key INTEGER REFERENCES dim_customer(customer_key),
    geo_key INTEGER REFERENCES dim_geography(geo_key),
    order_id BIGINT NOT NULL,
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    discount DECIMAL(10,2) DEFAULT 0,
    revenue DECIMAL(12,2) NOT NULL,
    cost DECIMAL(12,2) NOT NULL,
    profit DECIMAL(12,2) NOT NULL
);

-- Bitmap indexes ideal for fact table FKs (in warehouses)
CREATE INDEX idx_fact_date ON fact_sales(date_key);
CREATE INDEX idx_fact_product ON fact_sales(product_key);
CREATE INDEX idx_fact_customer ON fact_sales(customer_key);
CREATE INDEX idx_fact_geo ON fact_sales(geo_key);
```
</details>

---

## Key Takeaways

1. **Normalize to 3NF/BCNF for OLTP** — it prevents anomalies and maintains data integrity.
2. **Denormalize for OLAP** — star schemas with fact and dimension tables are optimized for analytics.
3. **The right choice depends on workload**: reads vs writes, simplicity vs performance, consistency vs availability.
4. **ER modeling** is the bridge between business requirements and table design.
5. **Surrogate keys** (auto-increment/UUID) decouple your schema from business identifiers.
6. **SCD Type 2** is the standard for tracking dimension history in warehouses.
7. **Design patterns** (polymorphic, self-referential, temporal, multi-tenant) each have trade-offs — there's no one-size-fits-all.

---

**Phase 1 Complete!** You now understand:
- The mathematical foundation (relational model)
- SQL from beginner to advanced (window functions, CTEs, recursive queries, EXPLAIN)
- Normalization theory (1NF through 6NF)
- Schema design for OLTP and OLAP
- Common design patterns for real-world problems

Next: [Phase 2 — Database Internals](../phase2/01-storage-engines.md) →
