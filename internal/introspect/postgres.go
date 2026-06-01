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

		comments, err := pi.extractColumnComments(ctx, pi.schemas[0], tableName)
		if err == nil {
			for i := range columns {
				if c, ok := comments[columns[i].Name]; ok {
					columns[i].Comment = c
				}
			}
		}

		for i := range columns {
			hint := detectGeneratorHint(columns[i])
			if hint == schema.HintAuto {
				hint = hintFromComment(columns[i].Comment)
			}
		}

		fks, err := pi.extractForeignKeys(ctx, pi.schemas[0], tableName)
		if err != nil {
			return nil, fmt.Errorf("extract foreign keys for %s: %w", tableName, err)
		}

		for i := range columns {
			for _, fk := range fks {
				if columns[i].Name == fk.ColumnName {
					columns[i].FKRef = &schema.FKRef{
						Table:  fk.RefTable,
						Column: fk.RefColumn,
					}
				}
			}
		}

		constraints, uniqueCols, err := pi.extractConstraints(ctx, pi.schemas[0], tableName)
		if err != nil {
			return nil, fmt.Errorf("extract constraints for %s: %w", tableName, err)
		}

		for i := range columns {
			if uniqueCols[columns[i].Name] {
				columns[i].Unique = true
			}
		}

		tables = append(tables, schema.Table{
			Name:        tableName,
			SchemaName:  pi.schemas[0],
			Columns:     columns,
			ForeignKeys: fks,
			Constraints: constraints,
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

		colType := mapColumnType(rc.UDTName, rc.DataType)

		col := schema.Column{
			Name:     rc.Name,
			Type:     colType,
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

		if colType == schema.TypeEnum {
			values, err := pi.extractEnumValues(ctx, schemaName, rc.UDTName)
			if err == nil {
				col.EnumValues = values
			}
		}

		columns = append(columns, col)
	}

	return columns, rows.Err()
}

func (pi *PostgresIntrospector) extractForeignKeys(ctx context.Context, schemaName, tableName string) ([]schema.ForeignKey, error) {
	rows, err := pi.pool.Query(ctx, `
		SELECT
			kcu.column_name,
			ccu.table_name AS ref_table,
			ccu.column_name AS ref_column,
			COALESCE(rc.update_rule, '') AS update_rule,
			COALESCE(rc.delete_rule, '') AS delete_rule,
			tc.constraint_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		JOIN information_schema.constraint_column_usage ccu
			ON ccu.constraint_name = tc.constraint_name
			AND ccu.constraint_schema = tc.table_schema
		LEFT JOIN information_schema.referential_constraints rc
			ON rc.constraint_name = tc.constraint_name
			AND rc.constraint_schema = tc.table_schema
		WHERE tc.constraint_type = 'FOREIGN KEY'
			AND tc.table_schema = $1
			AND tc.table_name = $2
		ORDER BY tc.constraint_name, kcu.ordinal_position
	`, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("query foreign keys: %w", err)
	}
	defer rows.Close()

	var fks []schema.ForeignKey
	for rows.Next() {
		var fk schema.ForeignKey
		if err := rows.Scan(
			&fk.ColumnName,
			&fk.RefTable,
			&fk.RefColumn,
			&fk.OnUpdate,
			&fk.OnDelete,
			&fk.ConstraintName,
		); err != nil {
			return nil, fmt.Errorf("scan foreign key row: %w", err)
		}
		fks = append(fks, fk)
	}

	return fks, rows.Err()
}

func (pi *PostgresIntrospector) extractColumnComments(ctx context.Context, schemaName, tableName string) (map[string]string, error) {
	rows, err := pi.pool.Query(ctx, `
		SELECT
			a.attname AS column_name,
			pg_catalog.col_description(a.attrelid, a.attnum) AS comment
		FROM pg_catalog.pg_attribute a
		JOIN pg_catalog.pg_class c ON a.attrelid = c.oid
		JOIN pg_catalog.pg_namespace n ON c.relnamespace = n.oid
		WHERE n.nspname = $1
			AND c.relname = $2
			AND a.attnum > 0
			AND NOT a.attisdropped
	`, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("query column comments: %w", err)
	}
	defer rows.Close()

	comments := make(map[string]string)
	for rows.Next() {
		var colName string
		var comment *string
		if err := rows.Scan(&colName, &comment); err != nil {
			return nil, fmt.Errorf("scan comment row: %w", err)
		}
		if comment != nil {
			comments[colName] = *comment
		}
	}

	return comments, rows.Err()
}

func (pi *PostgresIntrospector) extractConstraints(ctx context.Context, schemaName, tableName string) ([]schema.Constraint, map[string]bool, error) {
	uniqueCols := make(map[string]bool)
	var constraints []schema.Constraint

	rows, err := pi.pool.Query(ctx, `
		SELECT
			tc.constraint_type,
			tc.constraint_name,
			COALESCE(kcu.column_name, '') AS column_name,
			COALESCE(cc.check_clause, '') AS check_clause
		FROM information_schema.table_constraints tc
		LEFT JOIN information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		LEFT JOIN information_schema.check_constraints cc
			ON cc.constraint_name = tc.constraint_name
			AND cc.constraint_schema = tc.table_schema
		WHERE tc.table_schema = $1
			AND tc.table_name = $2
			AND tc.constraint_type IN ('UNIQUE', 'CHECK', 'PRIMARY KEY')
		ORDER BY tc.constraint_name, kcu.ordinal_position
	`, schemaName, tableName)
	if err != nil {
		return nil, nil, fmt.Errorf("query constraints: %w", err)
	}
	defer rows.Close()

	currentConstraint := ""
	var currentCols []string
	var currentType schema.ConstraintType
	var currentName string
	var currentCheck string

	flushConstraint := func() {
		if currentConstraint != "" {
			c := schema.Constraint{
				Type:    currentType,
				Name:    currentName,
				Columns: currentCols,
			}
			if currentType == schema.ConstraintCheck {
				c.Expression = currentCheck
			}
			constraints = append(constraints, c)

			if currentType == schema.ConstraintUnique {
				for _, col := range currentCols {
					uniqueCols[col] = true
				}
			}
		}
	}

	for rows.Next() {
		var constraintType, constraintName, columnName, checkClause string
		if err := rows.Scan(&constraintType, &constraintName, &columnName, &checkClause); err != nil {
			return nil, nil, fmt.Errorf("scan constraint row: %w", err)
		}

		if constraintName != currentConstraint {
			flushConstraint()
			currentConstraint = constraintName
			currentCols = nil
			currentName = constraintName
			currentCheck = checkClause
			switch constraintType {
			case "UNIQUE":
				currentType = schema.ConstraintUnique
			case "CHECK":
				currentType = schema.ConstraintCheck
			case "PRIMARY KEY":
				currentType = schema.ConstraintPrimaryKey
			}
		}

		if columnName != "" {
			currentCols = append(currentCols, columnName)
		}
	}
	flushConstraint()

	return constraints, uniqueCols, rows.Err()
}

func (pi *PostgresIntrospector) Close() {
	if pi.pool != nil {
		pi.pool.Close()
	}
}


