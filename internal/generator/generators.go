package generator

import (
	"context"
	"fmt"
	"math/rand"

	gen "github.com/tomiwa-a/seedling/pkg/generator"
	"github.com/tomiwa-a/seedling/pkg/schema"
)

func randomPick(rnd *rand.Rand, pool []string) string {
	return pool[rnd.Int63n(int64(len(pool)))]
}

func randomPickWithDefault(rnd *rand.Rand, custom, fallback []string) string {
	if len(custom) > 0 {
		return randomPick(rnd, custom)
	}
	return randomPick(rnd, fallback)
}

func randomInt(rnd *rand.Rand, min, max int64) int64 {
	return min + rnd.Int63n(max-min+1)
}

func newUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

type FKResolver struct {
	Table string
	pool  *FKPool
}

func (g *FKResolver) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return g.pool.Pick(g.Table)
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
		return &CurrencyGenerator{}, nil
	case schema.HintIP:
		return &IPv4Generator{}, nil
	case schema.HintNow:
		return &NowGenerator{}, nil
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
	case schema.TypeTime:
		return &TimeGenerator{}, nil
	case schema.TypeNumeric, schema.TypeFloat, schema.TypeDouble, schema.TypeReal, schema.TypeMoney:
		return &NumericGenerator{Min: 100, Max: 999999}, nil
	case schema.TypeJSON, schema.TypeJSONB:
		return &JSONGenerator{}, nil
	case schema.TypeBytea:
		return &ByteaGenerator{}, nil
	case schema.TypeInet:
		return &IPv4Generator{}, nil
	default:
		return &LoremGenerator{MinWords: 2, MaxWords: 8}, nil
	}
}

type Seeded interface {
	SetRand(rnd *rand.Rand)
}

func NewSeededRand(seed uint64) *rand.Rand {
	return rand.New(rand.NewSource(int64(seed)))
}

func ResolveGenerators(columns []*schema.Column, hints map[string]schema.GeneratorHint, pool *FKPool, masterSeed uint64, tableName string) (map[string]gen.Generator, error) {
	deriver := NewSeedDeriver(masterSeed)
	gens := make(map[string]gen.Generator, len(columns))
	for _, col := range columns {
		hint := hints[col.Name]
		if hint == "" || hint == schema.HintAuto {
			hint = col.Hint
		}
		g, err := ResolveGenerator(col, hint, pool)
		if err != nil {
			return nil, fmt.Errorf("resolve generator for %s: %w", col.Name, err)
		}
		if seeded, ok := g.(Seeded); ok {
			colSeed := deriver.ColumnSeed(tableName, col.Name)
			rnd := rand.New(rand.NewSource(int64(colSeed)))
			seeded.SetRand(rnd)
		}
		gens[col.Name] = g
	}
	return gens, nil
}
