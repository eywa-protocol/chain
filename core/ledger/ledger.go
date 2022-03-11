package ledger

import (
	"bytes"
	"fmt"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/common/log"
	"github.com/eywa-protocol/chain/core/payload"
	"github.com/eywa-protocol/chain/core/states"
	"github.com/eywa-protocol/chain/core/store"
	"github.com/eywa-protocol/chain/core/store/ledgerstore"
	"github.com/eywa-protocol/chain/core/types"
	"github.com/eywa-protocol/chain/native/event"
	cstate "github.com/eywa-protocol/chain/native/states"
)

var DefLedger *Ledger

type Ledger struct {
	ldgStore store.LedgerStore
}

func NewLedger(dataDir string) (*Ledger, error) {
	ldgStore, err := ledgerstore.NewLedgerStore(dataDir)
	if err != nil {
		return nil, fmt.Errorf("NewLedgerStore error %s", err)
	}
	return &Ledger{
		ldgStore: ldgStore,
	}, nil
}

func (self *Ledger) GetStore() store.LedgerStore {
	return self.ldgStore
}

func (self *Ledger) Init(genesisBlock *types.Block) error {
	err := self.ldgStore.InitLedgerStoreWithGenesisBlock(genesisBlock)
	if err != nil {
		return fmt.Errorf("InitLedgerStoreWithGenesisBlock error %s", err)
	}
	return nil
}

func (self *Ledger) AddHeaders(headers []*types.Header) error {
	return self.ldgStore.AddHeaders(headers)
}

func (self *Ledger) AddBlock(block *types.Block, stateMerkleRoot common.Uint256) error {
	err := self.ldgStore.AddBlock(block, stateMerkleRoot)
	if err != nil {
		log.Errorf("Ledger AddBlock BlockHeight:%d BlockHash:%x error:%s", block.Header.Height, block.Hash(), err)
	}
	return err
}

func (self *Ledger) ExecuteBlock(b *types.Block) (store.ExecuteResult, error) {
	return self.ldgStore.ExecuteBlock(b)
}

func (self *Ledger) SubmitBlock(b *types.Block, exec store.ExecuteResult) error {
	return self.ldgStore.SubmitBlock(b, exec)
}

func (self *Ledger) GetStateMerkleRoot(height uint64) (result common.Uint256, err error) {
	return self.ldgStore.GetStateMerkleRoot(height)
}

func (self *Ledger) GetCrossStateRoot(height uint64) (common.Uint256, error) {
	return self.ldgStore.GetCrossStateRoot(height)
}

func (self *Ledger) GetBlockRootWithPreBlockHashes(startHeight uint64, txRoots []common.Uint256) common.Uint256 {
	return self.ldgStore.GetBlockRootWithPreBlockHashes(startHeight, txRoots)
}

func (self *Ledger) GetBlockByHeight(height uint64) (*types.Block, error) {
	return self.ldgStore.GetBlockByHeight(height)
}

func (self *Ledger) GetBlockByHash(blockHash common.Uint256) (*types.Block, error) {
	return self.ldgStore.GetBlockByHash(blockHash)
}

func (self *Ledger) GetHeaderByHeight(height uint64) (*types.Header, error) {
	return self.ldgStore.GetHeaderByHeight(height)
}

func (self *Ledger) GetHeaderByHash(blockHash common.Uint256) (*types.Header, error) {
	return self.ldgStore.GetHeaderByHash(blockHash)
}

func (self *Ledger) GetBlockHash(height uint64) common.Uint256 {
	return self.ldgStore.GetBlockHash(height)
}

func (self *Ledger) GetTransaction(txHash common.Uint256) (payload.Payload, error) {
	tx, _, err := self.ldgStore.GetTransaction(txHash)
	return tx, err
}

func (self *Ledger) GetTransactionWithHeight(txHash common.Uint256) (payload.Payload, uint64, error) {
	return self.ldgStore.GetTransaction(txHash)
}

func (self *Ledger) GetCurrentBlockHeight() uint64 {
	return self.ldgStore.GetCurrentBlockHeight()
}

func (self *Ledger) GetCurrentBlockHash() common.Uint256 {
	return self.ldgStore.GetCurrentBlockHash()
}

func (self *Ledger) GetCurrentHeaderHeight() uint64 {
	return self.ldgStore.GetCurrentHeaderHeight()
}

func (self *Ledger) GetCurrentHeaderHash() common.Uint256 {
	return self.ldgStore.GetCurrentHeaderHash()
}

func (self *Ledger) IsContainTransaction(txHash common.Uint256) (bool, error) {
	return self.ldgStore.IsContainTransaction(txHash)
}

func (self *Ledger) IsContainBlock(blockHash common.Uint256) (bool, error) {
	return self.ldgStore.IsContainBlock(blockHash)
}

func (self *Ledger) GetCurrentStateRoot() (common.Uint256, error) {
	return common.Uint256{}, nil
}

func (self *Ledger) GetEpochState() (*states.EpochState, error) {
	return self.ldgStore.GetEpochState()
}

func (self *Ledger) GetStorageItem(codeHash common.Address, key []byte) ([]byte, error) {
	storageKey := &states.StorageKey{
		ContractAddress: codeHash,
		Key:             key,
	}
	storageItem, err := self.ldgStore.GetStorageItem(storageKey)
	if err != nil {
		return nil, err
	}
	return storageItem.Value, nil
}

func (self *Ledger) GetMerkleProof(proofHeight, rootHeight uint64) ([]byte, error) {
	blockHash := self.ldgStore.GetBlockHash(proofHeight)
	if bytes.Equal(blockHash.ToArray(), common.UINT256_EMPTY.ToArray()) {
		return nil, fmt.Errorf("GetBlockHash(%d) empty", proofHeight)
	}
	return self.ldgStore.GetMerkleProof(blockHash.ToArray(), proofHeight+1, rootHeight)
}

func (self *Ledger) GetCrossStatesProof(height uint64, key []byte) ([]byte, error) {
	return self.ldgStore.GetCrossStatesProof(height, key)
}

func (self *Ledger) PreExecuteContract(tx payload.Payload) (*cstate.PreExecResult, error) {
	return self.ldgStore.PreExecuteContract(tx)
}

func (self *Ledger) GetEventNotifyByTx(tx common.Uint256) (*event.ExecuteNotify, error) {
	return self.ldgStore.GetEventNotifyByTx(tx)
}

func (self *Ledger) GetEventNotifyByBlock(height uint64) ([]*event.ExecuteNotify, error) {
	return self.ldgStore.GetEventNotifyByBlock(height)
}

func (self *Ledger) Close() error {
	return self.ldgStore.Close()
}
