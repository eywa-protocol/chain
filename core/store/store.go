package store

import (
	"github.com/eywa-protocol/bls-crypto/bls"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/states"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/store/overlaydb"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/types"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/native/event"
	cstates "gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/native/states"
)

type ExecuteResult struct {
	WriteSet        *overlaydb.MemDB
	CrossHashes     []common.Uint256
	CrossStatesRoot common.Uint256
	Hash            common.Uint256
	MerkleRoot      common.Uint256
	Notify          []*event.ExecuteNotify
}

// LedgerStore provides func with store package.
type LedgerStore interface {
	InitLedgerStoreWithGenesisBlock(genesisblock *types.Block, defaultEpoch []bls.PublicKey) error
	Close() error
	AddHeaders(headers []*types.Header) error
	AddBlock(block *types.Block, stateMerkleRoot common.Uint256) error
	ExecuteBlock(b *types.Block) (ExecuteResult, error)   // called by consensus
	SubmitBlock(b *types.Block, exec ExecuteResult) error // called by consensus
	GetStateMerkleRoot(height uint32) (result common.Uint256, err error)
	GetCrossStateRoot(height uint32) (result common.Uint256, err error)
	GetCurrentBlockHash() common.Uint256
	GetCurrentBlockHeight() uint32
	GetCurrentHeaderHeight() uint32
	GetCurrentHeaderHash() common.Uint256
	GetBlockHash(height uint32) common.Uint256
	GetHeaderByHash(blockHash common.Uint256) (*types.Header, error)
	GetHeaderByHeight(height uint32) (*types.Header, error)
	GetBlockByHash(blockHash common.Uint256) (*types.Block, error)
	GetBlockByHeight(height uint32) (*types.Block, error)
	GetTransaction(txHash common.Uint256) (*types.Transaction, uint32, error)
	IsContainBlock(blockHash common.Uint256) (bool, error)
	IsContainTransaction(txHash common.Uint256) (bool, error)
	GetBlockRootWithPreBlockHashes(startHeight uint32, txRoots []common.Uint256) common.Uint256
	GetMerkleProof(raw []byte, m, n uint32) ([]byte, error)
	GetCrossStatesProof(height uint32, key []byte) ([]byte, error)
	GetEpochState() (*states.EpochState, error)
	GetStorageItem(key *states.StorageKey) (*states.StorageItem, error)
	PreExecuteContract(tx *types.Transaction) (*cstates.PreExecResult, error)
	GetEventNotifyByTx(tx common.Uint256) (*event.ExecuteNotify, error)
	GetEventNotifyByBlock(height uint32) ([]*event.ExecuteNotify, error)
}
