package generator

import (
	"context"
	"strings"
	"testing"
)

func TestUsernameGenerator(t *testing.T) {
	g := &UsernameGenerator{}
	ctx := context.Background()
	row := testRow{}

	for i := 0; i < 50; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		username := v.(string)
		if len(username) < 5 {
			t.Errorf("username too short: %q", username)
		}
		if strings.Contains(username, " ") {
			t.Errorf("username contains space: %q", username)
		}
	}
}

func TestULIDGenerator(t *testing.T) {
	g := &ULIDGenerator{}
	ctx := context.Background()
	row := testRow{}

	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		ulid := v.(string)
		if len(ulid) != 26 {
			t.Errorf("expected 26-char ULID, got %d: %q", len(ulid), ulid)
		}
		if seen[ulid] {
			t.Errorf("duplicate ULID: %s", ulid)
		}
		seen[ulid] = true
	}
}

func TestStringGenerator(t *testing.T) {
	g := &StringGenerator{Length: 16}
	ctx := context.Background()
	row := testRow{}

	for i := 0; i < 50; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		s := v.(string)
		if len(s) != 16 {
			t.Errorf("expected 16-char string, got %d: %q", len(s), s)
		}
	}
}

func TestStringGeneratorDefaultLength(t *testing.T) {
	g := &StringGenerator{}
	ctx := context.Background()
	row := testRow{}

	v, err := g.Generate(ctx, row)
	if err != nil {
		t.Fatal(err)
	}
	s := v.(string)
	if len(s) != 10 {
		t.Errorf("expected default 10-char string, got %d: %q", len(s), s)
	}
}

func TestEnumPickerAllValuesSeen(t *testing.T) {
	values := []string{"A", "B", "C"}
	g := &EnumPicker{Values: values}
	ctx := context.Background()
	row := testRow{}

	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		v, _ := g.Generate(ctx, row)
		seen[v.(string)] = true
	}
	for _, v := range values {
		if !seen[v] {
			t.Errorf("enum value %q never picked", v)
		}
	}
}
