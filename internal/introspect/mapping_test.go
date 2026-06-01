package introspect

import (
	"testing"

	"github.com/tomiwa-a/seedling/pkg/schema"
)

func TestMapColumnType(t *testing.T) {
	tests := []struct {
		udtName  string
		dataType string
		want     schema.ColumnType
	}{
		{"int4", "integer", schema.TypeInteger},
		{"int8", "bigint", schema.TypeBigInt},
		{"int2", "smallint", schema.TypeSmallInt},
		{"serial", "serial", schema.TypeSerial},
		{"serial4", "serial", schema.TypeSerial},
		{"bigserial", "bigserial", schema.TypeBigSerial},
		{"serial8", "bigserial", schema.TypeBigSerial},
		{"bool", "boolean", schema.TypeBoolean},
		{"varchar", "character varying", schema.TypeVarchar},
		{"text", "text", schema.TypeText},
		{"numeric", "numeric", schema.TypeNumeric},
		{"float4", "real", schema.TypeReal},
		{"float8", "double precision", schema.TypeDouble},
		{"timestamptz", "timestamp with time zone", schema.TypeTimestamptz},
		{"timestamp", "timestamp without time zone", schema.TypeTimestamp},
		{"date", "date", schema.TypeDate},
		{"uuid", "uuid", schema.TypeUUID},
		{"json", "json", schema.TypeJSON},
		{"jsonb", "jsonb", schema.TypeJSONB},
		{"bytea", "bytea", schema.TypeBytea},
		{"inet", "inet", schema.TypeInet},
		{"money", "money", schema.TypeMoney},
		{"some_enum", "USER-DEFINED", schema.TypeEnum},
		{"unknown_thing", "some_type", schema.TypeUnknown},
	}

	for _, tt := range tests {
		got := mapColumnType(tt.udtName, tt.dataType)
		if got != tt.want {
			t.Errorf("mapColumnType(%q, %q) = %q, want %q", tt.udtName, tt.dataType, got, tt.want)
		}
	}
}

func TestMapColumnTypeCaseInsensitive(t *testing.T) {
	got := mapColumnType("INT4", "INTEGER")
	if got != schema.TypeInteger {
		t.Errorf("mapColumnType(INT4) = %q, want integer", got)
	}
}

func TestIsSerialType(t *testing.T) {
	tests := []struct {
		udtName string
		want    bool
	}{
		{"serial", true},
		{"serial4", true},
		{"bigserial", true},
		{"serial8", true},
		{"int4", false},
		{"int8", false},
		{"varchar", false},
	}
	for _, tt := range tests {
		got := isSerialType(tt.udtName)
		if got != tt.want {
			t.Errorf("isSerialType(%q) = %v, want %v", tt.udtName, got, tt.want)
		}
	}
}
