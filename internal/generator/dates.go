package generator

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	gen "github.com/tomiwa-a/seedling/pkg/generator"
)

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

type NowGenerator struct{}

func (g *NowGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z"), nil
}

type TimeAgoGenerator struct {
	MaxDays int
}

func (g *TimeAgoGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	maxDays := g.MaxDays
	if maxDays <= 0 {
		maxDays = 365
	}
	daysAgo, _ := rand.Int(rand.Reader, big.NewInt(int64(maxDays)))
	t := time.Now().UTC().AddDate(0, 0, -int(daysAgo.Int64()))
	return t.Format("2006-01-02T15:04:05Z"), nil
}

type BusinessDaysGenerator struct {
	MinYear int
	MaxYear int
}

func (g *BusinessDaysGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	yearDelta := g.MaxYear - g.MinYear
	y, _ := rand.Int(rand.Reader, big.NewInt(int64(yearDelta+1)))
	m, _ := rand.Int(rand.Reader, big.NewInt(12))
	d, _ := rand.Int(rand.Reader, big.NewInt(28))
	t := time.Date(g.MinYear+int(y.Int64()), time.Month(m.Int64()+1), int(d.Int64()+1), 0, 0, 0, 0, time.UTC)
	for t.Weekday() == time.Saturday || t.Weekday() == time.Sunday {
		t = t.AddDate(0, 0, 1)
	}
	return t.Format("2006-01-02"), nil
}
