package generator

import (
	"context"
	"fmt"
	"math/rand"

	gen "github.com/tomiwa-a/seedling/pkg/generator"
)

type IPv4Generator struct {
	rnd *rand.Rand
}

func (g *IPv4Generator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *IPv4Generator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	return fmt.Sprintf("%d.%d.%d.%d",
		rnd.Intn(256), rnd.Intn(256), rnd.Intn(256), rnd.Intn(256)), nil
}

type IPv6Generator struct {
	rnd *rand.Rand
}

func (g *IPv6Generator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *IPv6Generator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	return fmt.Sprintf("%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x",
		rnd.Intn(256), rnd.Intn(256), rnd.Intn(256), rnd.Intn(256),
		rnd.Intn(256), rnd.Intn(256), rnd.Intn(256), rnd.Intn(256),
		rnd.Intn(256), rnd.Intn(256), rnd.Intn(256), rnd.Intn(256),
		rnd.Intn(256), rnd.Intn(256), rnd.Intn(256), rnd.Intn(256)), nil
}

type MACGenerator struct {
	rnd *rand.Rand
}

func (g *MACGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *MACGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		rnd.Intn(256), rnd.Intn(256), rnd.Intn(256),
		rnd.Intn(256), rnd.Intn(256), rnd.Intn(256)), nil
}

type UserAgentGenerator struct {
	rnd *rand.Rand
}

func (g *UserAgentGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *UserAgentGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	return randomPick(rnd, userAgents), nil
}
