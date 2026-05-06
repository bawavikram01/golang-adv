# 5.8 — Data Warehousing & OLAP

> OLTP asks: "What is the balance of account #12345?"  
> OLAP asks: "What was the average revenue per customer by region over the last 3 years?"  
> Different questions demand fundamentally different architectures.

---

## 1. OLTP vs. OLAP

```
Characteristic          OLTP                    OLAP
─────────────────────── ─────────────────────── ──────────────────────
Query pattern           Point reads, small txns Full-table scans, aggregations
Latency                 Milliseconds            Seconds to minutes
Data volume             GBs – low TBs           TBs – PBs
Schema                  Normalized (3NF)        Denormalized (star/snowflake)
Concurrency             Thousands of users       Dozens of analysts
Updates                 Frequent inserts/updates Append-mostly, batch loads
Storage format          Row-oriented            Column-oriented
Index usage             Heavy B-tree usage      Minimal (scan-focused)
Typical system          PostgreSQL, MySQL       ClickHouse, BigQuery, Redshift
```

---

## 2. Dimensional Modeling

### Star Schema

```
The backbone of every data warehouse.

                    ┌──────────────┐
                    │ dim_customer │
                    │──────────────│
                    │ customer_id  │─┐
                    │ name         │ │
                    │ segment      │ │
                    │ region       │ │
                    └──────────────┘ │
                                     │
┌──────────────┐   ┌───────────────┐│   ┌──────────────┐
│ dim_product  │   │  fact_sales   ││   │  dim_date    │
│──────────────│   │───────────────││   │──────────────│
│ product_id   │─┐ │ sale_id       ││   │ date_id      │─┐
│ name         │ │ │ date_id       │├───│ full_date    │ │
│ category     │ │ │ customer_id   │┘   │ year         │ │
│ brand        │ └─│ product_id    │    │ quarter      │ │
└──────────────┘   │ store_id      │──┐ │ month        │ │
                   │ quantity       │  │ │ day_of_week  │ │
                   │ unit_price     │  │ └──────────────┘ │
                   │ total_amount   │  │                   │
                   │ discount       │  │ ┌──────────────┐  │
                   └───────────────┘  │ │  dim_store   │  │
                                      │ │──────────────│  │
                                      └─│ store_id     │  │
                                        │ city         │  │
                                        │ state        │  │
                                        │ country      │  │
                                        └──────────────┘
Rules:
  - Fact tables hold measurable events (metrics/amounts)
  - Dimension tables describe context (who, what, where, when)
  - Fact tables have FKs to every relevant dimension
  - Star = one level of dimension tables (denormalized dimensions)
  - Snowflake = dimensions normalized into subdimensions (more joins)
```

### Slowly Changing Dimensions (SCD)

```sql
-- Type 1: Overwrite (lose history)
UPDATE dim_customer SET address = '456 Oak St' WHERE customer_id = 100;

-- Type 2: Add new row (preserve history) — most common
-- customer_id=100 moves from "SF" to "NYC":
INSERT INTO dim_customer 
  (customer_sk, customer_id, name, city, valid_from, valid_to, is_current) 
VALUES
  (nextval('dim_customer_sk_seq'), 100, 'Alice', 'NYC', '2024-01-15', '9999-12-31', TRUE);

UPDATE dim_customer 
SET valid_to = '2024-01-14', is_current = FALSE 
WHERE customer_id = 100 AND is_current = TRUE AND city = 'SF';

-- Type 3: Add column (limited history — previous value only)
ALTER TABLE dim_customer ADD COLUMN previous_city TEXT;
UPDATE dim_customer SET previous_city = city, city = 'NYC' WHERE customer_id = 100;

-- Type 6: Hybrid (1+2+3) — most complex, rarely used
```

### Kimball vs. Inmon vs. Data Vault

```
Approach       Kimball               Inmon                Data Vault
────────────── ────────────────────── ──────────────────── ────────────────────
Philosophy     Bottom-up              Top-down             Agile, auditable
Core structure Dimensional (star)     ER-normalized (3NF)  Hub/Link/Satellite
Build order    Build data marts first Build enterprise DW   Build incrementally
ETL target     Conformed dimensions   Central staging area Raw vault → bus. vault
Ease of query  Easy (star schema)     Complex (joins)      Complex (many joins)
Auditability   Moderate               High                 Highest
Adoption       Most common            Large enterprises     Growing (regulated)
```

---

## 3. Column-Oriented Storage

```
Why columnar for analytics?

ROW STORE (traditional):
  Row 1: [Alice, 35, NYC, $50K]
  Row 2: [Bob,   42, LA,  $75K]
  Row 3: [Carol, 28, SF,  $60K]
  
  To compute AVG(salary):
  → Read ALL columns of ALL rows (even name, age, city)
  → Bad I/O for aggregations

COLUMN STORE:
  name_column:   [Alice, Bob, Carol]
  age_column:    [35, 42, 28]
  city_column:   [NYC, LA, SF]
  salary_column: [$50K, $75K, $60K]
  
  To compute AVG(salary):
  → Read ONLY salary_column
  → Massive I/O savings (read 1/4 of data)
  → Same-type values compress extremely well (10x-100x)

Compression techniques in columnar stores:
  - Run-length encoding: [NYC, NYC, NYC, NYC, LA, LA] → [(NYC,4), (LA,2)]
  - Dictionary encoding: replace strings with integer codes
  - Delta encoding: store differences (timestamps → deltas)
  - Bit-packing: small integers in fewer bits
  - LZ4/ZSTD: general purpose on top of the above
```

---

## 4. Modern OLAP Engines

### ClickHouse

```sql
-- Columnar, vectorized, blazing fast for aggregations
-- MergeTree engine family is the foundation

CREATE TABLE events (
    event_date   Date,
    user_id      UInt64,
    event_type   LowCardinality(String),  -- dictionary encoded
    revenue      Decimal(10, 2),
    properties   String                    -- JSON stored as string
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(event_date)     -- monthly partitions
ORDER BY (event_type, user_id, event_date)  -- primary key (sparse index)
TTL event_date + INTERVAL 2 YEAR;     -- auto-delete after 2 years

-- Materialized views for pre-aggregation:
CREATE MATERIALIZED VIEW daily_revenue
ENGINE = SummingMergeTree()
ORDER BY (event_date, event_type)
AS SELECT
    event_date,
    event_type,
    sum(revenue)    AS total_revenue,
    count()         AS event_count
FROM events
GROUP BY event_date, event_type;

-- Insert 1M rows/sec on single node
-- Query TBs in seconds with vectorized execution
```

### DuckDB (Embedded OLAP)

```sql
-- "SQLite for analytics" — in-process, no server needed

-- Query Parquet files directly:
SELECT region, SUM(revenue) 
FROM read_parquet('s3://bucket/sales/*.parquet')
GROUP BY region;

-- Query CSV:
SELECT * FROM read_csv_auto('data.csv') WHERE amount > 100;

-- Query PostgreSQL directly:
ATTACH 'postgres:dbname=mydb' AS pg;
SELECT * FROM pg.public.orders WHERE created_at > '2024-01-01';

-- Columnar + vectorized + parallel execution
-- Handles GBs on a laptop, useful for local analytics
```

---

## 5. Lake, Lakehouse, and Table Formats

```
Generation 1: Data Warehouse (Redshift, BigQuery, Snowflake)
  Structured data → ETL → warehouse → BI tools
  
Generation 2: Data Lake (HDFS, S3)
  Raw data dumped in files → "swamp" problem (no schema, no ACID)

Generation 3: Lakehouse (Iceberg, Delta Lake, Hudi)
  Open file formats + ACID + schema enforcement + time travel
  Best of both worlds: cheap storage + warehouse features

Table Formats:
┌─────────────────────────────────────────────────────────────┐
│  Apache Iceberg         Delta Lake           Apache Hudi    │
│  ─────────────────────  ─────────────────── ───────────────│
│  Netflix → Apache       Databricks           Uber → Apache  │
│  Format agnostic        Spark-native          Upsert-first  │
│  (Spark, Flink, Trino)  (expanding ecosystem) (CDC ingestion)│
│  Best catalog support   Best Spark integration Near-realtime │
│  Hidden partitioning    Unity Catalog          Compaction    │
│  Time travel ✓          Time travel ✓          Time travel ✓ │
│  Schema evolution ✓     Schema evolution ✓     Schema evol ✓ │
│  ACID ✓                 ACID ✓                 ACID ✓        │
│                                                              │
│  ★ Industry converging on Iceberg (Snowflake, AWS, Databricks)│
└─────────────────────────────────────────────────────────────┘

File Formats:
  Parquet: Columnar, compressed, the standard for analytics
  ORC: Similar to Parquet, Hive ecosystem
  Avro: Row-based, good for streaming/CDC (schema in file)
```

---

## 6. Approximate Algorithms

```
When exact answers on billions of rows take too long:

HyperLogLog (COUNT DISTINCT approximation):
  - Error: ~0.8% typical
  - Memory: ~12KB regardless of cardinality
  - PostgreSQL: CREATE EXTENSION hll;
    SELECT hll_cardinality(hll_add_agg(hll_hash_text(user_id)))
    FROM events;

Count-Min Sketch (frequency estimation):
  - "How many times did user X visit?"
  - Fixed memory, slight overcount possible
  - Used in stream processing

t-digest (percentile/quantile approximation):
  - P50, P95, P99 without sorting all values
  - Mergeable across distributed nodes
  - TimescaleDB uses this for percentile_agg

Bloom Filter (membership test):
  - "Has this user been seen before?"  
  - False positives possible, false negatives impossible
  - PostgreSQL: CREATE EXTENSION bloom;

-- ClickHouse has many built-in:
SELECT uniqHLL12(user_id) FROM events;              -- HyperLogLog
SELECT quantileTDigest(0.99)(response_time) FROM requests;  -- t-digest
```

---

## 7. ETL vs. ELT

```
ETL (Extract → Transform → Load):
  - Transform BEFORE loading into warehouse
  - Traditional approach (Informatica, Talend, SSIS)
  - Good when compute is expensive (old warehouses)

ELT (Extract → Load → Transform):
  - Load raw data first, transform INSIDE the warehouse
  - Modern approach (dbt + Snowflake/BigQuery)
  - Leverages warehouse compute power
  - Easier to iterate on transformations

Modern Stack (ELT):
  Extract: Fivetran, Airbyte, Stitch (connectors to SaaS APIs, DBs)
  Load: Direct into warehouse/lakehouse
  Transform: dbt (SQL-based transformations, tested, versioned)

dbt example:
  -- models/staging/stg_orders.sql
  SELECT
      id AS order_id,
      customer_id,
      CAST(created_at AS TIMESTAMP) AS ordered_at,
      status,
      total_cents / 100.0 AS total_amount
  FROM {{ source('raw', 'orders') }}
  WHERE _deleted IS FALSE

  -- models/marts/revenue_by_region.sql
  SELECT
      c.region,
      DATE_TRUNC('month', o.ordered_at) AS month,
      SUM(o.total_amount) AS monthly_revenue,
      COUNT(DISTINCT o.customer_id) AS unique_customers
  FROM {{ ref('stg_orders') }} o
  JOIN {{ ref('stg_customers') }} c ON o.customer_id = c.customer_id
  GROUP BY 1, 2
```

---

## Key Takeaways

1. **Star schema is the foundation.** Fact tables hold measurements, dimension tables provide context. Master this before anything else.
2. **Column-oriented storage** is why OLAP engines can scan TBs in seconds — reading only needed columns, with massive compression ratios.
3. **The lakehouse is the future.** Apache Iceberg + Parquet gives you cheap storage with warehouse features (ACID, time travel, schema evolution).
4. **dbt revolutionized transformations.** SQL-based, version-controlled, tested transformations inside the warehouse (ELT > ETL for modern stacks).
5. **Approximate algorithms** (HyperLogLog, t-digest, Count-Min Sketch) trade tiny accuracy for orders-of-magnitude speed gains on analytics queries.

---

Next: [09-data-pipelines.md](09-data-pipelines.md) →
