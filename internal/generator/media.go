package generator

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	gen "github.com/tomiwa-a/seedling/pkg/generator"
)

type FileNameGenerator struct{}

func (g *FileNameGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	name := strings.ToLower(randomPick(loremWords))
	ext := randomPick(fileExtensions)
	return fmt.Sprintf("%s%s", name, ext), nil
}

type ImageURLGenerator struct{}

func (g *ImageURLGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	n, _ := rand.Int(rand.Reader, big.NewInt(10000))
	return fmt.Sprintf("https://picsum.photos/seed/%d/400/300", n.Int64()), nil
}
