package introspect

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/tomiwa-a/seedling/pkg/schema"
)

type MysqlIntrospector struct {
	db      *sql.DB
	schemas []string
}

type MysqlOption func(*MysqlIntrospector)

func WithMysqlDatabase(db string) MysqlOption {
	return func(mi *MysqlIntrospector) {
		mi.schemas = []string{db}
	}
}

func NewMysqlIntrospector(ctx context.Context, dsn string, opts ...MysqlOption) (*MysqlIntrospector, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to mysql: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	mi := &MysqlIntrospector{
		db:      db,
		schemas: []string{"*"},
	}

	for _, opt := range opts {
		opt(mi)
	}

	return mi, nil
}

func (mi *MysqlIntrospector) Introspect(ctx context.Context, dsn string) (*schema.Schema, error) {
	return mi.introspect(ctx)
}

func (mi *MysqlIntrospector) introspect(ctx context.Context) (*schema.Schema, error) {
	tables, err := mi.extractTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("extract tables: %w", err)
	}

	return &schema.Schema{
		Name:   mi.schemas[0],
		Tables: tables,
	}, nil
}

func (mi *MysqlIntrospector) extractTables(ctx context.Context) ([]*schema.Table, error) {
	query := `
		SELECT TABLE_NAME
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME
	`

	rows, err := mi.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query tables: %w", err)
	}
	defer rows.Close()

	var tables []*schema.Table
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("scan table row: %w", err)
		}

		columns, err := mi.extractColumns(ctx, tableName)
		if err != nil {
			return nil, fmt.Errorf("extract columns for %s: %w", tableName, err)
		}

		fks, err := mi.extractForeignKeys(ctx, tableName)
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

		constraints, uniqueCols, err := mi.extractConstraints(ctx, tableName)
		if err != nil {
			return nil, fmt.Errorf("extract constraints for %s: %w", tableName, err)
		}

		for i := range columns {
			if uniqueCols[columns[i].Name] {
				columns[i].Unique = true
			}
		}

		tables = append(tables, &schema.Table{
			Name:        tableName,
			Columns:     columns,
			ForeignKeys: fks,
			Constraints: constraints,
		})
	}

	return tables, rows.Err()
}

func (mi *MysqlIntrospector) extractColumns(ctx context.Context, tableName string) ([]*schema.Column, error) {
	query := `
		SELECT
			COLUMN_NAME,
			COLUMN_TYPE,
			IS_NULLABLE,
			COLUMN_DEFAULT,
			CHARACTER_MAXIMUM_LENGTH,
			NUMERIC_PRECISION,
			NUMERIC_SCALE,
			EXTRA
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	rows, err := mi.db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("query columns: %w", err)
	}
	defer rows.Close()

	var columns []*schema.Column
	for rows.Next() {
		var colName, colType, isNullable string
		var colDefault *string
		var charMaxLen, numPrec, numScale *int
		var extra string

		if err := rows.Scan(
			&colName, &colType, &isNullable, &colDefault,
			&charMaxLen, &numPrec, &numScale, &extra,
		); err != nil {
			return nil, fmt.Errorf("scan column row: %w", err)
		}

		mappedType := mapMysqlType(colType, extra)

		col := &schema.Column{
			Name:     colName,
			Type:     mappedType,
			RawType:  colType,
			Nullable: isNullable == "YES",
			Default:  colDefault,
		}

		if charMaxLen != nil {
			col.MaxLength = *charMaxLen
		}
		if numPrec != nil {
			col.NumericPrec = *numPrec
		}
		if numScale != nil {
			col.NumericScale = *numScale
		}

		if mappedType == schema.TypeEnum {
			values := extractEnumValuesFromType(colType)
			col.EnumValues = values
		}

		columns = append(columns, col)
	}

	return columns, rows.Err()
}

func (mi *MysqlIntrospector) extractForeignKeys(ctx context.Context, tableName string) ([]*schema.ForeignKey, error) {
	query := `
		SELECT
			kcu.COLUMN_NAME,
			kcu.REFERENCED_TABLE_NAME,
			kcu.REFERENCED_COLUMN_NAME,
			rc.UPDATE_RULE,
			rc.DELETE_RULE,
			tc.CONSTRAINT_NAME
		FROM information_schema.KEY_COLUMN_USAGE kcu
		JOIN information_schema.TABLE_CONSTRAINTS tc
			ON kcu.CONSTRAINT_NAME = tc.CONSTRAINT_NAME
			AND kcu.TABLE_SCHEMA = tc.TABLE_SCHEMA
		LEFT JOIN information_schema.REFERENTIAL_CONSTRAINTS rc
			ON kcu.CONSTRAINT_NAME = rc.CONSTRAINT_NAME
			AND kcu.TABLE_SCHEMA = rc.CONSTRAINT_SCHEMA
		WHERE tc.CONSTRAINT_TYPE = 'FOREIGN KEY'
			AND kcu.TABLE_SCHEMA = DATABASE()
			AND kcu.TABLE_NAME = ?
		ORDER BY tc.CONSTRAINT_NAME, kcu.ORDINAL_POSITION
	`

	rows, err := mi.db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("query foreign keys: %w", err)
	}
	defer rows.Close()

	var fks []*schema.ForeignKey
	for rows.Next() {
		fk := &schema.ForeignKey{}
		if err := rows.Scan(
			&fk.ColumnName, &fk.RefTable, &fk.RefColumn,
			&fk.OnUpdate, &fk.OnDelete, &fk.ConstraintName,
		); err != nil {
			return nil, fmt.Errorf("scan foreign key row: %w", err)
		}
		fks = append(fks, fk)
	}

	return fks, rows.Err()
}

func (mi *MysqlIntrospector) extractConstraints(ctx context.Context, tableName string) ([]*schema.Constraint, map[string]bool, error) {
	uniqueCols := make(map[string]bool)
	var constraints []*schema.Constraint

	query := `
		SELECT
			tc.CONSTRAINT_TYPE,
			tc.CONSTRAINT_NAME,
			COALESCE(kcu.COLUMN_NAME, '') AS column_name
		FROM information_schema.TABLE_CONSTRAINTS tc
		LEFT JOIN information_schema.KEY_COLUMN_USAGE kcu
			ON tc.CONSTRAINT_NAME = kcu.CONSTRAINT_NAME
			AND tc.TABLE_SCHEMA = kcu.TABLE_SCHEMA
		WHERE tc.TABLE_SCHEMA = DATABASE()
			AND tc.TABLE_NAME = ?
			AND tc.CONSTRAINT_TYPE IN ('UNIQUE', 'PRIMARY KEY')
		ORDER BY tc.CONSTRAINT_NAME, kcu.ORDINAL_POSITION
	`

	rows, err := mi.db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, nil, fmt.Errorf("query constraints: %w", err)
	}
	defer rows.Close()

	currentConstraint := ""
	var currentCols []string
	var currentType schema.ConstraintType
	var currentName string

	flushConstraint := func() {
		if currentConstraint != "" {
			c := &schema.Constraint{
				Type:    currentType,
				Name:    currentName,
				Columns: currentCols,
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
		var constraintType, constraintName, columnName string
		if err := rows.Scan(&constraintType, &constraintName, &columnName); err != nil {
			return nil, nil, fmt.Errorf("scan constraint row: %w", err)
		}

		if constraintName != currentConstraint {
			flushConstraint()
			currentConstraint = constraintName
			currentCols = nil
			currentName = constraintName
			switch constraintType {
			case "UNIQUE":
				currentType = schema.ConstraintUnique
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

func (mi *MysqlIntrospector) Close() {
	if mi.db != nil {
		mi.db.Close()
	}
}

func mapMysqlType(columnType, extra string) schema.ColumnType {
	colType := strings.ToLower(columnType)

	if strings.Contains(colType, "enum") {
		return schema.TypeEnum
	}

	if strings.Contains(extra, "auto_increment") {
		if strings.HasPrefix(colType, "bigint") {
			return schema.TypeBigSerial
		}
		return schema.TypeSerial
	}

	if strings.HasPrefix(colType, "tinyint") {
		if strings.Contains(colType, "unsigned") {
			return schema.TypeSmallInt
		}
		return schema.TypeBoolean
	}

	if strings.HasPrefix(colType, "smallint") {
		return schema.TypeSmallInt
	}

	if strings.HasPrefix(colType, "mediumint") {
		return schema.TypeInteger
	}

	if strings.HasPrefix(colType, "int") {
		if strings.Contains(colType, "unsigned") {
			return schema.TypeBigInt
		}
		return schema.TypeInteger
	}

	if strings.HasPrefix(colType, "bigint") {
		return schema.TypeBigInt
	}

	if strings.HasPrefix(colType, "float") {
		return schema.TypeFloat
	}

	if strings.HasPrefix(colType, "double") {
		return schema.TypeDouble
	}

	if strings.HasPrefix(colType, "decimal") || strings.HasPrefix(colType, "numeric") {
		return schema.TypeNumeric
	}

	if strings.HasPrefix(colType, "varchar") {
		return schema.TypeVarchar
	}

	if strings.HasPrefix(colType, "char") && !strings.HasPrefix(colType, "varchar") {
		return schema.TypeChar
	}

	if strings.HasPrefix(colType, "text") || colType == "tinytext" || colType == "mediumtext" || colType == "longtext" {
		return schema.TypeText
	}

	if strings.HasPrefix(colType, "blob") || strings.HasPrefix(colType, "binary") || strings.HasPrefix(colType, "varbinary") {
		return schema.TypeBytea
	}

	if strings.HasPrefix(colType, "date") && !strings.HasPrefix(colType, "datetime") {
		return schema.TypeDate
	}

	if strings.HasPrefix(colType, "datetime") {
		return schema.TypeTimestamp
	}

	if strings.HasPrefix(colType, "timestamp") {
		return schema.TypeTimestamptz
	}

	if strings.HasPrefix(colType, "time") {
		return schema.TypeTime
	}

	if colType == "json" || colType == "jsonb" {
		return schema.TypeJSON
	}

	return schema.TypeUnknown
}

func extractEnumValuesFromType(colType string) []string {
	if !strings.HasPrefix(colType, "enum") {
		return nil
	}

	start := strings.Index(colType, "(")
	end := strings.LastIndex(colType, ")")
	if start < 0 || end < 0 || end <= start {
		return nil
	}

	inner := colType[start+1 : end]
	var values []string
	for _, v := range strings.Split(inner, ",") {
		v = strings.TrimSpace(v)
		v = strings.Trim(v, "'")
		if v != "" {
			values = append(values, v)
		}
	}
	return values
}
