# 5.9 — Data Pipelines & CDC

> The database is not the end. It's the beginning.  
> Data must flow — from source to warehouse, from OLTP to OLAP,  
> from table to stream, from raw to refined.

---

## 1. Change Data Capture (CDC)

```
CDC: Capture every INSERT, UPDATE, DELETE from a database
and stream it somewhere else — in real time.

Why CDC matters:
  - Real-time data replication (OLTP → analytics)
  - Event-driven microservices (DB change → trigger action)
  - Cache invalidation (DB change → invalidate Redis)
  - Search index sync (DB change → update Elasticsearch)
  - Audit trail (every change captured)

CDC approaches:
┌────────────────────── ────────────────── ──────────────────────────┐
│ Approach              Lag                Trade-offs               │
├────────────────────── ────────────────── ──────────────────────────┤
│ Trigger-based         Near real-time     Adds overhead to writes  │
│ Timestamp polling     Seconds-minutes    Misses deletes, lag      │
│ Log-based (WAL/binlog)Real-time          ★ Best: no overhead on DB│
│ Application-level     Real-time          Requires code changes    │
└────────────────────── ────────────────── ──────────────────────────┘

Log-based CDC is the gold standard:
  PostgreSQL: reads WAL (Write-Ahead Log) via logical replication slots
  MySQL: reads binlog (binary log)
  No triggers, no polling, no overhead on the source database.
```

### Debezium (The Standard CDC Platform)

```
Architecture:
  ┌──────────┐    ┌──────────────┐    ┌─────────┐    ┌──────────────┐
  │PostgreSQL │──→ │   Debezium   │──→ │  Kafka  │──→ │ Consumers    │
  │ (WAL)    │    │  Connector   │    │ Topics  │    │ (warehouse,  │
  └──────────┘    └──────────────┘    └─────────┘    │  search, etc)│
                                                      └──────────────┘

Debezium reads the WAL/binlog and produces Kafka messages for each change.
```

```json
// Debezium change event (simplified):
{
  "op": "u",                          // c=create, u=update, d=delete, r=read(snapshot)
  "before": {                          // previous state (null for inserts)
    "id": 1001,
    "email": "old@email.com",
    "status": "active"
  },
  "after": {                           // new state (null for deletes)
    "id": 1001,
    "email": "new@email.com",
    "status": "active"
  },
  "source": {
    "version": "2.5.0",
    "connector": "postgresql",
    "db": "mydb",
    "schema": "public",
    "table": "users",
    "txId": 58293,
    "lsn": 3378291048,
    "ts_ms": 1706000000000
  },
  "ts_ms": 1706000000123
}
```

```json
// Debezium PostgreSQL connector config (Kafka Connect):
{
  "name": "pg-source-connector",
  "config": {
    "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
    "database.hostname": "pg-primary",
    "database.port": "5432",
    "database.user": "debezium",
    "database.password": "${vault:secret/debezium:password}",
    "database.dbname": "mydb",
    "topic.prefix": "myapp",
    "schema.include.list": "public",
    "table.include.list": "public.users,public.orders",
    "plugin.name": "pgoutput",
    "slot.name": "debezium_slot",
    "publication.name": "debezium_pub",
    "snapshot.mode": "initial",
    "transforms": "route",
    "transforms.route.type": "io.debezium.transforms.ByLogicalTableRouter",
    "transforms.route.topic.regex": "myapp\\.public\\.(.*)",
    "transforms.route.topic.replacement": "myapp.$1"
  }
}
```

```sql
-- PostgreSQL setup for Debezium:
-- 1. Set wal_level = logical in postgresql.conf
ALTER SYSTEM SET wal_level = 'logical';
-- Restart required

-- 2. Create replication user:
CREATE ROLE debezium WITH REPLICATION LOGIN PASSWORD 'secure_password';
GRANT SELECT ON ALL TABLES IN SCHEMA public TO debezium;

-- 3. Create publication:
CREATE PUBLICATION debezium_pub FOR TABLE users, orders;

-- Monitor replication slot lag:
SELECT slot_name,
       pg_wal_lsn_diff(pg_current_wal_lsn(), confirmed_flush_lsn) AS lag_bytes
FROM pg_replication_slots
WHERE slot_name = 'debezium_slot';
-- If lag_bytes keeps growing → consumer is too slow
-- ⚠ Unconsumed slots prevent WAL cleanup → disk fills up!
```

---

## 2. Kafka Connect

```
Kafka Connect: Framework for streaming data in/out of Kafka

Source Connectors (into Kafka):
  - Debezium (PostgreSQL, MySQL, MongoDB, SQL Server, Oracle)
  - JDBC Source (poll-based, any JDBC database)
  - FileStream, S3, GCS

Sink Connectors (out of Kafka):
  - JDBC Sink → write to any database
  - Elasticsearch Sink → search index
  - S3 Sink → data lake (Parquet/JSON/Avro)
  - BigQuery Sink, Snowflake Sink
  - ClickHouse Sink
  - Redis Sink

Architecture:
  ┌─────────┐                              ┌─────────────────┐
  │ Sources  │    ┌────────────────────┐    │ Sinks           │
  │          │    │   Kafka Connect    │    │                 │
  │ Postgres ├───→│  ┌──────────────┐  │───→│ Elasticsearch   │
  │ MySQL    ├───→│  │  Workers     │  │───→│ S3 (Parquet)    │
  │ MongoDB  ├───→│  │  (parallel)  │  │───→│ Snowflake       │
  │ S3       ├───→│  └──────────────┘  │───→│ Redis           │
  └─────────┘    └────────────────────┘    └─────────────────┘

Key concepts:
  - Connectors define the job (source/sink + config)
  - Tasks are the parallelism unit (one connector → multiple tasks)
  - Workers are the JVM processes that run tasks
  - Distributed mode: workers form a cluster, tasks auto-balanced
  - Single Message Transforms (SMTs): lightweight in-flight transformations
```

---

## 3. Schema Registry

```
Problem: Producers and consumers must agree on data format.
         Schema changes can break downstream consumers.

Solution: Schema Registry (Confluent, AWS Glue, Apicurio)

  Producer                  Schema Registry              Consumer
  ────────                  ───────────────              ────────
  1. Register schema ──→    Stores schema, assigns ID
  2. Serialize message      (schema ID + data bytes)
  3. Send to Kafka topic
                                                          4. Read message
                            5. ←── Fetch schema by ID     (get schema ID)
                                                          6. Deserialize
                                                             with schema

Compatibility modes:
  BACKWARD:  new schema can read old data (safe for consumers)
  FORWARD:   old schema can read new data (safe for producers)  
  FULL:      both backward and forward compatible
  NONE:      no compatibility checking

Safe schema changes (BACKWARD compatible):
  ✓ Add optional field (with default)
  ✓ Remove field (consumer ignores unknown fields)
  ✗ Add required field (breaks old consumers)
  ✗ Change field type (breaks everything)
  ✗ Rename field (breaks everything)

Formats:
  Avro:     Most popular for Kafka. Schema in registry, data compact.
  Protobuf: Google's format. Strongly typed, great for gRPC + Kafka.
  JSON Schema: Human-readable but larger payloads.
```

---

## 4. Stream Processing

### Apache Flink

```sql
-- Flink SQL: Process streams with SQL

-- Define a Kafka source:
CREATE TABLE orders (
    order_id    BIGINT,
    customer_id BIGINT,
    amount      DECIMAL(10,2),
    order_time  TIMESTAMP(3),
    WATERMARK FOR order_time AS order_time - INTERVAL '5' SECOND
) WITH (
    'connector' = 'kafka',
    'topic' = 'myapp.orders',
    'properties.bootstrap.servers' = 'kafka:9092',
    'format' = 'debezium-json'
);

-- Real-time aggregation with tumbling window:
SELECT
    customer_id,
    TUMBLE_START(order_time, INTERVAL '1' HOUR) AS window_start,
    COUNT(*) AS order_count,
    SUM(amount) AS total_amount
FROM orders
GROUP BY customer_id, TUMBLE(order_time, INTERVAL '1' HOUR);

-- Sliding window (every 5 min, look back 1 hour):
SELECT
    customer_id,
    HOP_START(order_time, INTERVAL '5' MINUTE, INTERVAL '1' HOUR) AS window_start,
    SUM(amount) AS rolling_1h_total
FROM orders
GROUP BY customer_id, HOP(order_time, INTERVAL '5' MINUTE, INTERVAL '1' HOUR);

-- Pattern detection (CEP): find users who place 3+ orders in 10 minutes
SELECT *
FROM orders
MATCH_RECOGNIZE (
    PARTITION BY customer_id
    ORDER BY order_time
    MEASURES
        FIRST(A.order_time) AS first_order,
        LAST(A.order_time) AS last_order,
        COUNT(A.order_id) AS order_count
    ONE ROW PER MATCH
    AFTER MATCH SKIP PAST LAST ROW
    PATTERN (A{3,})
    DEFINE
        A AS TRUE
    HAVING LAST(A.order_time) - FIRST(A.order_time) < INTERVAL '10' MINUTE
);
```

### Apache Spark Structured Streaming

```python
# Spark: micro-batch streaming (or continuous mode)
from pyspark.sql import SparkSession
from pyspark.sql.functions import window, sum, count

spark = SparkSession.builder.appName("orders-stream").getOrCreate()

# Read from Kafka
orders = (spark.readStream
    .format("kafka")
    .option("kafka.bootstrap.servers", "kafka:9092")
    .option("subscribe", "myapp.orders")
    .load()
    .selectExpr("CAST(value AS STRING)")
    # Parse JSON...
)

# Windowed aggregation
revenue = (orders
    .groupBy(
        window("order_time", "1 hour", "15 minutes"),
        "region"
    )
    .agg(
        sum("amount").alias("total_revenue"),
        count("*").alias("order_count")
    )
)

# Write to PostgreSQL
(revenue.writeStream
    .foreachBatch(write_to_postgres)
    .outputMode("update")
    .option("checkpointLocation", "/checkpoints/revenue")
    .start()
)
```

---

## 5. dbt (Data Build Tool)

```
dbt = SQL-based transformation layer for your warehouse.
"Software engineering for analytics."

Project structure:
  dbt_project/
  ├── dbt_project.yml          # project config
  ├── models/
  │   ├── staging/              # 1:1 with source tables (clean/rename)
  │   │   ├── stg_orders.sql
  │   │   ├── stg_customers.sql
  │   │   └── _staging.yml      # tests + docs
  │   ├── intermediate/         # business logic (joins, transforms)
  │   │   └── int_order_items.sql
  │   └── marts/                # final tables for analysts/BI
  │       ├── fct_revenue.sql
  │       └── dim_customers.sql
  ├── tests/                    # custom data tests
  ├── macros/                   # reusable SQL macros (Jinja)
  └── seeds/                    # small CSV lookups
```

```sql
-- models/staging/stg_orders.sql
{{ config(materialized='view') }}

SELECT
    id          AS order_id,
    user_id     AS customer_id,
    status,
    total_cents / 100.0 AS total_amount,
    created_at  AS ordered_at
FROM {{ source('raw', 'orders') }}
WHERE _deleted IS FALSE

-- models/marts/fct_revenue.sql
{{ config(materialized='incremental', unique_key='date_day') }}

SELECT
    DATE_TRUNC('day', o.ordered_at)  AS date_day,
    c.region,
    c.segment,
    COUNT(DISTINCT o.order_id)       AS order_count,
    COUNT(DISTINCT o.customer_id)    AS customer_count,
    SUM(o.total_amount)              AS revenue
FROM {{ ref('stg_orders') }} o
JOIN {{ ref('stg_customers') }} c ON o.customer_id = c.customer_id
WHERE o.status = 'completed'
{% if is_incremental() %}
  AND o.ordered_at >= (SELECT MAX(date_day) FROM {{ this }})
{% endif %}
GROUP BY 1, 2, 3
```

```yaml
# models/staging/_staging.yml — tests and documentation
version: 2
models:
  - name: stg_orders
    description: "Cleaned orders from source, one row per order"
    columns:
      - name: order_id
        tests:
          - unique
          - not_null
      - name: customer_id
        tests:
          - not_null
          - relationships:
              to: ref('stg_customers')
              field: customer_id
      - name: total_amount
        tests:
          - not_null
          - dbt_utils.accepted_range:
              min_value: 0
              max_value: 100000
```

```bash
# dbt commands:
dbt run                    # execute all models
dbt run --select marts.*   # only marts models
dbt test                   # run all tests
dbt build                  # run + test (proper order)
dbt docs generate && dbt docs serve  # auto-generated docs + lineage graph
```

---

## 6. Orchestration

```
Orchestrators schedule and monitor pipeline DAGs.

Apache Airflow (the standard):
  - Python DAGs define task dependencies
  - Web UI for monitoring, retries, alerting
  - Huge ecosystem of operators (DB, cloud, dbt, Spark)

Dagster:
  - Asset-centric (vs Airflow's task-centric)
  - Better local dev experience
  - Type system and testing built in

Prefect:
  - Python-native, simpler than Airflow
  - Good for smaller teams

Temporal:
  - Workflow engine (not just data pipelines)
  - Durable execution with automatic retries
```

```python
# Airflow DAG example:
from airflow import DAG
from airflow.operators.bash import BashOperator
from airflow.providers.postgres.operators.postgres import PostgresOperator
from datetime import datetime, timedelta

default_args = {
    'retries': 2,
    'retry_delay': timedelta(minutes=5),
}

with DAG(
    'daily_analytics',
    default_args=default_args,
    schedule_interval='@daily',
    start_date=datetime(2024, 1, 1),
    catchup=False,
) as dag:

    extract = PostgresOperator(
        task_id='extract_new_orders',
        postgres_conn_id='source_db',
        sql='sql/extract_orders.sql',
    )

    transform = BashOperator(
        task_id='dbt_run',
        bash_command='cd /dbt && dbt run --select marts.fct_revenue',
    )

    test = BashOperator(
        task_id='dbt_test',
        bash_command='cd /dbt && dbt test --select marts.fct_revenue',
    )

    extract >> transform >> test
```

---

## 7. End-to-End Pipeline Architecture

```
The Modern Data Stack:

┌──────────────────────────────────────────────────────────────────┐
│                        Sources                                    │
│  PostgreSQL  │  MySQL  │  SaaS APIs  │  Event Streams  │  Files  │
└──────┬───────┴────┬────┴──────┬──────┴───────┬─────────┴────┬───┘
       │            │           │              │              │
       ▼            ▼           ▼              ▼              ▼
┌──────────────────────────────────────────────────────────────────┐
│                     Ingestion Layer                               │
│  Debezium/CDC  │  Airbyte/Fivetran  │  Kafka  │  Custom ETL     │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────────┐
│                     Storage Layer                                 │
│  Data Lake (S3/GCS + Iceberg)  │  Warehouse (Snowflake/BigQuery) │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────────┐
│                  Transformation Layer                             │
│                    dbt (SQL models)                               │
│  staging → intermediate → marts                                  │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────────┐
│                    Serving Layer                                  │
│  BI (Metabase/Looker)  │  ML Features  │  Reverse ETL  │  APIs  │
└──────────────────────────────────────────────────────────────────┘

Orchestration: Airflow/Dagster manages the entire flow
Monitoring: Great Expectations / dbt tests for data quality
Catalog: DataHub / Amundsen / OpenMetadata for discovery
```

---

## Key Takeaways

1. **Log-based CDC (Debezium)** is the gold standard for streaming database changes. It reads the WAL/binlog with zero overhead on the source database.
2. **Schema Registry** prevents breaking changes in streaming pipelines. Always enforce at least BACKWARD compatibility.
3. **Flink for real-time, Spark for near-real-time.** Flink processes event-by-event; Spark uses micro-batches. Both support SQL.
4. **dbt transformed analytics engineering.** Version-controlled SQL, tests, docs, incremental builds — it's Git for your warehouse transformations.
5. **The modern data stack** is: CDC/connectors → Kafka → Lake/Warehouse → dbt → BI tools, orchestrated by Airflow.

---

Next: [10-cloud-db-services.md](10-cloud-db-services.md) →
