package ledgerstore

import (
	"github.com/eywa-protocol/chain/core/states"
	"github.com/hashicorp/golang-lru"
)

const (
	STATE_CACHE_SIZE = 100000
)

type StateCache struct {
	stateCache *lru.ARCCache
}

// TODO: add state cache to state store

func NewStateCache() (*StateCache, error) {
	stateCache, err := lru.NewARC(STATE_CACHE_SIZE)
	if err != nil {
		return nil, err
	}
	return &StateCache{
		stateCache: stateCache,
	}, nil
}

func (c *StateCache) GetState(key []byte) states.StateValue {
	state, ok := c.stateCache.Get(string(key))
	if !ok {
		return nil
	}
	return state.(states.StateValue)
}

func (c *StateCache) AddState(key []byte, state states.StateValue) {
	c.stateCache.Add(string(key), state)
}

func (c *StateCache) DeleteState(key []byte) {
	c.stateCache.Remove(string(key))
}
