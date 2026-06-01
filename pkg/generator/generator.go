package generator

import (
	"context"
	"sync"
)

type RowContext interface {
	Column(name string) (any, bool)
	RowIndex() int64
	TableName() string
}

type Generator interface {
	Generate(ctx context.Context, row RowContext) (any, error)
}

type GeneratorFunc func(ctx context.Context, row RowContext) (any, error)

func (f GeneratorFunc) Generate(ctx context.Context, row RowContext) (any, error) {
	return f(ctx, row)
}

type Config struct {
	TableName string                 `yaml:"table_name" json:"table_name"`
	Columns   map[string]ColumnConfig `yaml:"columns" json:"columns"`
	Count     int                    `yaml:"count,omitempty" json:"count,omitempty"`
}

type ColumnConfig struct {
	Generator string         `yaml:"generator" json:"generator"`
	Params    map[string]any `yaml:"params,omitempty" json:"params,omitempty"`
	Disabled  bool           `yaml:"disabled,omitempty" json:"disabled,omitempty"`
}

type Registry struct {
	mu         sync.RWMutex
	generators map[string]Generator
}

func NewRegistry() *Registry {
	return &Registry{
		generators: make(map[string]Generator),
	}
}

func (r *Registry) Register(name string, g Generator) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.generators[name] = g
}

func (r *Registry) Get(name string) (Generator, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	g, ok := r.generators[name]
	return g, ok
}

func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.generators))
	for n := range r.generators {
		names = append(names, n)
	}
	return names
}
