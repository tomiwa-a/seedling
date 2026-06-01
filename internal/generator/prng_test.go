package generator

import (
	"bytes"
	"testing"
)

func TestChaCha8Deterministic(t *testing.T) {
	key := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
		17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	nonce := [12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}

	c1 := NewChaCha8(key, nonce, 0)
	c2 := NewChaCha8(key, nonce, 0)

	buf1 := make([]byte, 256)
	buf2 := make([]byte, 256)
	c1.Read(buf1)
	c2.Read(buf2)

	if !bytes.Equal(buf1, buf2) {
		t.Error("same key+nonce should produce same output")
	}
}

func TestChaCha8DifferentNonce(t *testing.T) {
	key := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
		17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	nonce1 := [12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	nonce2 := [12]byte{12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	c1 := NewChaCha8(key, nonce1, 0)
	c2 := NewChaCha8(key, nonce2, 0)

	buf1 := make([]byte, 256)
	buf2 := make([]byte, 256)
	c1.Read(buf1)
	c2.Read(buf2)

	if bytes.Equal(buf1, buf2) {
		t.Error("different nonces should produce different output")
	}
}

func TestChaCha8DifferentKey(t *testing.T) {
	key1 := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
		17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	key2 := [32]byte{32, 31, 30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17,
		16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	nonce := [12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}

	c1 := NewChaCha8(key1, nonce, 0)
	c2 := NewChaCha8(key2, nonce, 0)

	buf1 := make([]byte, 256)
	buf2 := make([]byte, 256)
	c1.Read(buf1)
	c2.Read(buf2)

	if bytes.Equal(buf1, buf2) {
		t.Error("different keys should produce different output")
	}
}

func TestSeedDeriverDeterministic(t *testing.T) {
	d1 := NewSeedDeriver(42)
	d2 := NewSeedDeriver(42)

	if d1.TableSeed("users") != d2.TableSeed("users") {
		t.Error("same seed should derive same table seed")
	}
	if d1.ColumnSeed("users", "email") != d2.ColumnSeed("users", "email") {
		t.Error("same seed should derive same column seed")
	}
	if d1.RowSeed("users", "email", 0) != d2.RowSeed("users", "email", 0) {
		t.Error("same seed should derive same row seed")
	}
}

func TestSeedDeriverDifferentSeeds(t *testing.T) {
	d1 := NewSeedDeriver(42)
	d2 := NewSeedDeriver(99)

	if d1.TableSeed("users") == d2.TableSeed("users") {
		t.Error("different master seeds should derive different table seeds")
	}
}

func TestSeedDeriverDifferentTables(t *testing.T) {
	d := NewSeedDeriver(42)

	if d.TableSeed("users") == d.TableSeed("orders") {
		t.Error("different tables should derive different seeds")
	}
}

func TestSeedDeriverDifferentColumns(t *testing.T) {
	d := NewSeedDeriver(42)

	if d.ColumnSeed("users", "email") == d.ColumnSeed("users", "name") {
		t.Error("different columns should derive different seeds")
	}
}

func TestSeedDeriverDifferentRows(t *testing.T) {
	d := NewSeedDeriver(42)

	if d.RowSeed("users", "email", 0) == d.RowSeed("users", "email", 1) {
		t.Error("different rows should derive different seeds")
	}
}
