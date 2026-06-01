package generator

import (
	"context"
	"fmt"
	"math/rand"

	gen "github.com/tomiwa-a/seedling/pkg/generator"
)

type CityGenerator struct {
	rnd *rand.Rand
}

func (g *CityGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *CityGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	return randomPick(rnd, cities), nil
}

type CountryGenerator struct {
	rnd *rand.Rand
}

func (g *CountryGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *CountryGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	return randomPick(rnd, countries), nil
}

type CountryCodeGenerator struct {
	rnd *rand.Rand
}

func (g *CountryCodeGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *CountryCodeGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	return randomPick(rnd, countryCodes), nil
}

type AddressGenerator struct {
	rnd *rand.Rand
}

func (g *AddressGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *AddressGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	number := rnd.Int63n(999) + 1
	return fmt.Sprintf("%d %s %s", number, randomPick(rnd, streetNames), randomPick(rnd, streetTypes)), nil
}

type LatitudeGenerator struct {
	rnd *rand.Rand
}

func (g *LatitudeGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *LatitudeGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	lat := float64(rnd.Int63n(1800000)-900000) / 10000.0
	return fmt.Sprintf("%.6f", lat), nil
}

type LongitudeGenerator struct {
	rnd *rand.Rand
}

func (g *LongitudeGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *LongitudeGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	lon := float64(rnd.Int63n(3600000)-1800000) / 10000.0
	return fmt.Sprintf("%.6f", lon), nil
}

type PostalCodeGenerator struct {
	rnd *rand.Rand
}

func (g *PostalCodeGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *PostalCodeGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	return fmt.Sprintf("%05d", rnd.Int63n(100000)), nil
}

type CompanyGenerator struct {
	rnd *rand.Rand
}

func (g *CompanyGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *CompanyGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	return randomPick(rnd, companies), nil
}

type JobTitleGenerator struct {
	rnd *rand.Rand
}

func (g *JobTitleGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *JobTitleGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	return randomPick(rnd, jobTitles), nil
}
