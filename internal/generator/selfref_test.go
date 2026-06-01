package generator

import (
	"context"
	"math/rand"
	"testing"
)

func TestSelfFKGeneratorBasic(t *testing.T) {
	pool := NewFKPool()
	pool.Add("employees", int64(1))
	pool.Add("employees", int64(2))
	pool.Add("employees", int64(3))

	g := NewSelfFKGenerator("employees", "manager_id", pool, 3)
	g.SetRand(seedRand(42))

	ctx := context.Background()
	row := &testRow{index: 0, table: "employees"}

	val, err := g.Generate(ctx, row)
	if err != nil {
		t.Fatal(err)
	}

	if val == nil {
		t.Error("expected non-nil value")
	}

	id := val.(int64)
	if id < 1 || id > 3 {
		t.Errorf("self FK value %d out of range [1,3]", id)
	}
}

func TestSelfFKGeneratorDepthLimit(t *testing.T) {
	pool := NewFKPool()
	pool.Add("employees", int64(1))

	g := NewSelfFKGenerator("employees", "manager_id", pool, 1)

	ctx := context.Background()
	row := &testRow{index: 0, table: "employees"}

	val, err := g.Generate(ctx, row)
	if err != nil {
		t.Fatal(err)
	}
	if val == nil {
		t.Error("expected non-nil value at depth 0")
	}

	val, err = g.Generate(ctx, row)
	if err != nil {
		t.Fatal(err)
	}
	if val != nil {
		t.Error("expected nil at depth limit")
	}
}

func TestSelfFKGeneratorEmptyPool(t *testing.T) {
	pool := NewFKPool()
	g := NewSelfFKGenerator("employees", "manager_id", pool, 3)

	ctx := context.Background()
	row := &testRow{index: 0, table: "employees"}

	val, err := g.Generate(ctx, row)
	if err != nil {
		t.Fatal(err)
	}
	if val != nil {
		t.Error("expected nil for empty pool")
	}
}

func seedRand(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}
