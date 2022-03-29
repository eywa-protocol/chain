package ledger

import (
	"bytes"
	"fmt"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/core/payload"
	"github.com/eywa-protocol/chain/core/states"
	"github.com/eywa-protocol/chain/core/store"
	"github.com/eywa-protocol/chain/core/store/ledgerstore"
	"github.com/eywa-protocol/chain/core/types"
	"github.com/eywa-protocol/chain/native/event"
	cstate "github.com/eywa-protocol/chain/native/states"
	"github.com/sirupsen/logrus"
)

type Ledger struct {
	ldgStore store.LedgerStore
	chainId  uint64
}

func NewLedger(dataDir string, chainId uint64) (*Ledger, error) {
	ldgStore, err := ledgerstore.NewLedgerStore(dataDir)
	if err != nil {
		return nil, fmt.Errorf("NewLedgerStore error %s", err)
	}
	return &Ledger{
		ldgStore: ldgStore,
		chainId:  chainId,
	}, nil
}

func (l *Ledger) GetStore() store.LedgerStore {
	return l.ldgStore
}

func (l *Ledger) GetChainId() uint64 {
	return l.chainId
}

func (l *Ledger) Init(genesisBlock *types.Block) error {
	err := l.ldgStore.InitLedgerStoreWithGenesisBlock(genesisBlock)
	if err != nil {
		return fmt.Errorf("InitLedgerStoreWithGenesisBlock error %s", err)
	}
	return nil
}

func (l *Ledger) AddHeaders(headers []*types.Header) error {
	return l.ldgStore.AddHeaders(headers)
}

func (l *Ledger) AddBlock(block *types.Block, stateMerkleRoot common.Uint256) error {
	err := l.ldgStore.AddBlock(block, stateMerkleRoot)
	if err != nil {
		logrus.Errorf("Ledger AddBlock BlockHeight:%d BlockHash:%x error:%s", block.Header.Height, block.Hash(), err)
	}
	return err
}

func (l *Ledger) ExecuteBlock(b *types.Block) (store.ExecuteResult, error) {
	return l.ldgStore.ExecuteBlock(b)
}

func (l *Ledger) SubmitBlock(b *types.Block, exec store.ExecuteResult) error {
	return l.ldgStore.SubmitBlock(b, exec)
}

func (l *Ledger) GetStateMerkleRoot(height uint64) (result common.Uint256, err error) {
	return l.ldgStore.GetStateMerkleRoot(height)
}

func (l *Ledger) GetCrossStateRoot(height uint64) (common.Uint256, error) {
	return l.ldgStore.GetCrossStateRoot(height)
}

func (l *Ledger) GetBlockRootWithPreBlockHashes(startHeight uint64, txRoots []common.Uint256) common.Uint256 {
	return l.ldgStore.GetBlockRootWithPreBlockHashes(startHeight, txRoots)
}

func (l *Ledger) GetBlockByHeight(height uint64) (*types.Block, error) {
	return l.ldgStore.GetBlockByHeight(height)
}

func (l *Ledger) GetBlockByHash(blockHash common.Uint256) (*types.Block, error) {
	return l.ldgStore.GetBlockByHash(blockHash)
}

func (l *Ledger) GetHeaderByHeight(height uint64) (*types.Header, error) {
	return l.ldgStore.GetHeaderByHeight(height)
}

func (l *Ledger) GetHeaderByHash(blockHash common.Uint256) (*types.Header, error) {
	return l.ldgStore.GetHeaderByHash(blockHash)
}

func (l *Ledger) GetBlockHash(height uint64) common.Uint256 {
	return l.ldgStore.GetBlockHash(height)
}

func (l *Ledger) GetTransaction(txHash common.Uint256) (payload.Payload, error) {
	tx, _, err := l.ldgStore.GetTransaction(txHash)
	return tx, err
}

func (l *Ledger) GetTransactionByReqId(reqId [32]byte) (payload.Payload, error) {
	tx, _, err := l.ldgStore.GetTransactionByReqId(reqId)
	return tx, err
}

func (l *Ledger) GetRequestState(reqId [32]byte) (payload.ReqState, error) {
	return l.ldgStore.GetRequestState(reqId)
}

func (l *Ledger) GetTransactionWithHeight(txHash common.Uint256) (payload.Payload, uint64, error) {
	return l.ldgStore.GetTransaction(txHash)
}

func (l *Ledger) GetCurrentBlockHeight() uint64 {
	return l.ldgStore.GetCurrentBlockHeight()
}

func (l *Ledger) GetCurrentBlockHash() common.Uint256 {
	return l.ldgStore.GetCurrentBlockHash()
}

func (l *Ledger) GetCurrentHeaderHeight() uint64 {
	return l.ldgStore.GetCurrentHeaderHeight()
}

func (l *Ledger) GetCurrentHeaderHash() common.Uint256 {
	return l.ldgStore.GetCurrentHeaderHash()
}

func (l *Ledger) IsContainTransaction(txHash common.Uint256) (bool, error) {
	return l.ldgStore.IsContainTransaction(txHash)
}

func (l *Ledger) IsContainBlock(blockHash common.Uint256) (bool, error) {
	return l.ldgStore.IsContainBlock(blockHash)
}

func (l *Ledger) GetCurrentStateRoot() (common.Uint256, error) {
	return common.Uint256{}, nil
}

func (l *Ledger) GetEpochState() (*states.EpochState, error) {
	return l.ldgStore.GetEpochState()
}

func (l *Ledger) GetStorageItem(codeHash common.Address, key []byte) ([]byte, error) {
	storageKey := &states.StorageKey{
		ContractAddress: codeHash,
		Key:             key,
	}
	storageItem, err := l.ldgStore.GetStorageItem(storageKey)
	if err != nil {
		return nil, err
	}
	return storageItem.Value, nil
}

func (l *Ledger) GetMerkleProof(proofHeight, rootHeight uint64) ([]byte, error) {
	blockHash := l.ldgStore.GetBlockHash(proofHeight)
	if bytes.Equal(blockHash.ToArray(), common.UINT256_EMPTY.ToArray()) {
		return nil, fmt.Errorf("GetBlockHash(%d) empty", proofHeight)
	}
	return l.ldgStore.GetMerkleProof(blockHash.ToArray(), proofHeight+1, rootHeight)
}

func (l *Ledger) GetCrossStatesProof(height uint64, key []byte) ([]byte, error) {
	return l.ldgStore.GetCrossStatesProof(height, key)
}

func (l *Ledger) PreExecuteContract(tx payload.Payload) (*cstate.PreExecResult, error) {
	return l.ldgStore.PreExecuteContract(tx)
}

func (l *Ledger) GetEventNotifyByTx(tx common.Uint256) (*event.ExecuteNotify, error) {
	return l.ldgStore.GetEventNotifyByTx(tx)
}

func (l *Ledger) GetEventNotifyByBlock(height uint64) ([]*event.ExecuteNotify, error) {
	return l.ldgStore.GetEventNotifyByBlock(height)
}

func (l *Ledger) GetProcessedHeight() uint64 {
	return l.ldgStore.GetProcessedHeight()
}

func (l *Ledger) SetProcessedHeight(srcBlockHeight uint64) {
	l.ldgStore.SetProcessedHeight(srcBlockHeight)
}

func (l *Ledger) Close() error {

	return l.ldgStore.Close()
}
