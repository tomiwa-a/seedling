package generator

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
)

type FKPool struct {
	mu       sync.RWMutex
	pools    map[string][]any
	consumed map[string]map[any]bool
}

func NewFKPool() *FKPool {
	return &FKPool{
		pools:    make(map[string][]any),
		consumed: make(map[string]map[any]bool),
	}
}

func (p *FKPool) Add(table string, pk any) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pools[table] = append(p.pools[table], pk)
}

func (p *FKPool) Pick(table string) (any, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	pool, ok := p.pools[table]
	if !ok || len(pool) == 0 {
		return nil, fmt.Errorf("fkpool: no rows available for table %q", table)
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(pool))))
	if err != nil {
		return nil, err
	}
	return pool[n.Int64()], nil
}

func (p *FKPool) Consume(table string, consumerKey string) (any, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	pool, ok := p.pools[table]
	if !ok || len(pool) == 0 {
		return nil, fmt.Errorf("fkpool: no rows available for table %q", table)
	}

	usedKey := consumerKey
	if _, ok := p.consumed[usedKey]; !ok {
		p.consumed[usedKey] = make(map[any]bool)
	}
	used := p.consumed[usedKey]

	if len(used) >= len(pool) {
		return nil, fmt.Errorf("fkpool: all rows for table %q already consumed by %q", table, consumerKey)
	}

	shuffle, _ := rand.Int(rand.Reader, big.NewInt(int64(len(pool))))
	start := shuffle.Int64()
	for i := range pool {
		idx := (start + int64(i)) % int64(len(pool))
		val := pool[idx]
		if !used[val] {
			used[val] = true
			return val, nil
		}
	}

	return nil, fmt.Errorf("fkpool: no unconsumed rows for table %q for %q", table, consumerKey)
}

func (p *FKPool) Count(table string) int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.pools[table])
}

func (p *FKPool) Tables() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	tables := make([]string, 0, len(p.pools))
	for t := range p.pools {
		tables = append(tables, t)
	}
	return tables
}
