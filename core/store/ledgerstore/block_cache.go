package ledgerstore

import (
	"fmt"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/core/payload"
	"github.com/eywa-protocol/chain/core/types"
	lru "github.com/hashicorp/golang-lru"
)

const (
	BlockCacheSize       = 10    // Block cache size
	TransactionCacheSize = 10000 // Transaction cache size
)

// TransactionCacheValue value of transaction cache
type TransactionCacheValue struct {
	Tx     payload.Payload
	Height uint64
}

// BlockCache with block cache and transaction hash
type BlockCache struct {
	blockCache       *lru.ARCCache
	transactionCache *lru.ARCCache
	requestIdCache   *lru.ARCCache
}

// NewBlockCache return BlockCache instance
func NewBlockCache() (*BlockCache, error) {
	blockCache, err := lru.NewARC(BlockCacheSize)
	if err != nil {
		return nil, fmt.Errorf("NewARC block error %s", err)
	}
	transactionCache, err := lru.NewARC(TransactionCacheSize)
	if err != nil {
		return nil, fmt.Errorf("NewARC header error %s", err)
	}
	requestIdCache, err := lru.NewARC(TransactionCacheSize)
	if err != nil {
		return nil, fmt.Errorf("NewARC request id error %s", err)
	}
	return &BlockCache{
		blockCache:       blockCache,
		transactionCache: transactionCache,
		requestIdCache:   requestIdCache,
	}, nil
}

// AddBlock to cache
func (c *BlockCache) AddBlock(block *types.Block) {
	blockHash := block.Hash()
	c.blockCache.Add(string(blockHash.ToArray()), block)
}

// GetBlock return block by block hash from cache
func (c *BlockCache) GetBlock(blockHash common.Uint256) *types.Block {
	block, ok := c.blockCache.Get(string(blockHash.ToArray()))
	if !ok {
		return nil
	}
	return block.(*types.Block)
}

// ContainBlock return whether block is in cache
func (c *BlockCache) ContainBlock(blockHash common.Uint256) bool {
	return c.blockCache.Contains(string(blockHash.ToArray()))
}

// AddTransaction add transaction to block cache
func (c *BlockCache) AddTransaction(payload payload.Payload, height uint64) {
	tx := types.ToTransaction(payload)
	txHash := tx.Hash()
	value := &TransactionCacheValue{
		Tx:     payload,
		Height: height,
	}
	c.transactionCache.Add(string(txHash.ToArray()), value)
	if payload.RequestState() > 0 {
		reqId := payload.RequestId()
		c.requestIdCache.Add(string(reqId[:]), value)
	}
}

// GetTransaction return transaction by transaction hash from cache
func (c *BlockCache) GetTransaction(txHash common.Uint256) (payload.Payload, uint64) {
	value, ok := c.transactionCache.Get(string(txHash.ToArray()))
	if !ok {
		return nil, 0
	}
	txValue := value.(*TransactionCacheValue)
	return txValue.Tx, txValue.Height
}

// GetTransactionByReqId return transaction by request id from cache
func (c *BlockCache) GetTransactionByReqId(reqId [32]byte) (payload.Payload, uint64) {
	value, ok := c.requestIdCache.Get(string(reqId[:]))
	if !ok {
		return nil, 0
	}
	txValue := value.(*TransactionCacheValue)
	return txValue.Tx, txValue.Height
}

// ContainTransaction return whether transaction is in cache
func (c *BlockCache) ContainTransaction(txHash common.Uint256) bool {
	return c.transactionCache.Contains(string(txHash.ToArray()))
}

// ContainTransactionWithReqId return whether transaction is in cache
func (c *BlockCache) ContainTransactionWithReqId(reqId [32]byte) bool {
	return c.transactionCache.Contains(string(reqId[:]))
}
