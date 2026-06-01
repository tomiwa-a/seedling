package generator

import (
	"context"
	"math/rand"
	"strings"
	"testing"

	"github.com/tomiwa-a/seedling/pkg/schema"
)

type testRow struct {
	index int64
	table string
	vals  map[string]any
}

func (r testRow) Column(name string) (any, bool) {
	v, ok := r.vals[name]
	return v, ok
}
func (r testRow) RowIndex() int64   { return r.index }
func (r testRow) TableName() string { return r.table }

func TestSerialGenerator(t *testing.T) {
	g := NewSerialGenerator(0)
	ctx := context.Background()
	row := testRow{index: 0, table: "test"}

	v1, _ := g.Generate(ctx, row)
	v2, _ := g.Generate(ctx, row)
	v3, _ := g.Generate(ctx, row)

	if v1.(int64) != 1 {
		t.Errorf("expected 1, got %d", v1)
	}
	if v2.(int64) != 2 {
		t.Errorf("expected 2, got %d", v2)
	}
	if v3.(int64) != 3 {
		t.Errorf("expected 3, got %d", v3)
	}
}

func TestRandomIntGenerator(t *testing.T) {
	g := &RandomIntGenerator{Min: 10, Max: 20}
	ctx := context.Background()
	row := testRow{}

	for i := 0; i < 100; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		n := v.(int64)
		if n < 10 || n > 20 {
			t.Errorf("value %d out of range [10,20]", n)
		}
	}
}

func TestLoremGenerator(t *testing.T) {
	g := &LoremGenerator{MinWords: 3, MaxWords: 5}
	ctx := context.Background()
	row := testRow{}

	for i := 0; i < 50; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		s := v.(string)
		words := strings.Fields(s)
		if len(words) < 3 || len(words) > 5 {
			t.Errorf("expected 3-5 words, got %d: %q", len(words), s)
		}
	}
}

func TestEmailGenerator(t *testing.T) {
	g := &EmailGenerator{}
	ctx := context.Background()

	for i := int64(0); i < 50; i++ {
		row := testRow{index: i}
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		email := v.(string)
		if !strings.Contains(email, "@") {
			t.Errorf("invalid email: %q", email)
		}
		if !strings.Contains(email, ".com") && !strings.Contains(email, ".org") && !strings.Contains(email, ".net") && !strings.Contains(email, ".co") {
			t.Errorf("email missing valid domain: %q", email)
		}
	}
}

func TestNameGenerator(t *testing.T) {
	g := &NameGenerator{}
	ctx := context.Background()
	row := testRow{}

	for i := 0; i < 50; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		name := v.(string)
		parts := strings.Fields(name)
		if len(parts) < 2 {
			t.Errorf("expected first + last name, got: %q", name)
		}
	}
}

func TestPhoneGenerator(t *testing.T) {
	g := &PhoneGenerator{}
	ctx := context.Background()
	row := testRow{}

	for i := 0; i < 50; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		phone := v.(string)
		if !strings.HasPrefix(phone, "+234-") {
			t.Errorf("invalid phone prefix: %q", phone)
		}
	}
}

func TestBoolGenerator(t *testing.T) {
	g := &BoolGenerator{rnd: rand.New(rand.NewSource(42))}
	ctx := context.Background()
	row := testRow{}

	trueCount := 0
	for i := 0; i < 100; i++ {
		v, _ := g.Generate(ctx, row)
		if v.(bool) {
			trueCount++
		}
	}
	if trueCount == 0 || trueCount == 100 {
		t.Errorf("BoolGenerator seems deterministic: trues=%d/100", trueCount)
	}
}

func TestUUIDGenerator(t *testing.T) {
	g := &UUIDGenerator{}
	ctx := context.Background()
	row := testRow{}

	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		uuid := v.(string)
		if len(uuid) != 36 {
			t.Errorf("expected 36-char UUID, got %d: %q", len(uuid), uuid)
		}
		if seen[uuid] {
			t.Errorf("duplicate UUID: %s", uuid)
		}
		seen[uuid] = true
	}
}

func TestEnumPicker(t *testing.T) {
	values := []string{"ACTIVE", "INACTIVE", "PENDING"}
	g := &EnumPicker{Values: values, rnd: rand.New(rand.NewSource(42))}
	ctx := context.Background()
	row := testRow{}

	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		val := v.(string)
		if !seen[val] {
			seen[val] = true
		}
	}
	for _, v := range values {
		if !seen[v] {
			t.Errorf("enum value %q never picked", v)
		}
	}
}

func TestDateGenerator(t *testing.T) {
	g := &DateGenerator{MinYear: 2024, MaxYear: 2026}
	ctx := context.Background()
	row := testRow{}

	for i := 0; i < 50; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		date := v.(string)
		if !strings.Contains(date, "2024") && !strings.Contains(date, "2025") && !strings.Contains(date, "2026") {
			t.Errorf("date %q outside expected year range", date)
		}
	}
}

func TestTimestampGenerator(t *testing.T) {
	g := &TimestampGenerator{}
	ctx := context.Background()
	row := testRow{}

	for i := 0; i < 50; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		ts := v.(string)
		if !strings.HasSuffix(ts, "Z") {
			t.Errorf("timestamp missing Z suffix: %q", ts)
		}
	}
}

func TestNumericGenerator(t *testing.T) {
	g := &NumericGenerator{Min: 100, Max: 999999}
	ctx := context.Background()
	row := testRow{}

	for i := 0; i < 50; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		n := v.(float64)
		if n < 1.00 || n > 9999.99 {
			t.Errorf("numeric %f out of expected range", n)
		}
	}
}

func TestJSONGenerator(t *testing.T) {
	g := &JSONGenerator{}
	ctx := context.Background()
	row := testRow{}

	v, err := g.Generate(ctx, row)
	if err != nil {
		t.Fatal(err)
	}
	if v.(string) != "{}" {
		t.Errorf("expected {}, got %q", v)
	}
}

func TestByteaGenerator(t *testing.T) {
	g := &ByteaGenerator{}
	ctx := context.Background()
	row := testRow{}

	v, err := g.Generate(ctx, row)
	if err != nil {
		t.Fatal(err)
	}
	b := v.([]byte)
	if len(b) != 32 {
		t.Errorf("expected 32 bytes, got %d", len(b))
	}
}

func TestResolveGeneratorByFK(t *testing.T) {
	pool := NewFKPool()
	pool.Add("users", int64(1))
	pool.Add("users", int64(2))

	col := &schema.Column{
		Name: "user_id",
		FKRef: &schema.FKRef{
			Table:  "users",
			Column: "id",
		},
	}

	g, err := ResolveGenerator(col, schema.HintAuto, pool)
	if err != nil {
		t.Fatal(err)
	}

	row := testRow{}
	v, err := g.Generate(context.Background(), row)
	if err != nil {
		t.Fatal(err)
	}
	if v.(int64) != 1 && v.(int64) != 2 {
		t.Errorf("expected FK value 1 or 2, got %d", v)
	}
}

func TestResolveGeneratorByType(t *testing.T) {
	pool := NewFKPool()
	tests := []struct {
		col  *schema.Column
		name string
	}{
		{&schema.Column{Name: "id", Type: schema.TypeSerial}, "serial"},
		{&schema.Column{Name: "count", Type: schema.TypeInteger}, "integer"},
		{&schema.Column{Name: "active", Type: schema.TypeBoolean}, "boolean"},
		{&schema.Column{Name: "body", Type: schema.TypeText}, "text"},
		{&schema.Column{Name: "uid", Type: schema.TypeUUID}, "uuid"},
		{&schema.Column{Name: "created", Type: schema.TypeTimestamptz}, "timestamptz"},
		{&schema.Column{Name: "amount", Type: schema.TypeNumeric}, "numeric"},
		{&schema.Column{Name: "data", Type: schema.TypeJSONB}, "jsonb"},
		{&schema.Column{Name: "action", Type: schema.TypeEnum, EnumValues: []string{"A", "B"}}, "enum"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := ResolveGenerator(tt.col, schema.HintAuto, pool)
			if err != nil {
				t.Fatal(err)
			}
			if g == nil {
				t.Fatal("generator is nil")
			}
			row := testRow{}
			v, err := g.Generate(context.Background(), row)
			if err != nil {
				t.Fatal(err)
			}
			if v == nil {
				t.Error("generated value is nil")
			}
		})
	}
}

func TestResolveGeneratorByHint(t *testing.T) {
	pool := NewFKPool()
	tests := []struct {
		col  *schema.Column
		hint schema.GeneratorHint
		name string
	}{
		{&schema.Column{Name: "email", Type: schema.TypeVarchar}, schema.HintEmail, "email"},
		{&schema.Column{Name: "city", Type: schema.TypeVarchar}, schema.HintCity, "city"},
		{&schema.Column{Name: "phone", Type: schema.TypeVarchar}, schema.HintPhone, "phone"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := ResolveGenerator(tt.col, tt.hint, pool)
			if err != nil {
				t.Fatal(err)
			}
			if g == nil {
				t.Fatal("generator is nil")
			}
			row := testRow{}
			v, err := g.Generate(context.Background(), row)
			if err != nil {
				t.Fatal(err)
			}
			if v == nil {
				t.Error("generated value is nil")
			}
		})
	}
}

func TestResolveGenerators(t *testing.T) {
	pool := NewFKPool()
	columns := []*schema.Column{
		{Name: "id", Type: schema.TypeSerial},
		{Name: "email", Type: schema.TypeVarchar},
		{Name: "name", Type: schema.TypeVarchar},
	}
	hints := map[string]schema.GeneratorHint{
		"email": schema.HintEmail,
		"name":  schema.HintName,
	}

	gens, err := ResolveGenerators(columns, hints, pool, 42)
	if err != nil {
		t.Fatal(err)
	}

	if len(gens) != 3 {
		t.Errorf("expected 3 generators, got %d", len(gens))
	}
	for _, name := range []string{"id", "email", "name"} {
		if _, ok := gens[name]; !ok {
			t.Errorf("missing generator for %q", name)
		}
	}
}

func TestFKPool(t *testing.T) {
	p := NewFKPool()
	p.Add("users", int64(1))
	p.Add("users", int64(2))
	p.Add("users", int64(3))

	if p.Count("users") != 3 {
		t.Errorf("Count() = %d, want 3", p.Count("users"))
	}

	seen := make(map[int64]bool)
	for i := 0; i < 30; i++ {
		v, err := p.Pick("users")
		if err != nil {
			t.Fatal(err)
		}
		seen[v.(int64)] = true
	}

	if len(seen) != 3 {
		t.Errorf("expected all 3 values to be picked at least once, got %d", len(seen))
	}
}

func TestFKPoolEmpty(t *testing.T) {
	p := NewFKPool()
	_, err := p.Pick("nonexistent")
	if err == nil {
		t.Error("expected error for empty pool")
	}
}

func TestFKPoolTables(t *testing.T) {
	p := NewFKPool()
	p.Add("a", 1)
	p.Add("b", 2)

	tables := p.Tables()
	if len(tables) != 2 {
		t.Errorf("expected 2 tables, got %d", len(tables))
	}
}

func TestAddressGenerator(t *testing.T) {
	g := &AddressGenerator{}
	ctx := context.Background()
	row := testRow{}

	for i := 0; i < 50; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		addr := v.(string)
		if !strings.Contains(addr, "Street") && !strings.Contains(addr, "Road") && !strings.Contains(addr, "Avenue") && !strings.Contains(addr, "Drive") && !strings.Contains(addr, "Lane") {
			t.Errorf("address missing street type: %q", addr)
		}
	}
}
