package introspect

import (
	"testing"
)

func TestMapMysqlType(t *testing.T) {
	tests := []struct {
		colType string
		extra   string
		want    string
	}{
		{"tinyint(1)", "", "boolean"},
		{"tinyint(3) unsigned", "", "smallint"},
		{"smallint(5)", "", "smallint"},
		{"mediumint(8)", "", "integer"},
		{"int(11)", "", "integer"},
		{"int(11) unsigned", "", "bigint"},
		{"bigint(20)", "", "bigint"},
		{"float", "", "float"},
		{"double", "", "double"},
		{"decimal(10,2)", "", "numeric"},
		{"varchar(255)", "", "varchar"},
		{"char(10)", "", "char"},
		{"text", "", "text"},
		{"longtext", "", "text"},
		{"blob", "", "bytea"},
		{"date", "", "date"},
		{"datetime", "", "timestamp"},
		{"timestamp", "", "timestamptz"},
		{"time", "", "time"},
		{"json", "", "json"},
		{"int(11)", "auto_increment", "serial"},
		{"enum('a','b','c')", "", "enum"},
	}

	for _, tt := range tests {
		t.Run(tt.colType, func(t *testing.T) {
			got := mapMysqlType(tt.colType, tt.extra)
			if string(got) != tt.want {
				t.Errorf("mapMysqlType(%q) = %q, want %q", tt.colType, got, tt.want)
			}
		})
	}
}

func TestExtractEnumValues(t *testing.T) {
	colType := "enum('ACTIVE','INACTIVE','PENDING')"
	values := extractEnumValuesFromType(colType)

	if len(values) != 3 {
		t.Fatalf("expected 3 values, got %d", len(values))
	}
	if values[0] != "ACTIVE" {
		t.Errorf("expected ACTIVE, got %s", values[0])
	}
	if values[1] != "INACTIVE" {
		t.Errorf("expected INACTIVE, got %s", values[1])
	}
	if values[2] != "PENDING" {
		t.Errorf("expected PENDING, got %s", values[2])
	}
}

func TestExtractEnumValuesEmpty(t *testing.T) {
	values := extractEnumValuesFromType("varchar(255)")
	if len(values) != 0 {
		t.Errorf("expected 0 values, got %d", len(values))
	}
}
