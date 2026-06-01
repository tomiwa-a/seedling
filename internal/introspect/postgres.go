package introspect

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tomiwa-a/seedling/pkg/schema"
)

type PostgresIntrospector struct {
	pool    *pgxpool.Pool
	schemas []string
}

type PostgresOption func(*PostgresIntrospector)

func WithSchemas(schemas ...string) PostgresOption {
	return func(pi *PostgresIntrospector) {
		pi.schemas = schemas
	}
}

func NewPostgresIntrospector(ctx context.Context, dsn string, opts ...PostgresOption) (*PostgresIntrospector, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	pi := &PostgresIntrospector{
		pool:    pool,
		schemas: []string{"public"},
	}

	for _, opt := range opts {
		opt(pi)
	}

	return pi, nil
}

func (pi *PostgresIntrospector) Introspect(ctx context.Context, dsn string) (*schema.Schema, error) {
	return pi.introspect(ctx)
}

func (pi *PostgresIntrospector) introspect(ctx context.Context) (*schema.Schema, error) {
	tables, err := pi.extractTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("extract tables: %w", err)
	}

	return &schema.Schema{
		Name:   pi.schemas[0],
		Tables: tables,
	}, nil
}

func (pi *PostgresIntrospector) extractTables(ctx context.Context) ([]schema.Table, error) {
	rows, err := pi.pool.Query(ctx, `
		SELECT table_name, table_type
		FROM information_schema.tables
		WHERE table_schema = ANY($1)
		  AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`, pi.schemas)
	if err != nil {
		return nil, fmt.Errorf("query tables: %w", err)
	}
	defer rows.Close()

	var tables []schema.Table
	for rows.Next() {
		var tableName, tableType string
		if err := rows.Scan(&tableName, &tableType); err != nil {
			return nil, fmt.Errorf("scan table row: %w", err)
		}

		columns, err := pi.extractColumns(ctx, pi.schemas[0], tableName)
		if err != nil {
			return nil, fmt.Errorf("extract columns for %s: %w", tableName, err)
		}

		tables = append(tables, schema.Table{
			Name:       tableName,
			SchemaName: pi.schemas[0],
			Columns:    columns,
		})
	}

	return tables, rows.Err()
}

type rawColumn struct {
	Name              string
	DataType          string
	IsNullable        string
	ColumnDefault     *string
	CharacterMaxLen   *int
	NumericPrecision  *int
	NumericScale      *int
	UDTName           string
}

func (pi *PostgresIntrospector) extractColumns(ctx context.Context, schemaName, tableName string) ([]schema.Column, error) {
	rows, err := pi.pool.Query(ctx, `
		SELECT
			column_name,
			data_type,
			is_nullable,
			column_default,
			character_maximum_length,
			numeric_precision,
			numeric_scale,
			COALESCE(udt_name, '')
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
		ORDER BY ordinal_position
	`, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("query columns: %w", err)
	}
	defer rows.Close()

	var columns []schema.Column
	for rows.Next() {
		var rc rawColumn
		if err := rows.Scan(
			&rc.Name,
			&rc.DataType,
			&rc.IsNullable,
			&rc.ColumnDefault,
			&rc.CharacterMaxLen,
			&rc.NumericPrecision,
			&rc.NumericScale,
			&rc.UDTName,
		); err != nil {
			return nil, fmt.Errorf("scan column row: %w", err)
		}

		col := schema.Column{
			Name:     rc.Name,
			RawType:  rc.UDTName,
			Nullable: rc.IsNullable == "YES",
			Default:  rc.ColumnDefault,
		}

		if rc.CharacterMaxLen != nil {
			col.MaxLength = *rc.CharacterMaxLen
		}
		if rc.NumericPrecision != nil {
			col.NumericPrec = *rc.NumericPrecision
		}
		if rc.NumericScale != nil {
			col.NumericScale = *rc.NumericScale
		}

		columns = append(columns, col)
	}

	return columns, rows.Err()
}

func (pi *PostgresIntrospector) Close() {
	if pi.pool != nil {
		pi.pool.Close()
	}
}


