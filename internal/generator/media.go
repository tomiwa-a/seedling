package generator

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	gen "github.com/tomiwa-a/seedling/pkg/generator"
)

type FileNameGenerator struct {
	rnd *rand.Rand
}

func (g *FileNameGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *FileNameGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	name := strings.ToLower(randomPick(rnd, loremWords))
	ext := randomPick(rnd, fileExtensions)
	return fmt.Sprintf("%s%s", name, ext), nil
}

type ImageURLGenerator struct {
	rnd *rand.Rand
}

func (g *ImageURLGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *ImageURLGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	return fmt.Sprintf("https://picsum.photos/seed/%d/400/300", rnd.Int63n(10000)), nil
}
