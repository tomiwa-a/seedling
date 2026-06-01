package generator

import (
	"testing"
)

func TestFKPoolSequential(t *testing.T) {
	p := NewFKPool()
	p.SetStrategy("users", FKStrategySequential)
	p.Add("users", int64(1))
	p.Add("users", int64(2))
	p.Add("users", int64(3))

	for i := 0; i < 6; i++ {
		v, err := p.Pick("users")
		if err != nil {
			t.Fatal(err)
		}
		expected := int64((i % 3) + 1)
		if v.(int64) != expected {
			t.Errorf("pick %d: expected %d, got %d", i, expected, v)
		}
	}
}

func TestFKPoolWeighted(t *testing.T) {
	p := NewFKPool()
	p.SetStrategy("users", FKStrategyWeighted)
	p.SetWeights("users", []float64{0.9, 0.1})
	p.Add("users", int64(1))
	p.Add("users", int64(2))

	count1 := 0
	for i := 0; i < 1000; i++ {
		v, _ := p.Pick("users")
		if v.(int64) == 1 {
			count1++
		}
	}

	if count1 < 800 {
		t.Errorf("expected ~900 picks of value 1, got %d", count1)
	}
}

func TestFKPoolUniform(t *testing.T) {
	p := NewFKPool()
	p.Add("users", int64(1))
	p.Add("users", int64(2))
	p.Add("users", int64(3))

	seen := make(map[int64]bool)
	for i := 0; i < 100; i++ {
		v, _ := p.Pick("users")
		seen[v.(int64)] = true
	}

	if len(seen) != 3 {
		t.Errorf("expected all 3 values picked, got %d", len(seen))
	}
}
