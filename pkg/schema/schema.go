package schema

import "fmt"

type Schema struct {
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Tables      []*Table `json:"tables" yaml:"tables"`
}

type Table struct {
	Name        string        `json:"name" yaml:"name"`
	SchemaName  string        `json:"schema_name,omitempty" yaml:"schema_name,omitempty"`
	Columns     []*Column     `json:"columns" yaml:"columns"`
	ForeignKeys []*ForeignKey `json:"foreign_keys,omitempty" yaml:"foreign_keys,omitempty"`
	Constraints []*Constraint `json:"constraints,omitempty" yaml:"constraints,omitempty"`
}

func (t *Table) String() string {
	return fmt.Sprintf("%s.%s", t.SchemaName, t.Name)
}

func (t *Table) ColumnNames() []string {
	names := make([]string, len(t.Columns))
	for i, c := range t.Columns {
		names[i] = c.Name
	}
	return names
}

func (t *Table) FindColumn(name string) *Column {
	for i := range t.Columns {
		if t.Columns[i].Name == name {
			return t.Columns[i]
		}
	}
	return nil
}

type ColumnType string

const (
	TypeSerial       ColumnType = "serial"
	TypeBigSerial    ColumnType = "bigserial"
	TypeInteger      ColumnType = "integer"
	TypeBigInt       ColumnType = "bigint"
	TypeSmallInt     ColumnType = "smallint"
	TypeBoolean      ColumnType = "boolean"
	TypeText         ColumnType = "text"
	TypeVarchar      ColumnType = "varchar"
	TypeChar         ColumnType = "char"
	TypeNumeric      ColumnType = "numeric"
	TypeFloat        ColumnType = "float"
	TypeDouble       ColumnType = "double"
	TypeReal         ColumnType = "real"
	TypeTimestamp    ColumnType = "timestamp"
	TypeTimestamptz  ColumnType = "timestamptz"
	TypeDate         ColumnType = "date"
	TypeTime         ColumnType = "time"
	TypeInterval     ColumnType = "interval"
	TypeUUID         ColumnType = "uuid"
	TypeJSON         ColumnType = "json"
	TypeJSONB        ColumnType = "jsonb"
	TypeBytea        ColumnType = "bytea"
	TypeInet         ColumnType = "inet"
	TypeMACAddr      ColumnType = "macaddr"
	TypeMoney        ColumnType = "money"
	TypeEnum         ColumnType = "enum"
	TypeUnknown      ColumnType = "unknown"
)

type Column struct {
	Name         string        `json:"name" yaml:"name"`
	Type         ColumnType    `json:"type" yaml:"type"`
	RawType      string        `json:"raw_type,omitempty" yaml:"raw_type,omitempty"`
	Nullable     bool          `json:"nullable" yaml:"nullable"`
	Unique       bool          `json:"unique,omitempty" yaml:"unique,omitempty"`
	Default      *string       `json:"default,omitempty" yaml:"default,omitempty"`
	FKRef        *FKRef        `json:"fk_ref,omitempty" yaml:"fk_ref,omitempty"`
	Comment      string        `json:"comment,omitempty" yaml:"comment,omitempty"`
	Hint         GeneratorHint `json:"hint,omitempty" yaml:"hint,omitempty"`
	EnumValues   []string      `json:"enum_values,omitempty" yaml:"enum_values,omitempty"`
	MaxLength    int           `json:"max_length,omitempty" yaml:"max_length,omitempty"`
	NumericScale int           `json:"numeric_scale,omitempty" yaml:"numeric_scale,omitempty"`
	NumericPrec  int           `json:"numeric_precision,omitempty" yaml:"numeric_precision,omitempty"`
}

type FKRef struct {
	Table  string `json:"table" yaml:"table"`
	Column string `json:"column" yaml:"column"`
}

type ForeignKey struct {
	ColumnName    string `json:"column_name" yaml:"column_name"`
	RefTable      string `json:"ref_table" yaml:"ref_table"`
	RefColumn     string `json:"ref_column" yaml:"ref_column"`
	OnDelete      string `json:"on_delete,omitempty" yaml:"on_delete,omitempty"`
	OnUpdate      string `json:"on_update,omitempty" yaml:"on_update,omitempty"`
	ConstraintName string `json:"constraint_name,omitempty" yaml:"constraint_name,omitempty"`
}

type ConstraintType string

const (
	ConstraintUnique    ConstraintType = "UNIQUE"
	ConstraintCheck     ConstraintType = "CHECK"
	ConstraintNotNull   ConstraintType = "NOT NULL"
	ConstraintPrimaryKey ConstraintType = "PRIMARY KEY"
)

type Constraint struct {
	Type       ConstraintType `json:"type" yaml:"type"`
	Columns    []string       `json:"columns,omitempty" yaml:"columns,omitempty"`
	Expression string         `json:"expression,omitempty" yaml:"expression,omitempty"`
	Name       string         `json:"name,omitempty" yaml:"name,omitempty"`
}

type GeneratorHint string

const (
	HintAuto       GeneratorHint = "auto"
	HintEmail      GeneratorHint = "email"
	HintName       GeneratorHint = "full_name"
	HintCity       GeneratorHint = "city"
	HintCountry    GeneratorHint = "country"
	HintPhone      GeneratorHint = "phone"
	HintAddress    GeneratorHint = "address"
	HintCompany    GeneratorHint = "company"
	HintJobTitle   GeneratorHint = "job_title"
	HintURL        GeneratorHint = "url"
	HintIP         GeneratorHint = "ip"
	HintUUID       GeneratorHint = "uuid"
	HintCurrency   GeneratorHint = "currency"
	HintNow        GeneratorHint = "now"
	HintSequence   GeneratorHint = "sequence"
	HintCategorical GeneratorHint = "categorical"
)
