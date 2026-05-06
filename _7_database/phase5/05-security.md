# 5.5 — Database Security

> Security is not a feature — it's a property of the system.  
> One SQL injection or exposed backup can end a company.  
> Defense in depth: every layer adds protection.

---

## 1. Authentication

```sql
-- PostgreSQL authentication (pg_hba.conf):
-- TYPE   DATABASE  USER       ADDRESS        METHOD
host     all       all        10.0.0.0/8     scram-sha-256
host     replication replicator 10.0.0.0/8   scram-sha-256
hostssl  all       all        0.0.0.0/0      cert

-- Methods (weakest to strongest):
--   trust:        no password (NEVER in production)
--   md5:          MD5 hash (legacy, weak)
--   scram-sha-256: modern, salted, iterated (recommended)
--   cert:         TLS client certificates (strongest)
--   ldap:         LDAP/Active Directory
--   gss:          Kerberos (enterprise SSO)

-- Force SCRAM:
ALTER SYSTEM SET password_encryption = 'scram-sha-256';
-- Then: ALTER USER myuser PASSWORD 'new_password';

-- Certificate authentication:
-- Server: ssl_cert_file, ssl_key_file, ssl_ca_file
-- Client: sslcert, sslkey, sslrootcert in connection string
-- pg_hba.conf: hostssl all all 0.0.0.0/0 cert clientcert=verify-full
```

---

## 2. Authorization — Least Privilege

```sql
-- Principle: grant MINIMUM permissions needed.
-- NEVER use superuser for application connections.

-- Create roles (groups):
CREATE ROLE app_readonly;
CREATE ROLE app_readwrite;
CREATE ROLE app_admin;

-- Grant schema usage:
GRANT USAGE ON SCHEMA public TO app_readonly;
GRANT USAGE ON SCHEMA public TO app_readwrite;

-- Read-only role:
GRANT SELECT ON ALL TABLES IN SCHEMA public TO app_readonly;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO app_readonly;

-- Read-write role (no DDL):
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_readwrite;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO app_readwrite;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_readwrite;

-- Application users (inherit from roles):
CREATE USER api_service WITH PASSWORD 'strong_random_password';
GRANT app_readwrite TO api_service;

CREATE USER reporting_service WITH PASSWORD 'another_strong_password';
GRANT app_readonly TO reporting_service;

-- Revoke public access (important!):
REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE CREATE ON SCHEMA public FROM PUBLIC;

-- Column-level grants:
GRANT SELECT (id, name, email) ON users TO reporting_service;
-- reporting_service cannot see: password_hash, ssn, credit_card
```

---

## 3. SQL Injection Prevention

```
SQL injection: the #1 database security vulnerability (OWASP Top 10).

VULNERABLE (string concatenation):
  query = "SELECT * FROM users WHERE email = '" + user_input + "'"
  # Input: ' OR '1'='1' --
  # Result: SELECT * FROM users WHERE email = '' OR '1'='1' --'
  # → Returns ALL users!

  # Input: '; DROP TABLE users; --
  # Result: SELECT * FROM users WHERE email = ''; DROP TABLE users; --'
  # → Deletes the users table!

SAFE (parameterized queries / prepared statements):
  # Python (psycopg2):
  cursor.execute("SELECT * FROM users WHERE email = %s", (user_input,))
  
  # Java (JDBC):
  PreparedStatement ps = conn.prepareStatement("SELECT * FROM users WHERE email = ?");
  ps.setString(1, userInput);
  
  # Node.js (pg):
  client.query('SELECT * FROM users WHERE email = $1', [userInput]);
  
  # Go (pgx):
  rows, err := db.Query(ctx, "SELECT * FROM users WHERE email = $1", userInput)

Rules:
  1. ALWAYS use parameterized queries (no exceptions)
  2. NEVER concatenate user input into SQL strings
  3. Use an ORM with query builder (additional layer of protection)
  4. Validate and sanitize input at application boundary
  5. Database user should have MINIMUM privileges (limit damage)
```

---

## 4. Encryption

```
Encryption at rest:
  PostgreSQL:
    - OS-level: LUKS/dm-crypt, AWS EBS encryption
    - pgcrypto: column-level encryption
      INSERT INTO users (email, ssn_encrypted)
      VALUES ('alice@x.com', pgp_sym_encrypt('123-45-6789', 'encryption_key'));
    - TDE: not built-in (available in some forks: EnterpriseDB, Percona)
    
  MySQL:
    - InnoDB TDE: built-in (innodb_encrypt_tables = ON)
    - Keyring plugin manages encryption keys

Encryption in transit (TLS):
  PostgreSQL (postgresql.conf):
    ssl = on
    ssl_cert_file = 'server.crt'
    ssl_key_file = 'server.key'
    ssl_ca_file = 'ca.crt'
    ssl_min_protocol_version = 'TLSv1.3'
  
  Connection string:
    postgresql://user:pass@host/db?sslmode=verify-full&sslrootcert=ca.crt
  
  sslmode options:
    disable:     no TLS (never use in production)
    require:     TLS but no certificate verification (MITM possible!)
    verify-ca:   verify server cert is signed by trusted CA
    verify-full: verify cert + hostname match (recommended)

Secrets management:
  NEVER store database passwords in:
    ✗ Source code
    ✗ Environment variables in plain text
    ✗ Config files committed to git
  
  USE:
    ✓ HashiCorp Vault (dynamic database credentials!)
    ✓ AWS Secrets Manager / GCP Secret Manager
    ✓ Kubernetes Secrets (encrypted at rest)
    ✓ .pgpass file with strict permissions (chmod 600)
```

---

## 5. Audit Logging

```sql
-- PostgreSQL: pgaudit extension
CREATE EXTENSION pgaudit;
SET pgaudit.log = 'write, ddl';  -- log writes and schema changes

-- Log output:
-- AUDIT: SESSION,1,1,WRITE,INSERT,TABLE,public.users,
-- "INSERT INTO users (name, email) VALUES ('Alice', 'alice@x.com')"

-- Object-level audit:
SET pgaudit.role = 'auditor';
GRANT SELECT ON users TO auditor;  -- audit all SELECTs on users table

-- Data masking (for non-production environments):
-- pg_anonymize extension or custom views:
CREATE VIEW users_masked AS
SELECT id, 
       LEFT(name, 1) || '***' AS name,
       REGEXP_REPLACE(email, '(.).*@', '\1***@') AS email,
       '***-**-' || RIGHT(ssn, 4) AS ssn
FROM users;
```

---

## Key Takeaways

1. **scram-sha-256 + TLS verify-full** is the minimum for production PostgreSQL. No exceptions.
2. **Parameterized queries prevent SQL injection.** Never concatenate user input into SQL. Use prepared statements in every language.
3. **Least privilege**: application users get only SELECT/INSERT/UPDATE/DELETE. No DDL, no superuser, no CREATEROLE.
4. **Encrypt in transit (TLS 1.3) AND at rest** (disk-level or column-level). Store credentials in a secrets manager, never in code.
5. **Audit logging** with pgaudit is mandatory for compliance (SOC2, HIPAA, GDPR). Know who accessed what and when.

---

Next: [06-monitoring-and-observability.md](06-monitoring-and-observability.md) →
