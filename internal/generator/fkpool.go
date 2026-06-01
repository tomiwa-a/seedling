package generator

import (
	"fmt"
	"math/rand"
	"sync"
)

type FKStrategy int

const (
	FKStrategyUniform    FKStrategy = iota
	FKStrategySequential
	FKStrategyWeighted
	FKStrategyExclusive
)

type FKPool struct {
	mu         sync.RWMutex
	pools      map[string][]any
	consumed   map[string]map[any]bool
	rnd        *rand.Rand
	strategies map[string]FKStrategy
	roundRobin map[string]int
	weights    map[string][]float64
}

func NewFKPool() *FKPool {
	return &FKPool{
		pools:      make(map[string][]any),
		consumed:   make(map[string]map[any]bool),
		rnd:        rand.New(rand.NewSource(0)),
		strategies: make(map[string]FKStrategy),
		roundRobin: make(map[string]int),
		weights:    make(map[string][]float64),
	}
}

func (p *FKPool) SetRand(rnd *rand.Rand) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.rnd = rnd
}

func (p *FKPool) SetStrategy(table string, strategy FKStrategy) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.strategies[table] = strategy
}

func (p *FKPool) SetWeights(table string, weights []float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.weights[table] = weights
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

	strategy := p.strategies[table]
	switch strategy {
	case FKStrategySequential:
		idx := p.roundRobin[table] % len(pool)
		p.roundRobin[table] = idx + 1
		return pool[idx], nil
	case FKStrategyWeighted:
		return p.pickWeighted(table, pool), nil
	default:
		return pool[p.rnd.Int63n(int64(len(pool)))], nil
	}
}

func (p *FKPool) pickWeighted(table string, pool []any) any {
	w := p.weights[table]
	if len(w) == 0 {
		return pool[p.rnd.Int63n(int64(len(pool)))]
	}

	total := 0.0
	for _, weight := range w {
		total += weight
	}

	r := p.rnd.Float64() * total
	cumulative := 0.0
	for i, weight := range w {
		cumulative += weight
		if r < cumulative {
			if i < len(pool) {
				return pool[i]
			}
			return pool[len(pool)-1]
		}
	}
	return pool[len(pool)-1]
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

	start := p.rnd.Int63n(int64(len(pool)))
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
