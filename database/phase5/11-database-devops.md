# 5.11 — Database DevOps & Infrastructure as Code

> Treat your database like code:  
> version it, test it, review it, deploy it automatically.

---

## 1. Infrastructure as Code (Terraform)

```hcl
# Terraform: Define database infrastructure declaratively

# ── AWS RDS PostgreSQL ────────────────────────────────
resource "aws_db_instance" "primary" {
  identifier     = "myapp-primary"
  engine         = "postgres"
  engine_version = "16.2"
  instance_class = "db.r6g.xlarge"

  # Storage
  allocated_storage     = 100
  max_allocated_storage = 500    # auto-scaling up to 500GB
  storage_type          = "gp3"
  storage_encrypted     = true
  kms_key_id            = aws_kms_key.db.arn

  # Networking
  db_subnet_group_name   = aws_db_subnet_group.private.name
  vpc_security_group_ids = [aws_security_group.db.id]
  publicly_accessible    = false    # NEVER public

  # Credentials
  db_name  = "myapp"
  username = "myapp_admin"
  manage_master_user_password = true    # AWS Secrets Manager

  # High Availability
  multi_az = true

  # Backup
  backup_retention_period = 14
  backup_window           = "03:00-04:00"
  maintenance_window      = "Mon:04:00-Mon:05:00"
  copy_tags_to_snapshot   = true

  # Monitoring
  monitoring_interval          = 60
  monitoring_role_arn          = aws_iam_role.rds_monitoring.arn
  performance_insights_enabled = true
  performance_insights_retention_period = 731    # 2 years

  # Parameters
  parameter_group_name = aws_db_parameter_group.pg16.name

  # Lifecycle
  deletion_protection       = true
  skip_final_snapshot       = false
  final_snapshot_identifier = "myapp-primary-final"

  tags = {
    Environment = "production"
    Team        = "platform"
    ManagedBy   = "terraform"
  }
}

# ── Parameter Group ───────────────────────────────────
resource "aws_db_parameter_group" "pg16" {
  family = "postgres16"
  name   = "myapp-pg16"

  parameter {
    name         = "shared_preload_libraries"
    value        = "pg_stat_statements,auto_explain"
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "log_min_duration_statement"
    value = "1000"    # log queries > 1 second
  }

  parameter {
    name  = "max_connections"
    value = "200"
  }

  parameter {
    name  = "work_mem"
    value = "65536"    # 64MB
  }
}

# ── Read Replica ──────────────────────────────────────
resource "aws_db_instance" "read_replica" {
  identifier          = "myapp-read-1"
  replicate_source_db = aws_db_instance.primary.identifier
  instance_class      = "db.r6g.large"
  
  # Replica-specific
  publicly_accessible = false
  skip_final_snapshot = true
  
  tags = {
    Environment = "production"
    Role        = "read-replica"
  }
}
```

### Terraform Best Practices for Databases

```
1. State management:
   - Use remote state (S3 + DynamoDB lock)
   - NEVER store DB passwords in state (use AWS Secrets Manager)

2. Lifecycle rules:
   lifecycle {
     prevent_destroy = true          # prevent accidental deletion
     ignore_changes  = [password]    # password managed externally
   }

3. Separate environments:
   environments/
     production/
       main.tf
       terraform.tfvars
     staging/
       main.tf
       terraform.tfvars
   OR use Terraform workspaces / Terragrunt

4. Import existing databases:
   terraform import aws_db_instance.primary myapp-primary
   # Then write the HCL to match current state

5. Plan before apply — ALWAYS:
   terraform plan -out=plan.tfplan
   terraform apply plan.tfplan
   # Review the plan. DB changes can cause downtime (instance class change = reboot)
```

---

## 2. Kubernetes Operators

### CloudNativePG (PostgreSQL on Kubernetes)

```yaml
# CloudNativePG: The leading PostgreSQL operator for K8s
# Manages the full lifecycle: provisioning, HA, backups, scaling

apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: myapp-db
  namespace: database
spec:
  instances: 3    # 1 primary + 2 replicas (auto-failover)
  
  # PostgreSQL configuration
  postgresql:
    parameters:
      shared_buffers: "256MB"
      effective_cache_size: "768MB"
      max_connections: "200"
      shared_preload_libraries: "pg_stat_statements"
    pg_hba:
      - host all all 10.244.0.0/16 scram-sha-256    # pod network

  # Storage
  storage:
    size: 100Gi
    storageClass: gp3-encrypted

  # Backup to S3
  backup:
    barmanObjectStore:
      destinationPath: s3://myapp-backups/cnpg/
      s3Credentials:
        accessKeyId:
          name: aws-creds
          key: ACCESS_KEY_ID
        secretAccessKey:
          name: aws-creds
          key: SECRET_ACCESS_KEY
      wal:
        compression: gzip
    retentionPolicy: "30d"

  # Scheduled backups
  backup:
    barmanObjectStore: ...  # same as above
  
  # Resources
  resources:
    requests:
      memory: "1Gi"
      cpu: "1"
    limits:
      memory: "2Gi"
      cpu: "2"

  # Monitoring
  monitoring:
    enablePodMonitor: true    # Prometheus integration
---
# Scheduled backup
apiVersion: postgresql.cnpg.io/v1
kind: ScheduledBackup
metadata:
  name: daily-backup
spec:
  schedule: "0 3 * * *"    # 3 AM daily
  cluster:
    name: myapp-db
  backupOwnerReference: self
```

```bash
# CloudNativePG operations:
kubectl cnpg status myapp-db          # cluster health
kubectl cnpg promote myapp-db-2       # manual failover
kubectl cnpg psql myapp-db -- -c "SELECT 1"   # connect to primary
```

### Zalando PostgreSQL Operator

```yaml
# Alternative: Zalando's postgres-operator (used at Zalando)
apiVersion: acid.zalan.do/v1
kind: postgresql
metadata:
  name: myapp-db
  namespace: database
spec:
  teamId: "myteam"
  numberOfInstances: 3
  postgresql:
    version: "16"
    parameters:
      shared_buffers: "256MB"
  volume:
    size: 100Gi
    storageClass: gp3-encrypted
  users:
    myapp_user:
      - superuser
      - createdb
  databases:
    myapp: myapp_user
  resources:
    requests:
      cpu: "1"
      memory: "2Gi"
    limits:
      cpu: "2"
      memory: "4Gi"
  patroni:
    ttl: 30
    loop_wait: 10
    maximum_lag_on_failover: 33554432    # 32MB
```

### When to Use K8s Operators vs. Managed Services

```
Kubernetes Operators (CloudNativePG, Zalando):
  ✓ Full control over configuration
  ✓ No vendor lock-in
  ✓ Same deployment model as your app
  ✓ Cost-effective on existing K8s clusters
  ✗ Operational complexity (you manage EVERYTHING)
  ✗ Storage depends on underlying block storage
  ✗ Networking complexity (service mesh, ingress)
  ✗ Backup/restore is your responsibility

Managed Services (RDS, Aurora, Cloud SQL):
  ✓ Zero operational overhead
  ✓ Battle-tested at scale
  ✓ Built-in monitoring, backups, patching
  ✓ Better storage (Aurora's distributed storage, etc.)
  ✗ Vendor lock-in
  ✗ Less control over configuration
  ✗ More expensive (paying for management)

Recommendation:
  - Production: Managed services (unless you have a dedicated DB team)
  - Dev/test: K8s operators (ephemeral, disposable environments)
  - Strong K8s team + cost constraints: K8s operators can work for production
```

---

## 3. CI/CD for Databases

```yaml
# GitHub Actions: Database migration CI/CD pipeline

name: Database Migrations
on:
  push:
    branches: [main]
    paths: ['migrations/**']
  pull_request:
    paths: ['migrations/**']

jobs:
  # 1. Lint and validate migrations
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Lint SQL migrations
        run: |
          # squawk: PostgreSQL migration linter
          # Catches dangerous patterns (e.g., adding NOT NULL without default)
          npm install -g squawk-cli
          squawk migrations/*.sql
      
      - name: Check migration naming
        run: |
          # Ensure sequential numbering, no gaps
          ls migrations/ | sort -V | awk -F'_' '{
            if (NR != $1) { print "Gap at " NR; exit 1 }
          }'

  # 2. Test migrations against a fresh database
  test:
    runs-on: ubuntu-latest
    needs: validate
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_DB: test_db
          POSTGRES_USER: test_user
          POSTGRES_PASSWORD: test_pass
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
    steps:
      - uses: actions/checkout@v4
      
      - name: Run migrations (up)
        env:
          DATABASE_URL: postgres://test_user:test_pass@localhost:5432/test_db
        run: |
          # Using golang-migrate
          migrate -path migrations -database "$DATABASE_URL" up
      
      - name: Verify schema
        run: |
          PGPASSWORD=test_pass psql -h localhost -U test_user -d test_db \
            -c "\dt" -c "\di"
      
      - name: Test rollback
        env:
          DATABASE_URL: postgres://test_user:test_pass@localhost:5432/test_db
        run: |
          migrate -path migrations -database "$DATABASE_URL" down 1
          migrate -path migrations -database "$DATABASE_URL" up

  # 3. Deploy to staging (auto on merge to main)
  deploy-staging:
    runs-on: ubuntu-latest
    needs: test
    if: github.ref == 'refs/heads/main'
    environment: staging
    steps:
      - uses: actions/checkout@v4
      
      - name: Run migrations on staging
        env:
          DATABASE_URL: ${{ secrets.STAGING_DATABASE_URL }}
        run: |
          migrate -path migrations -database "$DATABASE_URL" up
      
      - name: Smoke test
        run: ./scripts/db-smoke-test.sh staging

  # 4. Deploy to production (manual approval)
  deploy-production:
    runs-on: ubuntu-latest
    needs: deploy-staging
    environment: production    # requires manual approval in GitHub
    steps:
      - uses: actions/checkout@v4
      
      - name: Create backup before migration
        run: ./scripts/create-backup.sh production
      
      - name: Run migrations on production
        env:
          DATABASE_URL: ${{ secrets.PRODUCTION_DATABASE_URL }}
        run: |
          migrate -path migrations -database "$DATABASE_URL" up
      
      - name: Post-migration health check
        run: ./scripts/db-health-check.sh production
```

### SQL Linting with squawk

```bash
# squawk catches dangerous PostgreSQL migration patterns:

$ squawk migrations/005_add_not_null.sql
migrations/005_add_not_null.sql:1:1: warning: adding-not-nullable-field
   Adding a NOT NULL column without a DEFAULT will lock the table and
   rewrite all rows. Add a DEFAULT or use a CHECK constraint.

   ALTER TABLE orders ADD COLUMN status TEXT NOT NULL;
   
   Prefer:
   ALTER TABLE orders ADD COLUMN status TEXT DEFAULT 'pending';
   -- Then: ALTER TABLE orders ALTER COLUMN status SET NOT NULL;

# Other rules squawk catches:
#   - adding-serial-primary-key (use IDENTITY instead)
#   - ban-drop-column (use soft deletes)
#   - prefer-text-field (avoid VARCHAR(n))
#   - require-concurrent-index-creation
#   - adding-foreign-key-constraint (use NOT VALID + VALIDATE)
```

---

## 4. Chaos Engineering for Databases

```bash
# Chaos engineering: intentionally break things to find weaknesses

# 1. Kill the primary and verify automatic failover:
#    Patroni/CloudNativePG/RDS Multi-AZ should promote a replica
#    Measure: failover time, application error rate, data loss

# 2. Network partition simulation (using tc or Toxiproxy):
# Add 500ms latency between app and database:
tc qdisc add dev eth0 root netem delay 500ms

# 3. Disk full simulation:
fallocate -l $(($(df --output=avail /pgdata | tail -1) * 1024 - 1048576)) /pgdata/fillfile
# Database should: stop accepting writes, alert, not corrupt data
rm /pgdata/fillfile

# 4. Connection storm:
pgbench -c 500 -j 50 -T 60 -h localhost mydb
# Verify: connection pooler (PgBouncer) handles it, no OOM

# 5. Long-running transaction blocking:
psql -c "BEGIN; SELECT * FROM orders LIMIT 1;"
# Leave this open. Watch: bloat, lock contention, autovacuum blocking
# Kill after observation: SELECT pg_terminate_backend(pid);

# 6. Replica lag injection:
# On replica: pg_ctl stop && sleep 300 && pg_ctl start
# Verify: application detects stale reads, alerts fire, replica catches up

# Tools:
# - LitmusChaos: Kubernetes-native chaos (has DB experiments)
# - Gremlin: SaaS chaos platform
# - Toxiproxy (Shopify): TCP proxy for simulating network conditions
# - kill -9 / docker stop: simple but effective
```

---

## 5. Database Environment Management

```
Production-like environments are critical for testing.

Strategies:
┌───────────────────── ───────────────── ──────────────────────────┐
│ Strategy             Speed/Cost        Fidelity                  │
├───────────────────── ───────────────── ──────────────────────────┤
│ Full copy            Slow, expensive   Perfect (real data)       │
│ Aurora clone         Instant, cheap     Perfect (copy-on-write)  │
│ Neon branch          Instant, free      Perfect (copy-on-write)  │
│ Subset copy          Medium            Good (representative data)│
│ Synthetic data       Fast              Ok (may miss edge cases)  │
│ Schema-only          Instant           Schema only, empty tables │
└───────────────────── ───────────────── ──────────────────────────┘

Data masking for non-production:
  - Anonymize PII before copying to dev/staging
  - Use PostgreSQL Anonymizer extension
  - Or mask during export:
    pg_dump | sed 's/real@email.com/fake@test.com/g'  # crude
    # Better: use purpose-built tools (Delphix, Tonic, Snaplet)
```

---

## 6. Containerized Database Development

```yaml
# docker-compose.yml for local development
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: myapp_dev
      POSTGRES_USER: developer
      POSTGRES_PASSWORD: devpassword
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
      - ./init-scripts:/docker-entrypoint-initdb.d    # auto-run on first start
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U developer -d myapp_dev"]
      interval: 5s
      timeout: 3s
      retries: 5
    # Performance tuning for dev:
    command: >
      postgres
        -c shared_buffers=256MB
        -c work_mem=64MB
        -c maintenance_work_mem=128MB
        -c shared_preload_libraries=pg_stat_statements
        -c log_min_duration_statement=100

  pgbouncer:
    image: edoburu/pgbouncer
    environment:
      DATABASE_URL: postgres://developer:devpassword@postgres:5432/myapp_dev
      POOL_MODE: transaction
      MAX_CLIENT_CONN: 200
      DEFAULT_POOL_SIZE: 20
    ports:
      - "6432:6432"
    depends_on:
      postgres:
        condition: service_healthy

volumes:
  pgdata:
```

```bash
# Makefile for database operations:
.PHONY: db-up db-down db-reset db-migrate db-seed db-console

db-up:
	docker compose up -d postgres pgbouncer

db-down:
	docker compose down

db-reset:
	docker compose down -v    # -v removes volumes (destroys data)
	docker compose up -d postgres pgbouncer
	$(MAKE) db-migrate
	$(MAKE) db-seed

db-migrate:
	migrate -path migrations -database "$$DATABASE_URL" up

db-seed:
	psql "$$DATABASE_URL" -f seeds/development.sql

db-console:
	psql "$$DATABASE_URL"

db-dump:
	pg_dump "$$DATABASE_URL" --schema-only > schema.sql

db-diff:
	# Compare current schema vs migration target
	migra "$$DATABASE_URL" "$$SHADOW_DATABASE_URL"
```

---

## Key Takeaways

1. **Terraform for infrastructure, migrations for schema.** Never manage database infrastructure manually. Terraform for the server, migration tools for the schema.
2. **CloudNativePG** is the leading Kubernetes PostgreSQL operator. For production, prefer managed services unless you have a dedicated platform team.
3. **CI/CD for migrations**: lint with squawk, test against a fresh database in CI, deploy to staging automatically, require manual approval for production.
4. **Chaos engineering** finds failures before your users do. Kill the primary, simulate network partitions, fill disks — verify your HA setup actually works.
5. **Instant database cloning** (Aurora, Neon) revolutionizes dev/test workflows. No more waiting hours for database copies.

---

Phase 5 complete. Next: [Phase 6 — God Tier](../phase6/) →
