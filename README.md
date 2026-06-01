# Seedling — Relational Test Data Factory

A schema-aware test data generator that understands your database. Point it at a Postgres schema, write minimal generator configs for tables with complex logic, and it produces millions of rows that respect every foreign key, unique constraint, and column type.

## The Problem

Generating realistic test data is painful.

| Approach | Pain |
|---|---|
| **`faker.js` / `ffaker`** | Generates random strings. Doesn't understand foreign keys, unique constraints, or column types. You end up writing glue code for every table. |
| **Custom seed scripts** | One-off SQL or scripts that break when the schema changes. 200 lines for a 10-table schema. Unmaintained after the author leaves. |
| **Production snapshots** | GDPR nightmare. Massive data. Doesn't fit in CI. Contains real user PII. |
| **Manual INSERTs** | Works for 5 rows. Not for 500K. Referential integrity is manual and error-prone. |
| **ORM factories** | Tied to one language/framework. Don't handle cross-service schemas. Schema changes break them silently. |

**Seedling's niche:** Reads your schema, auto-generates 90% of columns, lets you override the other 10% with a minimal Go DSL, and produces millions of referentially-integer, realistic rows. Schema change? Re-run and it adapts.

---

## Core Philosophy

1. **Schema is the source of truth.** Don't repeat yourself. Seedling reads FK constraints, types, defaults, and NOT NULL from your database.
2. **Sensible defaults, easy overrides.** 90% of columns auto-generated. Override only what matters.
3. **Referential integrity always.** Foreign keys resolved automatically. Circular dependencies handled.
4. **Deterministic & reproducible.** Same seed → same data. CI tests never flake.
5. **Fast by default.** Streams inserts. No in-memory accumulation of millions of rows.

---

## Architecture Overview

```
┌──────────────┐      ┌──────────────────────────────────────┐
│  Postgres /  │      │              Seedling                 │
│  MySQL DB    │◄─────│                                      │
│              │      │  ┌──────────┐   ┌──────────────────┐ │
│  Schema:     │      │  │  Schema  │   │ Generator DSL    │ │
│  - tables    │      │  │ Introsp. │   │ (Go-based over-  │ │
│  - columns   │      │  │          │   │  rides per table)│ │
│  - FKs       │      │  └────┬─────┘   └────────┬─────────┘ │
│  - types     │      │       │                   │          │
│  - defaults  │      │       ▼                   ▼          │
│  - indexes   │      │  ┌──────────────────────────────┐    │
│              │      │  │       Plan Builder           │    │
│              │──output│  │  - Topological sort (FKs)   │    │
│  Seed Data   │      │  │  - Column auto-generators    │    │
│  (SQL dump,  │      │  │  - Override merge            │    │
│   direct DB, │      │  │  - Batch planning            │    │
│   CSV, JSON) │      │  └──────────────┬───────────────┘    │
│              │      │                 │                    │
│              │      │                 ▼                    │
│              │      │  ┌──────────────────────────────┐    │
│              │      │  │      Stream Generator        │    │
│              │      │  │  - Per-row generation        │    │
│              │      │  │  - FK pool management        │    │
│              │      │  │  - Unique constraint tracking│    │
│              │      │  │  - Batch writer              │    │
│              │      │  └──────────────────────────────┘    │
└──────────────┘      └──────────────────────────────────────┘
```

### Components

- **Schema Introspector** — Reads information_schema / pg_catalog. Extracts tables, columns, types, FKs, UNIQUE, NOT NULL, CHECK, defaults, sequences.
- **Generator DSL** — Go code defining overrides per table. Compiled to a shared lib or run as a plugin.
- **Plan Builder** — Topologically sorts tables by FK dependency. Assigns column generators. Plans batch sizes.
- **Stream Generator** — Generates rows in dependency order. Tracks unique values. Manages FK pools. Streams to writer.
- **Writers** — SQL dump file, direct DB insert (batched COPY/INSERT), CSV, JSON Lines.

---

## Quick Start

```bash
# Install
go install github.com/user/seedling@latest

# Introspect your schema
seedling introspect --db postgres://localhost:5432/mydb --output schema.yaml

# This generates schema.yaml with all tables, columns, and auto-detected generators.
# Edit it to override domain-specific columns.

# Generate 10K rows to SQL file
seedling generate --count 10000 --output seed.sql

# Generate 500K rows directly into staging DB
seedling generate --count 500000 --db postgres://localhost:5432/staging

# Generate with a seed for reproducibility
seedling generate --count 50000 --seed 42 --output seed.sql

# Watch schema and regenerate on change
seedling watch --db postgres://localhost:5432/mydb --output seed.sql
```

---

## Generator DSL

The core of Seedling is the generator configuration. You write Go code that defines how each table's columns are generated. Only override columns that need domain-specific logic — everything else is auto-generated from schema introspection.

### Basic Table Definition

```go
// generators/users.go
package generators

import "github.com/user/seedling/generator"

var Users = generator.Table("users").
    // Auto-generated columns (from schema): id (serial), created_at (timestamptz), updated_at (timestamptz)
    // These need no config — Seedling handles serial PKs and timestamp defaults automatically.

    // Override: email with a custom domain
    Column("email", generator.Email(generator.Domain("mycompany.com"))).

    // Override: role with weighted distribution (5% admin, 95% user)
    Column("role", generator.WeightedChoice(
        generator.Weight("admin", 0.05),
        generator.Weight("user", 0.95),
    )).

    // Override: status with distribution
    Column("status", generator.WeightedChoice(
        generator.Weight("active", 0.80),
        generator.Weight("inactive", 0.15),
        generator.Weight("suspended", 0.05),
    )).

    // Override: signup_date with normal distribution around a mean
    Column("signup_date", generator.NormalDist("2025-01-01", 90, "day")).

    // Custom generator function for business logic
    Column("credit_score", generator.Func(func(ctx generator.Context) any {
        // Credit score depends on user tier
        tier := ctx.Column("tier").(string)
        switch tier {
        case "premium":
            return generator.Range(700, 850)
        case "standard":
            return generator.Range(600, 750)
        default:
            return generator.Range(300, 600)
        }
    }))
```

### Relationships (Foreign Keys)

Seedling automatically resolves foreign keys. It generates parent rows first, then uses their PKs for child rows.

```go
// generators/orders.go
var Orders = generator.Table("orders").
    // FK: users.id → orders.user_id (auto-resolved from schema)
    // FK: Nullable? Seedling generates some NULLs by default for nullable FKs.

    // Override user_id distribution: 80% of orders from 20% of users (Pareto)
    Column("user_id", generator.FKDistribution(
        generator.Strategy("pareto", 0.8, 0.2),
    )).

    // total_cents: log-normal distribution (most orders small, few large)
    Column("total_cents", generator.LogNormalDist(10.0, 1.5)).
    // cast to int

    // status with realistic ecommerce distribution
    Column("status", generator.WeightedChoice(
        generator.Weight("delivered", 0.65),
        generator.Weight("shipped", 0.15),
        generator.Weight("processing", 0.10),
        generator.Weight("cancelled", 0.07),
        generator.Weight("returned", 0.03),
    ))
```

### Dependent Columns

Columns whose values depend on other columns in the same row:

```go
// generators/order_items.go
var OrderItems = generator.Table("order_items").
    Column("unit_price_cents", generator.Range(99, 19999)).
    Column("quantity", generator.WeightedChoice(
        generator.Weight(1, 0.60),
        generator.Weight(2, 0.25),
        generator.Weight(3, 0.10),
        generator.Weight(4, 0.04),
        generator.Weight(5, 0.01),
    )).
    // Computed: line_total depends on unit_price and quantity
    Column("line_total_cents", generator.Computed(func(ctx generator.Context) any {
        price := ctx.Column("unit_price_cents").(int)
        qty := ctx.Column("quantity").(int)
        return price * qty
    }))
```

### Many-to-Many (Join Tables)

```go
// generators/product_tags.go
var ProductTags = generator.Table("product_tags").
    // Auto-FK to products and tags tables.
    // By default, 30% of products get 1-5 tags.
    Column("product_id", generator.FKDistribution("random-subset", generator.Range(1, 5))).
    Column("tag_id", generator.FKDistribution("paired", "product_id"))
```

### Self-Referencing FK (e.g., parent_id)

```go
// generators/categories.go
var Categories = generator.Table("categories").
    // parent_id is self-referencing. Seedling generates in passes:
    // Pass 1: root nodes (parent_id IS NULL)
    // Pass 2: children referencing pass-1 nodes, etc.
    Column("name", generator.Categorical(
        "Electronics", "Clothing", "Home", "Books", "Sports",
    )).
    Column("parent_id", generator.SelfFK(generator.MaxDepth(3)))
```

### JSON / JSONB Columns

```go
// generators/audit_logs.go
var AuditLogs = generator.Table("audit_logs").
    Column("metadata", generator.JSON(generator.Map{
        "ip":      generator.IPv4(),
        "agent":   generator.UserAgent(),
        "geo": generator.Map{
            "country": generator.Categorical("US", "GB", "NG", "KE", "IN"),
            "city":    generator.City(),
        },
    }))
```

---

## Built-in Generator Functions

### Identity & Strings

| Generator | Description |
|---|---|
| `Email(domain)` | Realistic email addresses |
| `Name()` | Full name |
| `FirstName()` | First name |
| `LastName()` | Last name |
| `Username()` | Unique username |
| `Phone()` | Phone number (configurable country) |
| `UUID()` | UUID v4 or v7 |
| `ULID()` | Sortable unique ID |
| `String(length)` | Random alphanumeric |
| `Lorem(min, max)` | Lorem ipsum text of variable length |
| `Categorical(...values)` | Pick from a list |
| `Regex(pattern)` | Generate strings matching a regex |

### Numbers

| Generator | Description |
|---|---|
| `Range(min, max)` | Uniform random in range |
| `NormalDist(mean, stddev)` | Normal/Gaussian distribution |
| `LogNormalDist(mu, sigma)` | Log-normal distribution |
| `ExponentialDist(lambda)` | Exponential distribution |
| `Sequence(start, step)` | Sequential values |
| `Constant(val)` | Fixed value |
| `WeightedChoice(weights...)` | Discrete weighted distribution |
| `Percent()` | 0.00 to 100.00 |

### Dates & Times

| Generator | Description |
|---|---|
| `Now()` | Current timestamp |
| `DateRange(from, to)` | Uniform random date in range |
| `NormalDist(mean, stddev, unit)` | Normally distributed dates around mean |
| `TimeAgo(maxDuration)` | Random time in the past |
| `Sequence(start, stepDuration)` | Sequential timestamps |
| `BusinessDays(from, to)` | Random business day |

### Network & Web

| Generator | Description |
|---|---|
| `IPv4()` | Random public or private IPv4 |
| `IPv6()` | Random IPv6 |
| `MAC()` | MAC address |
| `URL()` | Random URL |
| `Domain()` | Random domain name |
| `UserAgent()` | Realistic user agent string |
| `MIME()` | MIME type |

### Geography

| Generator | Description |
|---|---|
| `Latitude()` | -90 to 90 |
| `Longitude()` | -180 to 180 |
| `Country()` | Country name |
| `CountryCode()` | ISO 3166-1 alpha-2 |
| `City()` | City name |
| `Address()` | Street address |
| `PostalCode()` | Postal/zip code |

### Finance & Business

| Generator | Description |
|---|---|
| `Currency()` | Currency code (USD, EUR, etc.) |
| `Amount(min, max, currency)` | Monetary amount |
| `IBAN()` | IBAN number |
| `CreditCard()` | Valid-formatted CC number |
| `Company()` | Company name |
| `JobTitle()` | Job title |
| `Industry()` | Industry category |

### Files & Media

| Generator | Description |
|---|---|
| `FileName(ext)` | Random filename with extension |
| `MIME()` | MIME type |
| `ImageURL(width, height)` | Placeholder image URL |

---

## Foreign Key Strategies

Seedling offers multiple strategies for how FK values are distributed across child rows.

| Strategy | Description |
|---|---|
| `uniform` (default) | Each parent referenced roughly equally |
| `pareto(ratio, pct_parents)` | 80/20 rule: 80% of children reference 20% of parents |
| `random-subset(min, max)` | Each child references N parents (for many-to-many) |
| `paired(field)` | Value tracks another FK column (for join table pairing) |
| `sequential` | Round-robin through parent IDs |
| `weighted(weights)` | Custom weights per parent |
| `exclusive` | Each parent referenced at most once (1:1) |

---

## Circular Dependencies

Some schemas have circular FK references (e.g., `employees.manager_id → employees.id`). Seedling handles these with:

1. **NULL-able columns first** — Seedling generates rows with NULL in the circular FK column in pass 1.
2. **Deferred updates** — In pass 2, it fills in the FK values from already-generated rows.
3. **Deferred constraint mode** — For non-nullable circular FKs, Seedling uses `SET CONSTRAINTS ALL DEFERRED` within a transaction.

```go
// Seedling auto-detects circular dependencies.
// You can also hint:
var Employees = generator.Table("employees").
    Column("manager_id", generator.SelfFK(
        generator.MaxDepth(5),          // max management chain depth
        generator.NullRootProbability(0.10), // 10% have no manager (roots)
    ))
```

---

## Schema Introspection

```bash
seedling introspect --db postgres://localhost:5432/mydb
```

Outputs `schema.yaml`:

```yaml
tables:
  - name: users
    columns:
      - name: id
        type: integer
        generator: serial          # auto-detected from SERIAL/BIGSERIAL
      - name: email
        type: varchar(255)
        nullable: false
        unique: true
        generator: email           # auto-detected from column name
      - name: name
        type: varchar(100)
        generator: full_name       # auto-detected from column name
      - name: role
        type: varchar(20)
        default: "user"
        generator: default_value
      - name: created_at
        type: timestamptz
        default: now()
        generator: default_now
    foreign_keys: []

  - name: orders
    columns:
      - name: id
        type: bigint
        generator: serial
      - name: user_id
        type: integer
        nullable: false
        fk: users.id               # auto-detected FK
        generator: fk_uniform
      - name: total_cents
        type: integer
        generator: range_0_100000  # auto-detected from type + name hints
      - name: status
        type: varchar(20)
        generator: categorical     # flag: user should override
      - name: order_date
        type: timestamptz
        default: now()
        generator: default_now
    foreign_keys:
      - column: user_id
        references: users.id
```

You can edit this YAML directly, or write Go generators for richer logic.

---

## Output Formats

| Format | Command | Use Case |
|---|---|---|
| **SQL dump** | `--output seed.sql` | CI pipelines, staging refresh, load testing prep |
| **Direct DB insert** | `--db postgres://...` | One-shot staging population, demo environments |
| **CSV** | `--output data/ --format csv` | Data analysis, import to spreadsheets |
| **JSON Lines** | `--output data/ --format jsonl` | NoSQL-style ingestion, event streams |
| **Parquet** | `--output data/ --format parquet` | Data lake, Spark/Athena querying |

### Batch Sizing

```bash
# Batch INSERT with 1000 rows per statement
seedling generate --count 1000000 --batch-size 1000 --db postgres://...

# COPY protocol for max throughput
seedling generate --count 10000000 --copy --db postgres://...
```

---

## Presets

Save named generation configurations for reuse:

```bash
# Save a preset
seedling preset save small-dev --count 500 --seed 42 --db postgres://localhost:5432/dev

# Save a preset
seedling preset save load-test --count 10000000 --seed 123 --output load_test.sql

# List presets
seedling preset list

# small-dev    500 rows    postgres://localhost:5432/dev    seed=42
# load-test    10M rows    load_test.sql                    seed=123
# demo         2K rows     postgres://staging:5432/db        seed=77

# Use a preset
seedling generate --preset demo
```

---

## CLI Reference

```bash
# Schema introspection
seedling introspect --db <DSN> [--output schema.yaml] [--format yaml|json]

# Generate data
seedling generate
    --count <N>                     # Number of rows per root table
    --seed <int>                    # Deterministic seed (default: random)
    --db <DSN>                      # Direct database insert
    --output <file|dir>             # File output (SQL, CSV, JSONL, Parquet)
    --format <sql|csv|jsonl|parquet>
    --batch-size <N>               # Rows per batch
    --copy                         # Use COPY protocol (Postgres only)
    --truncate                     # TRUNCATE tables before inserting
    --generators <dir>             # Path to Go generator files
    --config <file>                # Path to seed.yaml config
    --preset <name>                # Use a saved preset
    --dry-run                      # Print generation plan without writing
    --verbose                      # Detailed progress output

# Watch mode
seedling watch
    --db <DSN>
    --generators <dir>
    --output <file>
    [--debounce 2s]

# Presets
seedling preset list
seedling preset save <name> [options]
seedling preset delete <name>
seedling preset show <name>

# Validate
seedling validate [--generators <dir>] [--db <DSN>]
```

---

## Config File (seedling.yaml)

```yaml
# seedling.yaml
version: "1"

database:
  dsn: "${DATABASE_URL}"

generators:
  dir: ./generators
  package: generators

output:
  format: sql
  file: ./seed/seed.sql
  batch_size: 1000
  use_copy: true

generation:
  count: 50000
  seed: 42
  truncate: false

tables:
  users:
    count: 10000         # Override: generate exactly 10K users
  orders:
    count: [5, 15]       # Per-user order count: random 5-15 per user
  audit_logs:
    enabled: false       # Skip this table entirely

schemas:
  - public
  - audit
  # - archive           # Skip archive schema
```

---

## CI Integration

```yaml
# GitHub Actions
- name: Seed test database
  run: |
    seedling generate \
      --preset ci-test \
      --db postgres://postgres:postgres@localhost:5432/testdb

- name: Run tests
  run: go test ./...
```

```yaml
# Docker Compose
services:
  db:
    image: postgres:16
    environment:
      POSTGRES_DB: testdb
    healthcheck:
      test: pg_isready -U postgres

  seed:
    image: ghcr.io/user/seedling:latest
    depends_on:
      db:
        condition: service_healthy
    command: generate --config seedling.yaml --db postgres://postgres:postgres@db:5432/testdb
```

---

## Schema Change Handling

Seedling watches your schema and adapts:

| Schema Change | Seedling Behavior |
|---|---|
| **ADD COLUMN** | Auto-generate values for that column with default generator. Existing presets still work. |
| **DROP COLUMN** | Skip the column. Existing generators that reference it will error on `validate`. |
| **ADD TABLE** | Auto-include in generation. Warn if no generator defined (uses all defaults). |
| **DROP TABLE** | Skip the table. Generators referencing dropped FKs from this table will warn. |
| **RENAME COLUMN** | Generator references break. `validate` catches this. Update generator code. |
| **CHANGE TYPE** | Generator output type-checked at runtime. Mismatch → error row logged. |
| **ADD FK** | Auto-included in topological sort. |
| **DROP FK** | Column reverts to standard generator behavior. |

---

## Determinism Guarantees

```
Same schema + Same generators + Same seed + Same count = Same output every time.
```

- Random number generator seeded per-column, per-row, from the master seed. Predictable traversal.
- FK ordering is stable (rows generated in sorted PK order of parent tables).
- Strings use a deterministic PRNG (ChaCha8).
- `-race` tested. No shared mutable state across goroutines during generation.

---

## Performance

| Scale | Target Time | Memory |
|---|---|---|
| 10K rows / 5 tables | <1 second | <50MB |
| 500K rows / 20 tables | <15 seconds | <100MB |
| 10M rows / 30 tables | <2 minutes | <200MB |
| 100M rows / 50 tables | <20 minutes | <512MB |

Key techniques:
- Streaming generation (no row accumulation in memory)
- Batched COPY protocol (Postgres) for max insert throughput
- FK pool pre-computation in sorted order
- Parallel generation across independent table subgraphs
- Unique constraint tracking via Bloom filter (probabilistic) + hash set (exact) when near capacity

---

## Development Phases

### Phase 1: Schema Introspection + Basic Generation
- Postgres schema reading (information_schema / pg_catalog)
- Auto-detection of column types → generators
- SQL dump output (INSERT statements)
- Topological sort of tables by FK dependency
- Basic FK handling (UNIFORM strategy)

**Goal:** Point at a Postgres schema, generate 100 rows per table to a `.sql` file.

### Phase 2: Generator DSL + Direct DB Insert
- Go-based generator DSL and compilation
- Direct DB insert (batched INSERT, COPY protocol)
- Column generators: NormalDist, WeightedChoice, Range, Sequence
- Unique constraint tracking
- Deterministic seeding

### Phase 3: Advanced FK + Constraints
- FK strategies: pareto, random-subset, paired, sequential, exclusive
- Circular dependency handling (NULL-passthrough + deferred constraints)
- Self-referencing FK (parent_id)
- Many-to-many join table generation
- CHECK constraint awareness

### Phase 4: MySQL + More Outputs
- MySQL schema introspection
- MySQL direct insert + LOAD DATA INFILE
- CSV output
- JSON Lines output
- Parquet output (via Go Parquet library)

### Phase 5: CLI Polish + Presets + Watch
- Full CLI (cobra)
- Preset management
- Watch mode (fsnotify on schema + generators)
- Config file (seedling.yaml)
- Progress bars and verbose output
- `validate` command

### Phase 6: Distribution + Docs
- Schema-based auto-generator improvement (better heuristics)
- Custom distribution functions
- Conditional generation (if column X = val, then column Y = ...)
- Comprehensive docs site
- Docker image
- CI integration guides (GitHub Actions, GitLab CI, CircleCI)

---

## Learning Opportunities

1. **Relational algebra** — Topological sorting of FK dependency graphs, constraint satisfaction
2. **Go code generation & plugin systems** — Compiling user Go code into runnable generators (Go plugin package, or wasm)
3. **Database driver internals** — pgx advanced usage, COPY protocol, MySQL LOAD DATA, batch INSERT optimization
4. **Deterministic PRNG** — Seeded generation with per-column/per-row derived seeds, ChaCha8 internals
5. **Streaming data pipelines** — Generating millions of rows without memory accumulation
6. **Schema introspection** — Deep dive into information_schema, pg_catalog, SHOW CREATE TABLE
7. **Statistical distributions** — Implementing normal, log-normal, exponential, Pareto distributions in Go
8. **Bloom filters** — Probabilistic unique constraint tracking at scale
9. **Foreign key cardinality modeling** — Realistic data distribution patterns (Pareto, long-tail)
10. **Performance optimization** — Go concurrency for parallel generation, batching, allocation reduction

---

## Why This Project?

- **Universal pain point.** Every developer who's ever seeded a database has felt this pain.
- **Technically interesting.** FK graph topology, streaming generation, distribution modeling — real CS problems.
- **Go-native.** Leverages Go's concurrency for parallel generation and streaming.
- **Immediate utility.** You'll use this on your very next project. So will everyone you know.
- **Manageable scope.** Core value (schema-aware generation) is achievable in Phase 1-2.
- **Open-source with clear revenue path.** Open-source core + cloud-hosted version with team management, schema history, and collaboration features.
