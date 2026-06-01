package generator

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
)

type FKPool struct {
	mu    sync.RWMutex
	pools map[string][]any
}

func NewFKPool() *FKPool {
	return &FKPool{
		pools: make(map[string][]any),
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
