package schema_test

import (
	"testing"

	"github.com/tomiwa-a/seedling/pkg/schema"
)

func TestTableString(t *testing.T) {
	tbl := schema.Table{
		SchemaName: "public",
		Name:       "users",
	}
	want := "public.users"
	if got := tbl.String(); got != want {
		t.Errorf("Table.String() = %q, want %q", got, want)
	}
}

func TestTableStringNoSchema(t *testing.T) {
	tbl := schema.Table{
		Name: "users",
	}
	want := ".users"
	if got := tbl.String(); got != want {
		t.Errorf("Table.String() = %q, want %q", got, want)
	}
}

func TestTableColumnNames(t *testing.T) {
	tbl := schema.Table{
		Name: "users",
		Columns: []schema.Column{
			{Name: "id", Type: schema.TypeSerial},
			{Name: "email", Type: schema.TypeVarchar},
			{Name: "name", Type: schema.TypeVarchar},
		},
	}
	want := []string{"id", "email", "name"}
	got := tbl.ColumnNames()
	if len(got) != len(want) {
		t.Fatalf("ColumnNames() length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ColumnNames()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestTableColumnNamesEmpty(t *testing.T) {
	tbl := schema.Table{Name: "empty"}
	got := tbl.ColumnNames()
	if len(got) != 0 {
		t.Errorf("ColumnNames() = %v, want empty slice", got)
	}
}

func TestFindColumnFound(t *testing.T) {
	tbl := schema.Table{
		Name: "users",
		Columns: []schema.Column{
			{Name: "id", Type: schema.TypeSerial},
			{Name: "email", Type: schema.TypeVarchar},
		},
	}
	col := tbl.FindColumn("email")
	if col == nil {
		t.Fatal("FindColumn() returned nil, expected non-nil")
	}
	if col.Name != "email" {
		t.Errorf("FindColumn().Name = %q, want %q", col.Name, "email")
	}
	if col.Type != schema.TypeVarchar {
		t.Errorf("FindColumn().Type = %q, want %q", col.Type, schema.TypeVarchar)
	}
}

func TestFindColumnNotFound(t *testing.T) {
	tbl := schema.Table{
		Name: "users",
		Columns: []schema.Column{
			{Name: "id", Type: schema.TypeSerial},
		},
	}
	col := tbl.FindColumn("nonexistent")
	if col != nil {
		t.Errorf("FindColumn() = %v, want nil", col)
	}
}

func TestFindColumnEmpty(t *testing.T) {
	tbl := schema.Table{Name: "empty"}
	col := tbl.FindColumn("anything")
	if col != nil {
		t.Errorf("FindColumn() = %v, want nil", col)
	}
}

func TestColumnTypesAreStrings(t *testing.T) {
	tests := []struct {
		ct   schema.ColumnType
		want string
	}{
		{schema.TypeSerial, "serial"},
		{schema.TypeBigSerial, "bigserial"},
		{schema.TypeInteger, "integer"},
		{schema.TypeVarchar, "varchar"},
		{schema.TypeTimestamptz, "timestamptz"},
		{schema.TypeUUID, "uuid"},
		{schema.TypeJSONB, "jsonb"},
		{schema.TypeEnum, "enum"},
		{schema.TypeUnknown, "unknown"},
	}
	for _, tt := range tests {
		if string(tt.ct) != tt.want {
			t.Errorf("ColumnType(%q) = %s, want %s", tt.want, string(tt.ct), tt.want)
		}
	}
}

func TestGeneratorHintsAreStrings(t *testing.T) {
	tests := []struct {
		gh   schema.GeneratorHint
		want string
	}{
		{schema.HintAuto, "auto"},
		{schema.HintEmail, "email"},
		{schema.HintCity, "city"},
		{schema.HintUUID, "uuid"},
		{schema.HintNow, "now"},
	}
	for _, tt := range tests {
		if string(tt.gh) != tt.want {
			t.Errorf("GeneratorHint(%q) = %s, want %s", tt.want, string(tt.gh), tt.want)
		}
	}
}

func TestSchemaStruct(t *testing.T) {
	s := schema.Schema{
		Name: "public",
		Tables: []schema.Table{
			{
				Name: "users",
				Columns: []schema.Column{
					{Name: "id", Type: schema.TypeSerial, Nullable: false},
				},
			},
		},
	}
	if len(s.Tables) != 1 {
		t.Errorf("Schema.Tables length = %d, want 1", len(s.Tables))
	}
	if s.Tables[0].Name != "users" {
		t.Errorf("Schema.Table[0].Name = %q, want %q", s.Tables[0].Name, "users")
	}
}

func TestColumnFKRef(t *testing.T) {
	col := schema.Column{
		Name: "user_id",
		Type: schema.TypeInteger,
		FKRef: &schema.FKRef{
			Table:  "users",
			Column: "id",
		},
	}
	if col.FKRef == nil {
		t.Fatal("FKRef is nil")
	}
	if col.FKRef.Table != "users" {
		t.Errorf("FKRef.Table = %q, want %q", col.FKRef.Table, "users")
	}
	if col.FKRef.Column != "id" {
		t.Errorf("FKRef.Column = %q, want %q", col.FKRef.Column, "id")
	}
}

func TestConstraintTypeValues(t *testing.T) {
	tests := []struct {
		ct   schema.ConstraintType
		want string
	}{
		{schema.ConstraintUnique, "UNIQUE"},
		{schema.ConstraintCheck, "CHECK"},
		{schema.ConstraintNotNull, "NOT NULL"},
		{schema.ConstraintPrimaryKey, "PRIMARY KEY"},
	}
	for _, tt := range tests {
		if string(tt.ct) != tt.want {
			t.Errorf("ConstraintType = %s, want %s", string(tt.ct), tt.want)
		}
	}
}

func TestForeignKeyStruct(t *testing.T) {
	fk := schema.ForeignKey{
		ColumnName:     "user_id",
		RefTable:       "users",
		RefColumn:      "id",
		OnDelete:       "CASCADE",
		OnUpdate:       "NO ACTION",
		ConstraintName: "fk_orders_user",
	}
	if fk.ColumnName != "user_id" {
		t.Errorf("ForeignKey.ColumnName = %q", fk.ColumnName)
	}
	if fk.ConstraintName != "fk_orders_user" {
		t.Errorf("ForeignKey.ConstraintName = %q", fk.ConstraintName)
	}
}
