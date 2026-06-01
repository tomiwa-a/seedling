package generator

import (
	"crypto/sha256"
	"encoding/binary"
)

type ChaCha8 struct {
	state [16]uint32
	buf   [64]byte
	pos   int
}

func NewChaCha8(key [32]byte, nonce [12]byte, counter uint64) *ChaCha8 {
	c := &ChaCha8{}
	c.state[0] = 0x61707865
	c.state[1] = 0x3320646e
	c.state[2] = 0x79622d32
	c.state[3] = 0x6b206574

	for i := 0; i < 8; i++ {
		c.state[4+i] = binary.LittleEndian.Uint32(key[i*4 : i*4+4])
	}
	c.state[12] = uint32(counter)
	c.state[13] = uint32(counter >> 32)
	c.state[14] = binary.LittleEndian.Uint32(nonce[0:4])
	c.state[15] = binary.LittleEndian.Uint32(nonce[4:8])

	c.pos = 64
	return c
}

func (c *ChaCha8) Read(p []byte) (int, error) {
	n := 0
	for n < len(p) {
		if c.pos >= 64 {
			c.block()
			c.pos = 0
		}
		remaining := 64 - c.pos
		toCopy := len(p) - n
		if toCopy > remaining {
			toCopy = remaining
		}
		copy(p[n:n+toCopy], c.buf[c.pos:c.pos+toCopy])
		n += toCopy
		c.pos += toCopy
	}
	return n, nil
}

func (c *ChaCha8) block() {
	var working [16]uint32
	copy(working[:], c.state[:])

	for i := 0; i < 10; i++ {
		quarterRound(&working, 0, 4, 8, 12)
		quarterRound(&working, 1, 5, 9, 13)
		quarterRound(&working, 2, 6, 10, 14)
		quarterRound(&working, 3, 7, 11, 15)
		quarterRound(&working, 0, 5, 10, 15)
		quarterRound(&working, 1, 6, 11, 12)
		quarterRound(&working, 2, 7, 8, 13)
		quarterRound(&working, 3, 4, 9, 14)
	}

	for i := 0; i < 16; i++ {
		working[i] += c.state[i]
		binary.LittleEndian.PutUint32(c.buf[i*4:i*4+4], working[i])
	}

	c.state[12]++
	if c.state[12] == 0 {
		c.state[13]++
	}
}

func quarterRound(state *[16]uint32, a, b, c, d uint32) {
	state[a] += state[b]
	state[d] ^= state[a]
	state[d] = (state[d] << 16) | (state[d] >> 16)

	state[c] += state[d]
	state[b] ^= state[c]
	state[b] = (state[b] << 12) | (state[b] >> 20)

	state[a] += state[b]
	state[d] ^= state[a]
	state[d] = (state[d] << 8) | (state[d] >> 24)

	state[c] += state[d]
	state[b] ^= state[c]
	state[b] = (state[b] << 7) | (state[b] >> 25)
}

type SeedDeriver struct {
	masterSeed uint64
}

func NewSeedDeriver(masterSeed uint64) *SeedDeriver {
	return &SeedDeriver{masterSeed: masterSeed}
}

func (d *SeedDeriver) TableSeed(tableName string) uint64 {
	return d.derive(d.masterSeed, tableName)
}

func (d *SeedDeriver) ColumnSeed(tableName, columnName string) uint64 {
	return d.derive(d.TableSeed(tableName), columnName)
}

func (d *SeedDeriver) RowSeed(tableName, columnName string, rowIndex int64) uint64 {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(rowIndex))
	return d.derive(d.ColumnSeed(tableName, columnName), string(buf[:]))
}

func (d *SeedDeriver) derive(parent uint64, label string) uint64 {
	h := sha256.New()
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], parent)
	h.Write(buf[:])
	h.Write([]byte(label))
	sum := h.Sum(nil)
	return binary.LittleEndian.Uint64(sum[:8])
}

func DeriveKey(seed uint64) [32]byte {
	var key [32]byte
	h := sha256.New()
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], seed)
	h.Write(buf[:])
	h.Write([]byte("seedling-chacha8-key"))
	copy(key[:], h.Sum(nil))
	return key
}

func DeriveNonce(seed uint64, table, column string, row int64) [12]byte {
	var nonce [12]byte
	h := sha256.New()
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], seed)
	h.Write(buf[:])
	h.Write([]byte(table))
	h.Write([]byte(column))
	binary.LittleEndian.PutUint64(buf[:], uint64(row))
	h.Write(buf[:])
	sum := h.Sum(nil)
	copy(nonce[:], sum[:12])
	return nonce
}
