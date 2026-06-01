package introspect

import (
	"context"
	"strings"

	"github.com/tomiwa-a/seedling/pkg/schema"
)

func mapColumnType(udtName, dataType string) schema.ColumnType {
	if dataType == "USER-DEFINED" {
		return schema.TypeEnum
	}

	switch strings.ToLower(udtName) {
	case "int2", "smallint":
		return schema.TypeSmallInt
	case "int4", "integer":
		return schema.TypeInteger
	case "int8", "bigint":
		return schema.TypeBigInt
	case "serial", "serial4":
		return schema.TypeSerial
	case "bigserial", "serial8":
		return schema.TypeBigSerial
	case "bool", "boolean":
		return schema.TypeBoolean
	case "varchar", "character varying":
		return schema.TypeVarchar
	case "char", "character", "bpchar":
		return schema.TypeChar
	case "text":
		return schema.TypeText
	case "numeric", "decimal":
		return schema.TypeNumeric
	case "float4", "real":
		return schema.TypeReal
	case "float8", "double precision":
		return schema.TypeDouble
	case "timestamp", "timestamp without time zone":
		return schema.TypeTimestamp
	case "timestamptz", "timestamp with time zone":
		return schema.TypeTimestamptz
	case "date":
		return schema.TypeDate
	case "time", "time without time zone", "timetz", "time with time zone":
		return schema.TypeTime
	case "interval":
		return schema.TypeInterval
	case "uuid":
		return schema.TypeUUID
	case "json":
		return schema.TypeJSON
	case "jsonb":
		return schema.TypeJSONB
	case "bytea":
		return schema.TypeBytea
	case "inet":
		return schema.TypeInet
	case "macaddr", "macaddr8":
		return schema.TypeMACAddr
	case "money":
		return schema.TypeMoney
	default:
		return schema.TypeUnknown
	}
}

func isSerialType(udtName string) bool {
	switch strings.ToLower(udtName) {
	case "serial", "serial4", "bigserial", "serial8":
		return true
	}
	return false
}

func (pi *PostgresIntrospector) extractEnumValues(ctx context.Context, schemaName, typeName string) ([]string, error) {
	rows, err := pi.pool.Query(ctx, `
		SELECT e.enumlabel
		FROM pg_enum e
		JOIN pg_type t ON e.enumtypid = t.oid
		JOIN pg_namespace n ON t.typnamespace = n.oid
		WHERE n.nspname = $1 AND t.typname = $2
		ORDER BY e.enumsortorder
	`, schemaName, typeName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var values []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	return values, rows.Err()
}
