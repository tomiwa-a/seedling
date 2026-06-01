package generator

import (
	"context"
	"fmt"
)

type TableConfig struct {
	TableName string
	Columns   map[string]ColumnOverride
}

type ColumnOverride struct {
	Generator Generator
	Disabled  bool
}

type TableBuilder struct {
	config *TableConfig
}

func Table(name string) *TableBuilder {
	return &TableBuilder{
		config: &TableConfig{
			TableName: name,
			Columns:   make(map[string]ColumnOverride),
		},
	}
}

func (b *TableBuilder) Column(name string, gen Generator) *TableBuilder {
	b.config.Columns[name] = ColumnOverride{Generator: gen}
	return b
}

func (b *TableBuilder) ColumnDisabled(name string) *TableBuilder {
	b.config.Columns[name] = ColumnOverride{Disabled: true}
	return b
}

func (b *TableBuilder) Build() *TableConfig {
	return b.config
}

type FuncGenerator struct {
	fn func(ctx context.Context, row RowContext) (any, error)
}

func Func(fn func(ctx context.Context, row RowContext) (any, error)) *FuncGenerator {
	return &FuncGenerator{fn: fn}
}

func (g *FuncGenerator) Generate(ctx context.Context, row RowContext) (any, error) {
	return g.fn(ctx, row)
}

type Context struct {
	Row RowContext
}

func (c *Context) Column(name string) (any, bool) {
	return c.Row.Column(name)
}

func (c *Context) RowIndex() int64 {
	return c.Row.RowIndex()
}

func (c *Context) TableName() string {
	return c.Row.TableName()
}

type ValueGenerator struct {
	Value any
}

func Value(v any) *ValueGenerator {
	return &ValueGenerator{Value: v}
}

func (g *ValueGenerator) Generate(ctx context.Context, row RowContext) (any, error) {
	return g.Value, nil
}

func TableConfigsFromConfigs(configs []Config) map[string]*TableConfig {
	result := make(map[string]*TableConfig)
	for _, cfg := range configs {
		tc := &TableConfig{
			TableName: cfg.TableName,
			Columns:   make(map[string]ColumnOverride),
		}
		for colName, colCfg := range cfg.Columns {
			if colCfg.Disabled {
				tc.Columns[colName] = ColumnOverride{Disabled: true}
			} else {
				g, err := resolveBuiltinGenerator(colCfg.Generator, colCfg.Params)
				if err != nil {
					continue
				}
				tc.Columns[colName] = ColumnOverride{Generator: g}
			}
		}
		result[cfg.TableName] = tc
	}
	return result
}

func resolveBuiltinGenerator(name string, params map[string]any) (Generator, error) {
	switch name {
	case "email":
		return &emailGen{}, nil
	case "name":
		return &nameGen{}, nil
	case "phone":
		return &phoneGen{}, nil
	case "uuid":
		return &uuidGen{}, nil
	case "bool":
		return &boolGen{}, nil
	case "lorem":
		minWords := 2
		maxWords := 8
		if v, ok := params["min_words"].(int); ok {
			minWords = v
		}
		if v, ok := params["max_words"].(int); ok {
			maxWords = v
		}
		return &loremGen{minWords: minWords, maxWords: maxWords}, nil
	case "constant":
		val := params["value"]
		return Value(val), nil
	default:
		return nil, fmt.Errorf("unknown builtin generator: %s", name)
	}
}

type emailGen struct{}

func (g *emailGen) Generate(ctx context.Context, row RowContext) (any, error) {
	first, _ := row.Column("first_name")
	last, _ := row.Column("last_name")
	if first == nil {
		first = "user"
	}
	if last == nil {
		last = "test"
	}
	return fmt.Sprintf("%v.%v_%d@example.com", first, last, row.RowIndex()), nil
}

type nameGen struct{}

func (g *nameGen) Generate(ctx context.Context, row RowContext) (any, error) {
	return "John Doe", nil
}

type phoneGen struct{}

func (g *phoneGen) Generate(ctx context.Context, row RowContext) (any, error) {
	return "+1-555-0100", nil
}

type uuidGen struct{}

func (g *uuidGen) Generate(ctx context.Context, row RowContext) (any, error) {
	return "00000000-0000-0000-0000-000000000000", nil
}

type boolGen struct{}

func (g *boolGen) Generate(ctx context.Context, row RowContext) (any, error) {
	return true, nil
}

type loremGen struct {
	minWords int
	maxWords int
}

func (g *loremGen) Generate(ctx context.Context, row RowContext) (any, error) {
	return "lorem ipsum dolor sit amet", nil
}
