# 5.4 — Backup & Disaster Recovery

> Backups are worthless if you've never tested a restore.  
> The only backup that matters is one you've successfully recovered from.

---

## 1. Backup Types

```
Full backup:      entire database (large, slow, self-contained)
Incremental:      only changes since last backup (small, fast, needs chain)
Differential:     changes since last FULL backup (medium, needs full + latest diff)

Logical backup:   SQL dump (portable, slow, table-level)
Physical backup:  raw data files + WAL (fast, binary-level, full cluster)

                 Full    Incremental  Differential
Backup speed:    Slow    Fast         Medium
Restore speed:   Fast    Slow (chain) Medium
Storage:         Large   Small        Medium
Complexity:      Low     High         Medium
```

---

## 2. PostgreSQL Backup Tools

```bash
# LOGICAL BACKUPS:

# pg_dump (single database):
pg_dump -Fc mydb > mydb.dump          # custom format (compressed, parallel restore)
pg_dump -Fp mydb > mydb.sql           # plain SQL
pg_dump -Fd -j 4 mydb -f ./dump_dir   # directory format (parallel dump)

# pg_dumpall (all databases + globals):
pg_dumpall > cluster.sql               # roles, tablespaces, all databases

# Restore:
pg_restore -d mydb -j 4 mydb.dump     # parallel restore from custom format
psql mydb < mydb.sql                   # plain SQL restore

# Limitations of logical backups:
# - Slow for large databases (>100 GB)
# - Point-in-time is the moment of dump start
# - Need to re-create indexes (slow)
# - Good for: selective restore, cross-version migration

# PHYSICAL BACKUPS:

# pg_basebackup (built-in):
pg_basebackup -h primary -D /backup/base -Fp -Xs -P
# -Fp: plain format
# -Xs: stream WAL during backup (ensure no gaps)
# -P: progress reporting

# PITR (Point-in-Time Recovery):
# 1. Take base backup (pg_basebackup)
# 2. Archive WAL continuously (archive_command or pgBackRest)
# 3. To recover to a specific time:
#    - Restore base backup
#    - Set recovery_target_time in postgresql.conf:
#      restore_command = 'cp /archive/%f %p'
#      recovery_target_time = '2024-06-15 14:30:00'
#    - Start PostgreSQL → replays WAL up to that point

# pgBackRest (production standard):
# Full backup:
pgbackrest --stanza=mydb backup --type=full
# Incremental:
pgbackrest --stanza=mydb backup --type=incr
# Differential:
pgbackrest --stanza=mydb backup --type=diff

# Restore to point in time:
pgbackrest --stanza=mydb --type=time \
  --target="2024-06-15 14:30:00+00" restore

# pgBackRest features:
# - Parallel backup and restore
# - Compression (lz4, zstd)
# - Encryption at rest
# - Remote backup via TLS
# - S3/GCS/Azure blob storage
# - Backup verification (checksum validation)

# Barman (another production tool):
barman backup myserver                  # full backup
barman recover myserver 20240615T143000 /restore/path \
  --target-time "2024-06-15 14:30:00"   # PITR restore
```

---

## 3. MySQL Backup Tools

```bash
# Logical: mysqldump
mysqldump --single-transaction --routines --triggers mydb > mydb.sql

# Physical: Percona XtraBackup (hot backup, no locks for InnoDB)
xtrabackup --backup --target-dir=/backup/full
xtrabackup --prepare --target-dir=/backup/full    # apply redo logs
xtrabackup --copy-back --target-dir=/backup/full   # restore

# Incremental:
xtrabackup --backup --target-dir=/backup/incr \
  --incremental-basedir=/backup/full
```

---

## 4. Backup Strategy for Production

```
Typical production backup strategy:

Schedule:
  Daily:  full backup (pgBackRest full or differential)
  Hourly: incremental backup
  Continuous: WAL archiving to S3 (every WAL segment, ~16 MB)

Retention:
  Keep daily backups for 30 days
  Keep weekly backups for 3 months
  Keep monthly backups for 1 year

Storage:
  Primary: local fast storage (NVMe) for recent backups
  Secondary: object storage (S3/GCS) for all backups
  Tertiary: cross-region S3 bucket for DR

Testing (CRITICAL):
  Weekly: automated restore test to a test server
  Monthly: full DR drill (restore from scratch, measure RTO)
  
  # Automated restore test (cron job):
  pgbackrest --stanza=mydb restore --target-dir=/tmp/restore-test
  pg_isready -h /tmp/restore-test -p 5433
  psql -h /tmp/restore-test -p 5433 -c "SELECT count(*) FROM orders;"
  # Compare count with production → alert on mismatch

RTO/RPO targets:
  RPO = 0:     synchronous replication (no backup can achieve this)
  RPO < 1 min: continuous WAL archiving
  RPO < 1 hr:  hourly incremental backups
  RTO < 5 min: standby server (already running, just promote)
  RTO < 1 hr:  restore from physical backup
  RTO > 1 hr:  restore from logical backup (large databases)
```

---

## Key Takeaways

1. **pgBackRest is the production standard** for PostgreSQL backups. Full + incremental + S3 + encryption + parallel.
2. **Continuous WAL archiving** gives you RPO of seconds (recover to any point in time).
3. **Test your restores.** Automate weekly restore tests. A backup you've never restored is a backup you don't have.
4. **Physical backups for speed, logical for portability.** Use pg_basebackup/pgBackRest for production, pg_dump for migrations and selective restore.
5. **Cross-region backup replication** is mandatory for true disaster recovery. If the region goes down, your backups shouldn't be there too.

---

Next: [05-security.md](05-security.md) →
