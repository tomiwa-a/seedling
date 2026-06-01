package generator

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"sync/atomic"

	gen "github.com/tomiwa-a/seedling/pkg/generator"
	"github.com/tomiwa-a/seedling/pkg/schema"
)

type SerialGenerator struct {
	counter atomic.Int64
}

func NewSerialGenerator(start int64) *SerialGenerator {
	s := &SerialGenerator{}
	s.counter.Store(start)
	return s
}

func (g *SerialGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return g.counter.Add(1), nil
}

type RandomIntGenerator struct {
	Min int64
	Max int64
}

func (g *RandomIntGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	delta := g.Max - g.Min + 1
	n, err := rand.Int(rand.Reader, big.NewInt(delta))
	if err != nil {
		return nil, err
	}
	return g.Min + n.Int64(), nil
}

type LoremGenerator struct {
	MinWords int
	MaxWords int
}

func (g *LoremGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	delta := g.MaxWords - g.MinWords + 1
	n, err := rand.Int(rand.Reader, big.NewInt(int64(delta)))
	if err != nil {
		return nil, err
	}
	count := g.MinWords + int(n.Int64())

	words := make([]string, count)
	for i := 0; i < count; i++ {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(loremWords))))
		words[i] = loremWords[idx.Int64()]
	}
	return strings.Join(words, " "), nil
}

type EmailGenerator struct {
	Domains []string
}

func (g *EmailGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	name := randomPick(firstNames)
	domain := randomPickWithDefault(g.Domains, emailDomains)
	return fmt.Sprintf("%s.%s_%d@%s",
		strings.ToLower(name),
		strings.ToLower(randomPick(lastNames)),
		row.RowIndex(),
		domain,
	), nil
}

type NameGenerator struct{}

func (g *NameGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return randomPick(firstNames) + " " + randomPick(lastNames), nil
}

type CityGenerator struct{}

func (g *CityGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return randomPick(cities), nil
}

type CountryGenerator struct{}

func (g *CountryGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return randomPick(countries), nil
}

type PhoneGenerator struct{}

func (g *PhoneGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000000))
	digits := fmt.Sprintf("%09d", n.Int64())
	return fmt.Sprintf("+234-%s-%s-%s", digits[:3], digits[3:6], digits[6:]), nil
}

type AddressGenerator struct{}

func (g *AddressGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	n, _ := rand.Int(rand.Reader, big.NewInt(999))
	number := n.Int64() + 1
	return fmt.Sprintf("%d %s %s", number, randomPick(streetNames), randomPick(streetTypes)), nil
}

type CompanyGenerator struct{}

func (g *CompanyGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return randomPick(companies), nil
}

type JobTitleGenerator struct{}

func (g *JobTitleGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return randomPick(jobTitles), nil
}

type BoolGenerator struct{}

func (g *BoolGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	n, _ := rand.Int(rand.Reader, big.NewInt(2))
	return n.Int64() == 1, nil
}

type UUIDGenerator struct{}

func (g *UUIDGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return newUUID(), nil
}

type EnumPicker struct {
	Values []string
}

func (g *EnumPicker) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return randomPick(g.Values), nil
}

type DateGenerator struct {
	MinYear int
	MaxYear int
}

func (g *DateGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	yearDelta := g.MaxYear - g.MinYear
	y, _ := rand.Int(rand.Reader, big.NewInt(int64(yearDelta+1)))
	m, _ := rand.Int(rand.Reader, big.NewInt(12))
	d, _ := rand.Int(rand.Reader, big.NewInt(28))
	return fmt.Sprintf("%04d-%02d-%02d",
		g.MinYear+int(y.Int64()),
		m.Int64()+1,
		d.Int64()+1,
	), nil
}

type TimestampGenerator struct{}

func (g *TimestampGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	y, _ := rand.Int(rand.Reader, big.NewInt(3))
	m, _ := rand.Int(rand.Reader, big.NewInt(12))
	d, _ := rand.Int(rand.Reader, big.NewInt(28))
	h, _ := rand.Int(rand.Reader, big.NewInt(24))
	mi, _ := rand.Int(rand.Reader, big.NewInt(60))
	s, _ := rand.Int(rand.Reader, big.NewInt(60))
	return fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02dZ",
		2024+int(y.Int64()), m.Int64()+1, d.Int64()+1,
		h.Int64(), mi.Int64(), s.Int64(),
	), nil
}

type NumericGenerator struct {
	Min int64
	Max int64
}

func (g *NumericGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	delta := g.Max - g.Min + 1
	n, err := rand.Int(rand.Reader, big.NewInt(delta))
	if err != nil {
		return nil, err
	}
	cents := g.Min + n.Int64()
	return float64(cents) / 100, nil
}

type JSONGenerator struct{}

func (g *JSONGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return "{}", nil
}

type ByteaGenerator struct{}

func (g *ByteaGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	b := make([]byte, 32)
	rand.Read(b)
	return b, nil
}

type URLGenerator struct{}

func (g *URLGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return fmt.Sprintf("https://%s.com/%s",
		strings.ToLower(randomPick(companies)),
		strings.ToLower(randomPick(loremWords)),
	), nil
}

type FKResolver struct {
	Table string
	pool  *FKPool
}

func (g *FKResolver) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return g.pool.Pick(g.Table)
}

func newUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func randomPick(pool []string) string {
	idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(pool))))
	return pool[idx.Int64()]
}

func randomPickWithDefault(custom, fallback []string) string {
	if len(custom) > 0 {
		return randomPick(custom)
	}
	return randomPick(fallback)
}

func ResolveGenerator(col *schema.Column, hint schema.GeneratorHint, pool *FKPool) (gen.Generator, error) {
	if col.FKRef != nil {
		return &FKResolver{Table: col.FKRef.Table, pool: pool}, nil
	}

	switch col.Type {
	case schema.TypeSerial, schema.TypeBigSerial:
		return NewSerialGenerator(0), nil
	case schema.TypeEnum:
		if len(col.EnumValues) == 0 {
			return &LoremGenerator{MinWords: 1, MaxWords: 3}, nil
		}
		return &EnumPicker{Values: col.EnumValues}, nil
	}

	switch hint {
	case schema.HintEmail:
		return &EmailGenerator{}, nil
	case schema.HintName:
		return &NameGenerator{}, nil
	case schema.HintCity:
		return &CityGenerator{}, nil
	case schema.HintCountry:
		return &CountryGenerator{}, nil
	case schema.HintPhone:
		return &PhoneGenerator{}, nil
	case schema.HintAddress:
		return &AddressGenerator{}, nil
	case schema.HintCompany:
		return &CompanyGenerator{}, nil
	case schema.HintJobTitle:
		return &JobTitleGenerator{}, nil
	case schema.HintURL:
		return &URLGenerator{}, nil
	case schema.HintUUID:
		return &UUIDGenerator{}, nil
	case schema.HintCurrency:
		return &LoremGenerator{MinWords: 1, MaxWords: 1}, nil
	case schema.HintIP:
		return &LoremGenerator{MinWords: 1, MaxWords: 1}, nil
	case schema.HintNow:
		return &TimestampGenerator{}, nil
	}

	switch col.Type {
	case schema.TypeInteger, schema.TypeSmallInt:
		return &RandomIntGenerator{Min: 1, Max: 100000}, nil
	case schema.TypeBigInt:
		return &RandomIntGenerator{Min: 1, Max: 10000000}, nil
	case schema.TypeBoolean:
		return &BoolGenerator{}, nil
	case schema.TypeVarchar, schema.TypeChar, schema.TypeText:
		return &LoremGenerator{MinWords: 2, MaxWords: 8}, nil
	case schema.TypeUUID:
		return &UUIDGenerator{}, nil
	case schema.TypeDate:
		return &DateGenerator{MinYear: 2024, MaxYear: 2026}, nil
	case schema.TypeTimestamp, schema.TypeTimestamptz:
		return &TimestampGenerator{}, nil
	case schema.TypeNumeric, schema.TypeFloat, schema.TypeDouble, schema.TypeReal, schema.TypeMoney:
		return &NumericGenerator{Min: 100, Max: 999999}, nil
	case schema.TypeJSON, schema.TypeJSONB:
		return &JSONGenerator{}, nil
	case schema.TypeBytea:
		return &ByteaGenerator{}, nil
	case schema.TypeInet:
		return &LoremGenerator{MinWords: 1, MaxWords: 1}, nil
	default:
		return &LoremGenerator{MinWords: 2, MaxWords: 8}, nil
	}
}

func ResolveGenerators(columns []*schema.Column, hints map[string]schema.GeneratorHint, pool *FKPool) (map[string]gen.Generator, error) {
	gens := make(map[string]gen.Generator, len(columns))
	for _, col := range columns {
		hint := hints[col.Name]
		g, err := ResolveGenerator(col, hint, pool)
		if err != nil {
			return nil, fmt.Errorf("resolve generator for %s: %w", col.Name, err)
		}
		gens[col.Name] = g
	}
	return gens, nil
}
