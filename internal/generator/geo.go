package generator

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

	gen "github.com/tomiwa-a/seedling/pkg/generator"
)

type CityGenerator struct{}

func (g *CityGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return randomPick(cities), nil
}

type CountryGenerator struct{}

func (g *CountryGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return randomPick(countries), nil
}

type CountryCodeGenerator struct{}

func (g *CountryCodeGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return randomPick(countryCodes), nil
}

type AddressGenerator struct{}

func (g *AddressGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	n, _ := rand.Int(rand.Reader, big.NewInt(999))
	number := n.Int64() + 1
	return fmt.Sprintf("%d %s %s", number, randomPick(streetNames), randomPick(streetTypes)), nil
}

type LatitudeGenerator struct{}

func (g *LatitudeGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	n, _ := rand.Int(rand.Reader, big.NewInt(1800000))
	lat := float64(n.Int64()-900000) / 10000.0
	return fmt.Sprintf("%.6f", lat), nil
}

type LongitudeGenerator struct{}

func (g *LongitudeGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	n, _ := rand.Int(rand.Reader, big.NewInt(3600000))
	lon := float64(n.Int64()-1800000) / 10000.0
	return fmt.Sprintf("%.6f", lon), nil
}

type PostalCodeGenerator struct{}

func (g *PostalCodeGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	code, _ := rand.Int(rand.Reader, big.NewInt(100000))
	return fmt.Sprintf("%05d", code.Int64()), nil
}

type CompanyGenerator struct{}

func (g *CompanyGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return randomPick(companies), nil
}

type JobTitleGenerator struct{}

func (g *JobTitleGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return randomPick(jobTitles), nil
}
