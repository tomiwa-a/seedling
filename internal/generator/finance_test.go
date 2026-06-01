package generator

import (
	"context"
	"testing"
)

func TestCurrencyGenerator(t *testing.T) {
	g := &CurrencyGenerator{}
	ctx := context.Background()
	row := testRow{}

	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		code := v.(string)
		if len(code) != 3 {
			t.Errorf("expected 3-char currency code, got %d: %q", len(code), code)
		}
		seen[code] = true
	}
	if len(seen) < 5 {
		t.Errorf("expected at least 5 different currencies, got %d", len(seen))
	}
}

func TestAmountGenerator(t *testing.T) {
	g := &AmountGenerator{MinCents: 100, MaxCents: 10000}
	ctx := context.Background()
	row := testRow{}

	for i := 0; i < 100; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		amount := v.(float64)
		if amount < 1.0 || amount > 100.0 {
			t.Errorf("amount %f out of range [1.0, 100.0]", amount)
		}
	}
}

func TestAmountGeneratorDefaults(t *testing.T) {
	g := &AmountGenerator{}
	ctx := context.Background()
	row := testRow{}

	v, err := g.Generate(ctx, row)
	if err != nil {
		t.Fatal(err)
	}
	amount := v.(float64)
	if amount < 1.0 {
		t.Errorf("amount %f too small", amount)
	}
}
