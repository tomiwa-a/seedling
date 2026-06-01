package generator

import (
	"context"
	"crypto/rand"
	"fmt"

	gen "github.com/tomiwa-a/seedling/pkg/generator"
)

type IPv4Generator struct{}

func (g *IPv4Generator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%d.%d.%d.%d", b[0], b[1], b[2], b[3]), nil
}

type IPv6Generator struct{}

func (g *IPv6Generator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x",
		b[0], b[1], b[2], b[3], b[4], b[5], b[6], b[7],
		b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15]), nil
}

type MACGenerator struct{}

func (g *MACGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	b := make([]byte, 6)
	rand.Read(b)
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		b[0], b[1], b[2], b[3], b[4], b[5]), nil
}

type UserAgentGenerator struct{}

func (g *UserAgentGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return randomPick(userAgents), nil
}
