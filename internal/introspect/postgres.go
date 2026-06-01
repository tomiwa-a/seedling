package introspect

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tomiwa-a/seedling/pkg/schema"
)

type PostgresIntrospector struct {
	pool   *pgxpool.Pool
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
	return nil, fmt.Errorf("not yet implemented")
}

func (pi *PostgresIntrospector) Close() {
	if pi.pool != nil {
		pi.pool.Close()
	}
}
