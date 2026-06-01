package generator

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"sync/atomic"
	"time"

	gen "github.com/tomiwa-a/seedling/pkg/generator"
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

type PhoneGenerator struct{}

func (g *PhoneGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000000))
	digits := fmt.Sprintf("%09d", n.Int64())
	return fmt.Sprintf("+234-%s-%s-%s", digits[:3], digits[3:6], digits[6:]), nil
}

type UUIDGenerator struct{}

func (g *UUIDGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return newUUID(), nil
}

type ULIDGenerator struct{}

func (g *ULIDGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return newULID(), nil
}

const base32Alphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

func newULID() string {
	t := time.Now().UnixMilli()
	ts := make([]byte, 6)
	ts[0] = byte(t >> 40)
	ts[1] = byte(t >> 32)
	ts[2] = byte(t >> 24)
	ts[3] = byte(t >> 16)
	ts[4] = byte(t >> 8)
	ts[5] = byte(t)

	randBytes := make([]byte, 10)
	rand.Read(randBytes)

	return encodeBits(ts, 48) + encodeBits(randBytes, 80)
}

func encodeBits(data []byte, bits int) string {
	result := make([]byte, 0, (bits+4)/5)
	acc := 0
	bitsLeft := 0
	for _, b := range data {
		acc = (acc << 8) | int(b)
		bitsLeft += 8
		for bitsLeft >= 5 {
			bitsLeft -= 5
			result = append(result, base32Alphabet[(acc>>bitsLeft)&0x1f])
		}
	}
	if bitsLeft > 0 {
		result = append(result, base32Alphabet[(acc<<(5-bitsLeft))&0x1f])
	}
	return string(result)
}

type UsernameGenerator struct{}

func (g *UsernameGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	n, _ := rand.Int(rand.Reader, big.NewInt(10000))
	return fmt.Sprintf("%s%s%d",
		strings.ToLower(randomPick(firstNames)),
		strings.ToLower(randomPick(lastNames)),
		n.Int64(),
	), nil
}

type StringGenerator struct {
	Length int
}

func (g *StringGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	length := g.Length
	if length <= 0 {
		length = 10
	}
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[idx.Int64()]
	}
	return string(b), nil
}

type EnumPicker struct {
	Values []string
}

func (g *EnumPicker) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return randomPick(g.Values), nil
}

type BoolGenerator struct{}

func (g *BoolGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	n, _ := rand.Int(rand.Reader, big.NewInt(2))
	return n.Int64() == 1, nil
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
