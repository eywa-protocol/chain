package store

import (
	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/core/payload"
	"github.com/eywa-protocol/chain/core/states"
	"github.com/eywa-protocol/chain/core/store/overlaydb"
	"github.com/eywa-protocol/chain/core/types"
	"github.com/eywa-protocol/chain/native/event"
	cstates "github.com/eywa-protocol/chain/native/states"
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
	InitLedgerStoreWithGenesisBlock(genesisblock *types.Block) error
	Close() error
	AddHeaders(headers []*types.Header) error
	AddBlock(block *types.Block, stateMerkleRoot common.Uint256) error
	ExecuteBlock(b *types.Block) (ExecuteResult, error)   // called by consensus
	SubmitBlock(b *types.Block, exec ExecuteResult) error // called by consensus
	GetStateMerkleRoot(height uint64) (result common.Uint256, err error)
	GetCrossStateRoot(height uint64) (result common.Uint256, err error)
	GetCurrentBlockHash() common.Uint256
	GetCurrentBlockHeight() uint64
	GetCurrentHeaderHeight() uint64
	GetCurrentHeaderHash() common.Uint256
	GetBlockHash(height uint64) common.Uint256
	GetHeaderByHash(blockHash common.Uint256) (*types.Header, error)
	GetHeaderByHeight(height uint64) (*types.Header, error)
	GetBlockByHash(blockHash common.Uint256) (*types.Block, error)
	GetBlockByHeight(height uint64) (*types.Block, error)
	GetTransaction(txHash common.Uint256) (payload.Payload, uint64, error)
	GetTransactionByReqId(reqId [32]byte) (payload.Payload, uint64, error)
	GetRequestState(reqId [32]byte) (payload.ReqState, error)
	IsContainBlock(blockHash common.Uint256) (bool, error)
	IsContainTransaction(txHash common.Uint256) (bool, error)
	GetBlockRootWithPreBlockHashes(startHeight uint64, txRoots []common.Uint256) common.Uint256
	GetMerkleProof(raw []byte, m, n uint64) ([]byte, error)
	GetCrossStatesProof(height uint64, key []byte) ([]byte, error)
	GetEpochState() (*states.EpochState, error)
	GetStorageItem(key *states.StorageKey) (*states.StorageItem, error)
	PreExecuteContract(tx payload.Payload) (*cstates.PreExecResult, error)
	GetEventNotifyByTx(tx common.Uint256) (*event.ExecuteNotify, error)
	GetEventNotifyByBlock(height uint64) ([]*event.ExecuteNotify, error)
	GetProcessedHeight() uint64
	SetProcessedHeight(srcBlockHeight uint64)
}
