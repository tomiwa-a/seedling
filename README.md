# Seedling — Relational Test Data Factory

[![CI](https://github.com/tomiwa-a/seedling/actions/workflows/ci.yml/badge.svg)](https://github.com/tomiwa-a/seedling/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.25-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

A schema-aware test data generator. Point it at a Postgres or MySQL schema, and it produces referentially-integer, realistic rows that respect every foreign key, unique constraint, and column type.

## Quick Start

```bash
# Introspect your database
seedling introspect --db postgres://localhost:5432/mydb --output schema.yaml

# Generate 10K rows to a SQL file
seedling generate --count 10000 --output seed.sql

# Generate 500K rows directly into staging (auto-detects MySQL vs Postgres)
seedling generate --count 500000 --db postgres://localhost:5432/staging

# Deterministic generation (same seed = same output)
seedling generate --count 50000 --seed 42 --output seed.sql
```

## How It Works

1. **Introspect** — Reads `information_schema` from Postgres or MySQL, producing a `schema.yaml` with all tables, columns, types, FKs, and constraints.
2. **Auto-detect generators** — Column types and names are mapped to built-in generators (e.g., `email` columns get email generators, `serial` PKs get sequences, FK columns get FK lookup generators).
3. **Override (optional)** — Provide a YAML config to override specific columns with different generators.
4. **Generate** — Rows are streamed in dependency order (topological sort by FK), respecting unique constraints and circular dependencies.

## CLI Reference

### `seedling introspect`

```bash
seedling introspect --db <DSN> [--output schema.yaml] [--format yaml|json]
```

Reads a live database schema. Auto-detects Postgres vs MySQL from the DSN.

### `seedling generate`

```bash
seedling generate
    --count <N>                     Rows per root table (default: 100)
    --seed <int>                    Deterministic seed (0 = random)
    --schema <file>                 Schema file (default: schema.yaml)
    --output <file|dir>             Output destination
    --format <sql|csv|jsonl|parquet> Output format (default: sql)
    --batch-size <N>                Rows per batch (default: 1000)
    --db <DSN>                      Direct database insert
    --copy                          Use COPY protocol (Postgres only)
    --truncate                      TRUNCATE tables before inserting
    --generators <file>             YAML generator overrides
    --config <file>                 seedling.yaml config
    --dry-run                       Print plan without generating
    --parallel                      Generate tables in parallel (breaks determinism)
    --verbose                       Progress bars and summary
```

### `seedling watch`

```bash
seedling watch --schema schema.yaml [--generators generators.yaml] [--debounce 2s]
```

Watches schema/generators files for changes and auto-regenerates.

### `seedling validate`

```bash
seedling validate [--generators <file>] [--db <DSN>]
```

Stub — validates generator configs against schema (not yet implemented).

## Generator Overrides

Override auto-detected generators with a YAML file:

```yaml
users:
  email:
    generator: email
  role:
    generator: weighted_choice
    params:
      choices: {"admin": 5, "user": 95}
  created_at:
    generator: now
  internal_code:
    disabled: true
```

Use with: `seedling generate --generators overrides.yaml`

### Built-in Generators

| Name | Description |
|---|---|
| `email` | Realistic email (first.last_N@example.com) |
| `name` | Full name |
| `phone` | Phone number |
| `uuid` | UUID |
| `bool` | Random boolean |
| `lorem` | Lorem ipsum text (params: `min_words`, `max_words`) |
| `constant` | Fixed value (params: `value`) |

## Config File (seedling.yaml)

All CLI flags can be expressed as YAML. Supports `${ENV_VAR}` interpolation.

```yaml
database:
  dsn: "${DATABASE_URL}"

output:
  format: sql
  file: seed.sql
  batch_size: 1000
  use_copy: true

generation:
  count: 50000
  seed: 42
  verbose: true

schema:
  file: schema.yaml
```

Use with: `seedling generate --config seedling.yaml`

## Output Formats

| Format | Command | Use Case |
|---|---|---|
| SQL dump | `--output seed.sql` | CI pipelines, staging refresh |
| Direct DB insert | `--db postgres://...` | One-shot staging population |
| COPY (Postgres) | `--db ... --copy` | Max throughput inserts |
| CSV | `--output data/ --format csv` | Data analysis, spreadsheets |
| JSON Lines | `--output data/ --format jsonl` | Event streams, NoSQL |
| Parquet | `--output data/ --format parquet` | Data lake (TSV placeholder) |

## Determinism

Same schema + generators + seed + count = identical output every time.

```bash
seedling generate --count 50000 --seed 42 --output seed.sql
# Run again — output is byte-identical
```

## Database Support

- **Postgres** — Full introspection, batched INSERT, COPY protocol
- **MySQL/MariaDB** — Full introspection, batched INSERT with type-safe clamping

## Installation

```bash
go install github.com/tomiwa-a/seedling/cmd/seedling@latest
```

Or use Docker:

```bash
docker build -t seedling .
docker run seedling introspect --help
```

## Architecture

```
DB Schema → Introspector → schema.yaml
                                ↓
Generator Overrides → PlanBuilder (topo sort, FK order)
                                ↓
                         StreamGenerator (row generation, FK pools, unique tracking)
                                ↓
                         Writer (SQL/CSV/JSONL/DB/Copy)
```
