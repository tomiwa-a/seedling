package generator

import (
	"context"
	"fmt"
	"math/rand"
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
	rnd      *rand.Rand
}

func (g *LoremGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *LoremGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	delta := g.MaxWords - g.MinWords + 1
	count := g.MinWords + rnd.Intn(delta)

	words := make([]string, count)
	for i := 0; i < count; i++ {
		words[i] = loremWords[rnd.Intn(len(loremWords))]
	}
	return strings.Join(words, " "), nil
}

type EmailGenerator struct {
	Domains []string
	rnd     *rand.Rand
}

func (g *EmailGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *EmailGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	name := randomPick(rnd, firstNames)
	domain := randomPickWithDefault(rnd, g.Domains, emailDomains)
	return fmt.Sprintf("%s.%s_%d@%s",
		strings.ToLower(name),
		strings.ToLower(randomPick(rnd, lastNames)),
		row.RowIndex(),
		domain,
	), nil
}

type NameGenerator struct {
	rnd *rand.Rand
}

func (g *NameGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *NameGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	return randomPick(rnd, firstNames) + " " + randomPick(rnd, lastNames), nil
}

type PhoneGenerator struct {
	rnd *rand.Rand
}

func (g *PhoneGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *PhoneGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	digits := fmt.Sprintf("%09d", rnd.Int63n(1000000000))
	return fmt.Sprintf("+234-%s-%s-%s", digits[:3], digits[3:6], digits[6:]), nil
}

type UUIDGenerator struct {
	rnd *rand.Rand
}

func (g *UUIDGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *UUIDGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	b := make([]byte, 16)
	rnd.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

type ULIDGenerator struct {
	rnd *rand.Rand
}

func (g *ULIDGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *ULIDGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	t := time.Now().UnixMilli()
	ts := make([]byte, 6)
	ts[0] = byte(t >> 40)
	ts[1] = byte(t >> 32)
	ts[2] = byte(t >> 24)
	ts[3] = byte(t >> 16)
	ts[4] = byte(t >> 8)
	ts[5] = byte(t)

	randBytes := make([]byte, 10)
	for i := range randBytes {
		randBytes[i] = byte(rnd.Intn(256))
	}

	return encodeBits(ts, 48) + encodeBits(randBytes, 80), nil
}

type UsernameGenerator struct {
	rnd *rand.Rand
}

func (g *UsernameGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *UsernameGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	return fmt.Sprintf("%s%s%d",
		strings.ToLower(randomPick(rnd, firstNames)),
		strings.ToLower(randomPick(rnd, lastNames)),
		rnd.Int63n(10000),
	), nil
}

type StringGenerator struct {
	Length int
	rnd    *rand.Rand
}

func (g *StringGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *StringGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	length := g.Length
	if length <= 0 {
		length = 10
	}
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rnd.Intn(len(charset))]
	}
	return string(b), nil
}

type EnumPicker struct {
	Values []string
	rnd    *rand.Rand
}

func (g *EnumPicker) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *EnumPicker) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	return randomPick(rnd, g.Values), nil
}

type BoolGenerator struct {
	rnd *rand.Rand
}

func (g *BoolGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *BoolGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	return rnd.Intn(2) == 1, nil
}

type JSONGenerator struct{}

func (g *JSONGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return "{}", nil
}

type ByteaGenerator struct {
	rnd *rand.Rand
}

func (g *ByteaGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *ByteaGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	b := make([]byte, 32)
	rnd.Read(b)
	return b, nil
}

type URLGenerator struct {
	rnd *rand.Rand
}

func (g *URLGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *URLGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	return fmt.Sprintf("https://%s.com/%s",
		strings.ToLower(randomPick(rnd, companies)),
		strings.ToLower(randomPick(rnd, loremWords)),
	), nil
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
	for i := range randBytes {
		randBytes[i] = byte(time.Now().UnixNano() & 0xff)
		time.Sleep(time.Nanosecond)
	}

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
