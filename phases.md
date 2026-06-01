# Seedling — 10-Phase Development Plan

---

## Phase 1: Project Scaffolding & Core Architecture

**Goal:** Establish the Go module, CLI skeleton, core data types, and public interfaces. Nothing generates data yet.

- [x] Initialize Go module (`go mod init github.com/tomiwa-a/seedling`), set up project directory structure (`cmd/`, `pkg/`, `internal/`)
- [x] Define core data types: `Table`, `Column`, `ForeignKey`, `Constraint`, `Schema` structs
- [x] Define public interfaces: `Generator`, `Writer`, `Introspector`, `PlanBuilder`, `StreamGenerator`
- [x] Scaffold CLI entry point with Cobra (`cmd/seedling/main.go`) and stub commands: `introspect`, `generate`, `validate`, `watch`
- [x] Set up CI (Go lint, vet, test on push), README badges, contribution guidelines
- [x] Write unit tests for all core types (serialization, equality, validation)

---

## Phase 2: Postgres Schema Introspection Engine

**Goal:** Read a live Postgres database and produce a structured `Schema` object with all tables, columns, types, FKs, constraints, and defaults.

- [x] Implement `pg_introspector.go` — connect via `pgx`, query `information_schema.tables` + `information_schema.columns`
- [x] Extract column type info: SQL type → Seedling type mapping (e.g., `varchar(255)` → `String`, `serial` → `Serial`)
- [x] Extract foreign keys from `information_schema.table_constraints` + `key_column_usage`
- [x] Extract constraints: `NOT NULL`, `UNIQUE`, `CHECK`, `DEFAULT` values, sequence ownership
- [x] Extract column comments (if any) as generator hints
- [x] Write `seedling introspect` CLI command with `--db` and `--output` flags
- [x] Output introspection result as YAML (`schema.yaml`) and JSON (`schema.json`)

---

## Phase 3: Basic Generation Pipeline

**Goal:** From an introspected schema, generate rows to a SQL dump file. No user overrides yet — all auto-generated.

- [x] Build `PlanBuilder` — topological sort of tables by FK dependency (Kahn's algorithm)
- [x] Implement auto-generator assignment per column type (e.g., `serial` → sequence, `timestamptz` → now(), `varchar` → random string, `email` → email)
- [x] Build `StreamGenerator` — iterate tables in dependency order, generate rows, resolve FK lookups
- [x] Implement `FKPool` — store generated parent PKs, provide random/sequential lookup for child rows
- [x] Write `SqlWriter` — output `INSERT INTO ...` statements with batched rows
- [x] Wire `seedling generate` command — read schema, plan, generate, write SQL
- [x] Deliverable: `seedling introspect --db ... && seedling generate --count 100 --output seed.sql` produces valid SQL

---

## Phase 4: Built-in Generator Library

**Goal:** Ship a comprehensive library of generators for common data types — strings, numbers, dates, geography, finance, networking.

- [x] **Strings:** `Email`, `Name`, `Username`, `Phone`, `UUID`, `ULID`, `String`, `Lorem` — `FirstName`/`LastName` not needed (use `Name` twice), `Categorical`/`Regex` skipped (niche)
- [x] **Numbers:** `RandomInt` (Range), `Serial` (Sequence), `Constant`, `WeightedChoice`, `FloatRange`, `Numeric` — `NormalDist`/`LogNormalDist`/`ExponentialDist`/`Percent` skipped (niche)
- [x] **Dates:** `Now`, `DateGenerator` (DateRange), `TimeAgo`, `BusinessDays`, `Timestamp` — `NormalDist(date)`/`Sequence(date)` skipped (niche)
- [x] **Geography:** `Latitude`, `Longitude`, `Country`, `CountryCode`, `City`, `Address`, `PostalCode`
- [x] **Network/Finance/Media:** `IPv4`, `IPv6`, `MAC`, `URL`, `UserAgent`, `Currency`, `Amount`, `Company`, `JobTitle`, `FileName`, `ImageURL` — `IBAN`/`CreditCard` skipped (niche)
- [x] Write property tests for each generator (output type matches, no panics, bounds respected)
- [x] Split generators into domain files (`strings.go`, `numbers.go`, `dates.go`, `geo.go`, `network.go`, `finance.go`, `media.go`)
- [x] Fix hints wiring: introspect detects hints → stores in `schema.Column.Hint` → `ResolveGenerators` reads them → correct generator auto-selected

---

## Phase 5: Generator DSL & User Overrides

**Goal:** Allow users to override auto-generated columns with a Go-based DSL. Compile user generators alongside Seedling.

- [ ] Design `generator.Table(name)` DSL — chainable `.Column(name, gen)` calls returning a `TableConfig`
- [ ] Implement DSL merge logic: user overrides merged on top of auto-detected generators per column
- [ ] Build dependency injection for generators referencing other columns in the same row (`generator.Context`, `Computed`)
- [ ] Implement `generator.Func(func(ctx Context) any)` — custom inline generator functions
- [ ] Add `seedling generate --generators ./generators` flag to load user generator packages
- [ ] Write `seedling validate` command — validate user generators against schema (detect stale column refs, type mismatches)
- [ ] Document DSL with examples for common patterns

---

## Phase 6: Deterministic Seeding & Unique Constraints

**Goal:** Fully reproducible output. Same schema + generators + seed + count = identical data every run. Enforce UNIQUE constraints at generation time.

- [x] Implement ChaCha8-based deterministic PRNG — master seed derives per-column, per-row sub-seeds
- [x] Seed all generators from the deterministic PRNG (remove all `crypto/rand` usage)
- [x] Build `UniqueTracker` — track unique column values, detect collisions, retry with new values
- [ ] Implement Bloom filter for low-memory uniqueness checks at large scale (fallback to exact hash set)
- [x] Add `--seed <int>` flag to `seedling generate`, default to random if omitted
- [x] Write determinism test: run twice with same seed, diff output — must be identical
- [x] Add `-race` test for shared mutable state (no data races during concurrent generation)

---

## Phase 7: Advanced FK Strategies & Circular Dependencies

**Goal:** Handle real-world FK patterns — Pareto distributions, many-to-many join tables, self-referencing hierarchies, circular FK loops.

- [x] Implement FK strategies: `uniform`, `sequential`, `weighted`, `exclusive`
- [x] Build `SelfFK` generator for self-referencing tables (e.g., `employees.manager_id → employees.id`): multi-pass generation with configurable max depth
- [x] Detect circular FK dependencies in `PlanBuilder` — split into pass groups
- [x] Handle nullable circular columns: generate NULL in pass 1, fill in pass 2
- [x] Handle non-nullable circular columns: use `SET CONSTRAINTS ALL DEFERRED` within a transaction
- [ ] Add DSL methods: `generator.FKDistribution(strategy, ...args)`, `generator.SelfFK(opts...)`
- [x] Write integration tests with circular FK schemas

---

## Phase 8: Direct Database Insert & Performance

**Goal:** Generate directly into a live database with batched INSERTs and COPY protocol for maximum throughput.

- [x] Implement `DbWriter` — connect to Postgres, batch INSERT with configurable batch size (`--batch-size`)
- [x] Implement `CopyWriter` — use pgx COPY protocol for row streaming (max performance)
- [x] Add `--truncate` flag — `TRUNCATE ... CASCADE` tables before generation
- [x] Add `--dry-run` flag — print generation plan (row counts per table, FK order, estimated size) without writing
- [x] Add `--verbose` progress output (per-table progress, rows/sec, ETA)
- [x] Performance benchmark: 10K/500K/10M rows, measure throughput, memory, and CPU profiles
- [x] Parallel generation: identify independent table subgraphs and generate concurrently

---

## Phase 9: MySQL Support & Multiple Output Formats

**Goal:** Support MySQL/MariaDB schemas and output to CSV, JSON Lines, and Parquet.

- [ ] Implement `mysql_introspector.go` — connect via `go-sql-driver/mysql`, query `information_schema`
- [ ] Map MySQL types to Seedling types (e.g., `TINYINT(1)` → Boolean, `ENUM('a','b')` → Categorical)
- [ ] Implement `MysqlWriter` with `LOAD DATA INFILE` for high-throughput inserts
- [ ] Implement `CsvWriter` — one file per table, configurable delimiter and quoting
- [ ] Implement `JsonLinesWriter` — one JSON object per row, streaming per table
- [ ] Implement `ParquetWriter` — columnar output using `github.com/xitongsys/parquet-go`
- [ ] Add `--format <sql|csv|jsonl|parquet>` flag to `seedling generate`

---

## Phase 10: CLI Polish, Presets, Watch Mode & Distribution

**Goal:** Production-ready CLI with config files, preset management, file watching, and distribution artifacts (Docker, CI templates, docs).

- [ ] Implement config file support (`seedling.yaml`) — all CLI flags expressible as YAML, with `${ENV_VAR}` interpolation
- [ ] Build preset system: `seedling preset save/list/show/delete` — presets stored in `~/.config/seedling/presets.yaml`
- [ ] Implement `seedling watch` — fsnotify on schema.yaml and generators dir, auto-regenerate on change with configurable debounce
- [ ] Add progress bars (`charmbracelet/bubbletea` or `cheggaaa/pb`) for generation and introspection
- [ ] Write comprehensive CI integration guides: GitHub Actions, GitLab CI, CircleCI, Docker Compose
- [ ] Build Docker image (multi-arch: amd64, arm64), publish to GHCR
- [ ] Set up release automation (goreleaser) — binaries for Linux, macOS, Windows
- [ ] Ship phase-end demo: point at real schema, generate 1M rows, verify determinism, time it

---

## Summary

| Phase | Deliverable | Dependencies |
|-------|-------------|-------------|
| 1 | Go project skeleton, CLI, core types | None |
| 2 | `seedling introspect` → schema.yaml | Phase 1 |
| 3 | `seedling generate` → seed.sql (auto) | Phase 2 |
| 4 | 40+ built-in generators | Phase 3 |
| 5 | Go generator DSL + user overrides | Phase 4 |
| 6 | Deterministic output + UNIQUE enforcement | Phase 5 |
| 7 | FK strategies, circular deps, self-FK | Phase 6 |
| 8 | Direct DB insert, COPY, parallel gen | Phase 7 |
| 9 | MySQL, CSV, JSONL, Parquet | Phase 8 |
| 10 | Config, presets, watch, Docker, CI, docs | Phase 9 |
