package generator

import (
	"context"
	"testing"
)

func TestFloatRangeGenerator(t *testing.T) {
	g := &FloatRangeGenerator{Min: 1.5, Max: 9.5}
	ctx := context.Background()
	row := testRow{}

	for i := 0; i < 100; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		f := v.(float64)
		if f < 1.5 || f > 9.5 {
			t.Errorf("float %f out of range [1.5, 9.5]", f)
		}
	}
}

func TestConstantGenerator(t *testing.T) {
	g := &ConstantGenerator{Value: "hello"}
	ctx := context.Background()
	row := testRow{}

	for i := 0; i < 10; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		if v != "hello" {
			t.Errorf("expected 'hello', got %v", v)
		}
	}
}

func TestConstantGeneratorInt(t *testing.T) {
	g := &ConstantGenerator{Value: 42}
	ctx := context.Background()
	row := testRow{}

	v, _ := g.Generate(ctx, row)
	if v != 42 {
		t.Errorf("expected 42, got %v", v)
	}
}

func TestWeightedChoiceGenerator(t *testing.T) {
	g := &WeightedChoiceGenerator{
		Weights: []WeightedChoice{
			{Label: "common", Weight: 90},
			{Label: "rare", Weight: 10},
		},
	}
	ctx := context.Background()
	row := testRow{}

	commonCount := 0
	for i := 0; i < 1000; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		if v.(string) == "common" {
			commonCount++
		}
	}
	if commonCount < 800 {
		t.Errorf("expected ~900 common picks, got %d", commonCount)
	}
}

func TestWeightedChoiceSingleValue(t *testing.T) {
	g := &WeightedChoiceGenerator{
		Weights: []WeightedChoice{
			{Label: "only", Weight: 1},
		},
	}
	ctx := context.Background()
	row := testRow{}

	v, _ := g.Generate(ctx, row)
	if v != "only" {
		t.Errorf("expected 'only', got %v", v)
	}
}
