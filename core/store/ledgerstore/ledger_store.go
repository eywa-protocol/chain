package ledgerstore

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/eywa-protocol/chain/core/payload"
	"github.com/eywa-protocol/chain/core/states"
	scom "github.com/eywa-protocol/chain/core/store/common"
	"github.com/eywa-protocol/chain/native"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/core/store"
	"github.com/eywa-protocol/chain/core/types"
	"github.com/eywa-protocol/chain/merkle"
	"github.com/eywa-protocol/chain/native/event"
	cstates "github.com/eywa-protocol/chain/native/states"
	sstate "github.com/eywa-protocol/chain/native/states"
	"github.com/eywa-protocol/chain/native/storage"
	"github.com/sirupsen/logrus"
)

const (
	SYSTEM_VERSION          = byte(1)      // Version of ledger store
	HEADER_INDEX_BATCH_SIZE = uint64(2000) // Bath size of saving header index
)

var (
	// Storage save path.
	DBDirEvent          = "ledgerevent"
	DBDirBlock          = "block"
	DBDirState          = "states"
	MerkleTreeStorePath = "merkle_tree.db"
)

// LedgerStoreImp is main store struct fo ledger
type LedgerStoreImp struct {
	blockStore           *BlockStore                      // BlockStore for saving block & transaction data
	stateStore           *StateStore                      // StateStore for saving state data, like balance, smart contract execution result, and so on.
	eventStore           *EventStore                      // EventStore for saving log those gen after smart contract executed.
	storedIndexCount     uint64                           // record the count of have saved block index
	currBlockHeight      uint64                           // Current block height
	currBlockHash        common.Uint256                   // Current block hash
	processedHeight      uint64                           // Processed source block height
	chainId              uint64                           // Ledger chain id
	headerCache          map[common.Uint256]*types.Header // BlockHash => Header
	headerIndex          map[uint64]common.Uint256        // Header index, Mapping header height => block hash
	savingBlockSemaphore chan bool
	lock                 sync.RWMutex
}

// NewLedgerStore return LedgerStoreImp instance
func NewLedgerStore(dataDir string) (*LedgerStoreImp, error) {
	ledgerStore := &LedgerStoreImp{
		headerIndex:          make(map[uint64]common.Uint256),
		headerCache:          make(map[common.Uint256]*types.Header, 0),
		savingBlockSemaphore: make(chan bool, 1),
	}

	blockStore, err := NewBlockStore(fmt.Sprintf("%s%s%s", dataDir, string(os.PathSeparator), DBDirBlock), true)
	if err != nil {
		return nil, fmt.Errorf("NewBlockStore error %s", err)
	}
	ledgerStore.blockStore = blockStore

	dbPath := fmt.Sprintf("%s%s%s", dataDir, string(os.PathSeparator), DBDirState)
	merklePath := fmt.Sprintf("%s%s%s", dataDir, string(os.PathSeparator), MerkleTreeStorePath)
	stateStore, err := NewStateStore(dbPath, merklePath)
	if err != nil {
		return nil, fmt.Errorf("NewStateStore error %s", err)
	}
	ledgerStore.stateStore = stateStore

	eventState, err := NewEventStore(fmt.Sprintf("%s%s%s", dataDir, string(os.PathSeparator), DBDirEvent))
	if err != nil {
		return nil, fmt.Errorf("NewEventStore error %s", err)
	}
	ledgerStore.eventStore = eventState

	return ledgerStore, nil
}

// InitLedgerStoreWithGenesisBlock init the ledger store with genesis block. It's the first operation after NewLedgerStore.
func (s *LedgerStoreImp) InitLedgerStoreWithGenesisBlock(genesisBlock *types.Block) error {
	hasInit, err := s.hasAlreadyInitGenesisBlock()
	if err != nil {
		return fmt.Errorf("hasAlreadyInit error %s", err)
	}
	if !hasInit {
		err = s.blockStore.ClearAll()
		if err != nil {
			return fmt.Errorf("blockStore.ClearAll error %s", err)
		}
		err = s.stateStore.ClearAll()
		if err != nil {
			return fmt.Errorf("stateStore.ClearAll error %s", err)
		}
		err = s.eventStore.ClearAll()
		if err != nil {
			return fmt.Errorf("eventStore.ClearAll error %s", err)
		}

		// bookkeeperState := &states.EpochState{
		// 	CurrEpoch: defaultEpoch,
		// 	NextEpoch: defaultEpoch,
		// }
		// err = this.stateStore.SaveEpochState(bookkeeperState)
		// if err != nil {
		// 	return fmt.Errorf("SaveEpochState error %s", err)
		// }

		result, err := s.executeBlock(genesisBlock)
		if err != nil {
			return err
		}
		err = s.submitBlock(genesisBlock, result)
		if err != nil {
			return fmt.Errorf("save genesis block error %s", err)
		}
		err = s.initGenesisBlock()
		if err != nil {
			return fmt.Errorf("init error %s", err)
		}
		s.processedHeight = genesisBlock.Header.SourceHeight
		s.chainId = genesisBlock.Header.ChainID
		s.currBlockHash = genesisBlock.Hash()
		logrus.WithFields(logrus.Fields{
			"chain_id":           s.chainId,
			"height":             s.currBlockHeight,
			"source_height":      s.processedHeight,
			"processed_height":   s.processedHeight,
			"genesis_block_hash": s.currBlockHash.ToHexString(),
		}).Infof("Ledger initialized with new genesis block.")
	} else {
		genesisHash := genesisBlock.Hash()
		exist, err := s.blockStore.ContainBlock(genesisHash)
		if err != nil {
			return fmt.Errorf("HashBlockExist error %s", err)
		}
		if !exist {
			return fmt.Errorf("GenesisBlock is not inited correctly")
		}
		err = s.init()
		if err != nil {
			return fmt.Errorf("init error %s", err)
		}
	}

	return err
}

func (s *LedgerStoreImp) hasAlreadyInitGenesisBlock() (bool, error) {
	version, err := s.blockStore.GetVersion()
	if err != nil && err != scom.ErrNotFound {
		return false, fmt.Errorf("GetVersion error %s", err)
	}
	return version == SYSTEM_VERSION, nil
}

func (s *LedgerStoreImp) initGenesisBlock() error {
	return s.blockStore.SaveVersion(SYSTEM_VERSION)
}

func (s *LedgerStoreImp) init() error {
	err := s.loadCurrentBlock()
	if err != nil {
		return fmt.Errorf("loadCurrentBlock error: %w", err)
	}
	err = s.loadHeaderIndexList()
	if err != nil {
		return fmt.Errorf("loadHeaderIndexList error: %w", err)
	}
	err = s.recoverStore()
	if err != nil {
		return fmt.Errorf("recoverStore error: %w", err)
	}
	err = s.loadProcessedHeight()
	if err != nil {
		return fmt.Errorf("loadProcessedHeight error: %w", err)
	}

	return nil
}

func (s *LedgerStoreImp) loadCurrentBlock() error {
	currentBlockHash, currentBlockHeight, err := s.blockStore.GetCurrentBlock()
	if err != nil {
		return fmt.Errorf("LoadCurrentBlock error %s", err)
	}
	s.currBlockHash = currentBlockHash
	s.currBlockHeight = currentBlockHeight
	return nil
}

func (s *LedgerStoreImp) loadHeaderIndexList() error {
	currBlockHeight := s.GetCurrentBlockHeight()
	headerIndex, err := s.blockStore.GetHeaderIndexList()
	if err != nil {
		return fmt.Errorf("LoadHeaderIndexList error %w", err)
	}
	storeIndexCount := uint64(len(headerIndex))
	s.headerIndex = headerIndex
	s.storedIndexCount = storeIndexCount

	for i := storeIndexCount; i <= currBlockHeight; i++ {
		height := i
		blockHash, err := s.blockStore.GetBlockHash(height)
		if err != nil {
			return fmt.Errorf("LoadBlockHash height %d error %w", height, err)
		}
		if blockHash == common.UINT256_EMPTY {
			return fmt.Errorf("LoadBlockHash height %d hash nil", height)
		}
		s.headerIndex[height] = blockHash
	}
	return nil
}

func (s *LedgerStoreImp) loadProcessedHeight() error {
	if processedHeight, err := s.stateStore.getProcessedHeight(); err != nil {
		return fmt.Errorf("stateStore.getProcessedHeight error %w", err)
	} else if blockHash, _, err := s.blockStore.GetCurrentBlock(); err != nil {
		return fmt.Errorf("blockStore.GetCurrentBlock error %w", err)
	} else if head, err := s.blockStore.GetHeader(blockHash); err != nil {
		return fmt.Errorf("blockStore.GetHeader error %w", err)
	} else {
		s.chainId = head.ChainID
		if processedHeight < head.SourceHeight {
			s.processedHeight = head.SourceHeight
		} else if processedHeight > 0 {
			s.processedHeight = processedHeight
		} else {
			return fmt.Errorf("processed height is zerro")
		}
		logrus.WithFields(logrus.Fields{
			"chain_id":         s.chainId,
			"height":           head.Height,
			"source_height":    head.SourceHeight,
			"processed_height": s.processedHeight,
		}).Infof("Ledger initialized.")
		return nil
	}

}

func (s *LedgerStoreImp) recoverStore() error {
	blockHeight := s.GetCurrentBlockHeight()

	_, stateHeight, err := s.stateStore.GetCurrentBlock()
	if err != nil {
		return fmt.Errorf("stateStore.GetCurrentBlock error %w", err)
	}
	for i := stateHeight; i < blockHeight; i++ {
		blockHash, err := s.blockStore.GetBlockHash(i)
		if err != nil {
			return fmt.Errorf("blockStore.GetBlockHash height:%d error:%w", i, err)
		}
		block, err := s.blockStore.GetBlock(blockHash)
		if err != nil {
			return fmt.Errorf("blockStore.GetBlock height:%d error:%w", i, err)
		}
		s.eventStore.NewBatch()
		s.stateStore.NewBatch()
		result, err := s.executeBlock(block)
		if err != nil {
			return err
		}
		err = s.saveBlockToStateStore(block, result)
		if err != nil {
			return fmt.Errorf("save to state store height:%d error:%s", i, err)
		}
		err = s.saveBlockToEventStore(block)
		if err != nil {
			return fmt.Errorf("save to event store height:%d error:%s", i, err)
		}
		err = s.eventStore.CommitTo()
		if err != nil {
			return fmt.Errorf("eventStore.CommitTo height:%d error %s", i, err)
		}
		err = s.stateStore.CommitTo()
		if err != nil {
			return fmt.Errorf("stateStore.CommitTo height:%d error %s", i, err)
		}
	}

	return nil
}

func (s *LedgerStoreImp) setHeaderIndex(height uint64, blockHash common.Uint256) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.headerIndex[height] = blockHash
}

func (s *LedgerStoreImp) getHeaderIndex(height uint64) common.Uint256 {
	s.lock.RLock()
	defer s.lock.RUnlock()
	blockHash, ok := s.headerIndex[height]
	if !ok {
		return common.Uint256{}
	}
	return blockHash
}

// GetCurrentHeaderHeight return the current header height.
// In block sync states, Header height is usually higher than block height that is has already committed to storage
func (s *LedgerStoreImp) GetCurrentHeaderHeight() uint64 {
	s.lock.RLock()
	defer s.lock.RUnlock()
	size := len(s.headerIndex)
	if size == 0 {
		return 0
	}
	return uint64(size) - 1
}

// GetCurrentHeaderHash return the current header hash. The current header means the latest header.
func (s *LedgerStoreImp) GetCurrentHeaderHash() common.Uint256 {
	s.lock.RLock()
	defer s.lock.RUnlock()
	size := len(s.headerIndex)
	if size == 0 {
		return common.Uint256{}
	}
	return s.headerIndex[uint64(size)-1]
}

func (s *LedgerStoreImp) setCurrentBlock(height uint64, blockHash common.Uint256) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.currBlockHash = blockHash
	s.currBlockHeight = height
	return
}

// GetCurrentBlock return the current block height, and block hash.
// Current block means the latest block in store.
func (s *LedgerStoreImp) GetCurrentBlock() (uint64, common.Uint256) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.currBlockHeight, s.currBlockHash
}

// GetCurrentBlockHash return the current block hash
func (s *LedgerStoreImp) GetCurrentBlockHash() common.Uint256 {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.currBlockHash
}

// GetCurrentBlockHeight return the current block height
func (s *LedgerStoreImp) GetCurrentBlockHeight() uint64 {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.currBlockHeight
}

func (s *LedgerStoreImp) addHeaderCache(header *types.Header) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.headerCache[header.Hash()] = header
}

func (s *LedgerStoreImp) delHeaderCache(blockHash common.Uint256) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.headerCache, blockHash)
}

func (s *LedgerStoreImp) getHeaderCache(blockHash common.Uint256) *types.Header {
	s.lock.RLock()
	defer s.lock.RUnlock()
	header, ok := s.headerCache[blockHash]
	if !ok {
		return nil
	}
	return header
}

func (s *LedgerStoreImp) verifyHeader(header *types.Header) error {
	if header.Height == 0 {
		return nil
	}
	var prevHeader *types.Header
	prevHeaderHash := header.PrevBlockHash
	prevHeader, err := s.GetHeaderByHash(prevHeaderHash)
	if err != nil && err != scom.ErrNotFound {
		return fmt.Errorf("get prev header error %s", err)
	}
	if prevHeader == nil {
		return fmt.Errorf("cannot find pre header by blockHash %s", prevHeaderHash.ToHexString())
	}

	if prevHeader.Height+1 != header.Height {
		return fmt.Errorf("block height is incorrect: prevheight %d curHeight %d", prevHeader.Height+1, header.Height)
	}
	if prevHeader.SourceHeight >= header.SourceHeight {
		return fmt.Errorf("block source height [%d] missmatch to prev block source [%d]",
			header.SourceHeight, prevHeader.SourceHeight)
	}
	return nil
}

// AddHeader add header to cache, and add the mapping of block height to block hash. Using in block sync
func (s *LedgerStoreImp) AddHeader(header *types.Header) error {
	nextHeaderHeight := s.GetCurrentHeaderHeight() + 1
	if header.Height != nextHeaderHeight {
		return fmt.Errorf("header height %d not equal next header height %d", header.Height, nextHeaderHeight)
	}
	err := s.verifyHeader(header)
	if err != nil {
		return fmt.Errorf("AddHeader verifyHeader error %s", err)
	}
	s.addHeaderCache(header)
	s.setHeaderIndex(header.Height, header.Hash())
	return nil
}

// AddHeaders bath add header.
func (s *LedgerStoreImp) AddHeaders(headers []*types.Header) error {
	sort.Slice(headers, func(i, j int) bool {
		return headers[i].Height < headers[j].Height
	})
	var err error
	for _, header := range headers {
		err = s.AddHeader(header)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *LedgerStoreImp) GetStateMerkleRoot(height uint64) (common.Uint256, error) {
	return s.stateStore.GetStateMerkleRoot(height)
}

func (s *LedgerStoreImp) GetCrossStateRoot(height uint64) (common.Uint256, error) {
	return s.stateStore.GetCrossStateRoot(height)
}

func (s *LedgerStoreImp) ExecuteBlock(block *types.Block) (result store.ExecuteResult, err error) {
	s.getSavingBlockLock()
	defer s.releaseSavingBlockLock()
	currBlockHeight := s.GetCurrentBlockHeight()
	blockHeight := block.Header.Height
	if blockHeight <= currBlockHeight {
		result.MerkleRoot, err = s.GetStateMerkleRoot(blockHeight)
		return
	}
	nextBlockHeight := currBlockHeight + 1
	if blockHeight != nextBlockHeight {
		err = fmt.Errorf("block height %d not equal next block height %d", blockHeight, nextBlockHeight)
		return
	}
	result, err = s.executeBlock(block)
	return
}

func (s *LedgerStoreImp) SubmitBlock(block *types.Block, result store.ExecuteResult) error {
	s.getSavingBlockLock()
	defer s.releaseSavingBlockLock()
	currBlockHeight := s.GetCurrentBlockHeight()
	blockHeight := block.Header.Height
	if blockHeight <= currBlockHeight {
		return nil
	}
	nextBlockHeight := currBlockHeight + 1
	if blockHeight != nextBlockHeight {
		return fmt.Errorf("block height %d not equal next block height %d", blockHeight, nextBlockHeight)
	}
	err := s.verifyHeader(block.Header)
	if err != nil {
		return fmt.Errorf("SubmitBlock verifyHeader error %s", err)
	}

	err = block.VerifyIntegrity()
	if err != nil {
		return fmt.Errorf("SubmitBlock block integrity error %s", err)
	}

	err = s.submitBlock(block, result)
	if err != nil {
		return fmt.Errorf("saveBlock error %s", err)
	}
	s.delHeaderCache(block.Hash())
	return nil
}

// AddBlock add the block to store.
// When the block is not the next block, it will be cache. until the missing block arrived
func (s *LedgerStoreImp) AddBlock(block *types.Block, stateMerkleRoot common.Uint256) error {
	currBlockHeight := s.GetCurrentBlockHeight()
	blockHeight := block.Header.Height
	if blockHeight <= currBlockHeight {
		return nil
	}
	nextBlockHeight := currBlockHeight + 1
	if blockHeight != nextBlockHeight {
		return fmt.Errorf("block height %d not equal next block height %d", blockHeight, nextBlockHeight)
	}
	err := s.verifyHeader(block.Header)
	if err != nil {
		return fmt.Errorf("AddBlock stateMerkleRoot verifyHeader error %s", err)
	}

	err = s.saveBlock(block, stateMerkleRoot)
	if err != nil {
		return fmt.Errorf("saveBlock error %s", err)
	}
	s.delHeaderCache(block.Hash())
	return nil
}

func (s *LedgerStoreImp) GetCrossStatesProof(height uint64, key []byte) ([]byte, error) {
	hashes, err := s.stateStore.GetCrossStates(height)
	if err != nil {
		return nil, fmt.Errorf("GetCrossStates:%s", err)
	}
	state, err := s.stateStore.GetStorageValue(key)
	if err != nil {
		return nil, fmt.Errorf("GetStorageState key:%x", key)
	}
	path, err := merkle.MerkleLeafPath(state, hashes)
	if err != nil {
		return nil, err
	}
	return path, nil
}

func (s *LedgerStoreImp) saveBlockToBlockStore(block *types.Block) error {
	blockHash := block.Hash()
	blockHeight := block.Header.Height

	s.setHeaderIndex(blockHeight, blockHash)
	err := s.saveHeaderIndexList()
	if err != nil {
		return fmt.Errorf("saveHeaderIndexList error %s", err)
	}
	err = s.blockStore.SaveCurrentBlock(blockHeight, blockHash)
	if err != nil {
		return fmt.Errorf("SaveCurrentBlock error %s", err)
	}
	s.blockStore.SaveBlockHash(blockHeight, blockHash)
	err = s.blockStore.SaveBlock(block)
	if err != nil {
		return fmt.Errorf("SaveBlock height %d hash %s error %s", blockHeight, blockHash.ToHexString(), err)
	}
	return nil
}

// TODO: remove block param if not needed

func (s *LedgerStoreImp) executeBlock(block *types.Block) (result store.ExecuteResult, err error) {
	overlay := s.stateStore.NewOverlayDB()
	result.Hash = overlay.ChangeHash()
	result.WriteSet = overlay.GetWriteSet()
	result.MerkleRoot = s.stateStore.GetStateMerkleRootWithNewHash(result.Hash)
	return
}

func (s *LedgerStoreImp) saveBlockToStateStore(block *types.Block, result store.ExecuteResult) error {
	blockHash := block.Hash()
	blockHeight := block.Header.Height

	err := s.stateStore.AddStateMerkleTreeRoot(blockHeight, result.Hash)
	if err != nil {
		return fmt.Errorf("AddBlockMerkleTreeRoot error %s", err)
	}

	err = s.stateStore.AddBlockMerkleTreeRoot(block.Header.PrevBlockHash)
	if err != nil {
		return fmt.Errorf("AddBlockMerkleTreeRoot error %s", err)
	}

	err = s.stateStore.SaveCurrentBlock(blockHeight, blockHash)
	if err != nil {
		return fmt.Errorf("SaveCurrentBlock error %s", err)
	}

	err = s.stateStore.AddCrossStates(blockHeight, result.CrossHashes, result.CrossStatesRoot)
	if err != nil {
		return err
	}

	logrus.Debugf("the state transition hash of block %d is:%s", blockHeight, result.Hash.ToHexString())

	result.WriteSet.ForEach(func(key, val []byte) {
		if len(val) == 0 {
			s.stateStore.BatchDeleteRawKey(key)
		} else {
			s.stateStore.BatchPutRawKeyVal(key, val)
		}
	})

	return nil
}

func (s *LedgerStoreImp) saveBlockToEventStore(block *types.Block) error {
	blockHash := block.Hash()
	blockHeight := block.Header.Height
	txs := make([]common.Uint256, 0)
	for _, tx := range block.Transactions {
		txHash := tx.Hash()
		txs = append(txs, txHash)
	}
	if len(txs) > 0 {
		err := s.eventStore.SaveEventNotifyByBlock(block.Header.Height, txs)
		if err != nil {
			return fmt.Errorf("SaveEventNotifyByBlock error %s", err)
		}
	}
	err := s.eventStore.SaveCurrentBlock(blockHeight, blockHash)
	if err != nil {
		return fmt.Errorf("SaveCurrentBlock error %s", err)
	}
	return nil
}

func (s *LedgerStoreImp) tryGetSavingBlockLock() (hasLocked bool) {
	select {
	case s.savingBlockSemaphore <- true:
		return false
	default:
		return true
	}
}

func (s *LedgerStoreImp) getSavingBlockLock() {
	s.savingBlockSemaphore <- true
}

func (s *LedgerStoreImp) releaseSavingBlockLock() {
	select {
	case <-s.savingBlockSemaphore:
		return
	default:
		panic("can not release in unlocked state")
	}
}

// saveBlock do the job of execution samrt contract and commit block to store.
func (s *LedgerStoreImp) submitBlock(block *types.Block, result store.ExecuteResult) error {
	blockHash := block.Hash()
	blockHeight := block.Header.Height

	// blockRoot := this.GetBlockRootWithPreBlockHashes(block.Header.Height, []common.Uint256{block.Header.PrevBlockHash})
	// if block.Header.Height != 0 && blockRoot != block.Header.BlockRoot {
	// 	return fmt.Errorf("wrong block root at height:%d, expected:%s, got:%s",
	// 		block.Header.Height, blockRoot.ToHexString(), block.Header.BlockRoot.ToHexString())
	// }

	s.blockStore.NewBatch()
	s.stateStore.NewBatch()
	s.eventStore.NewBatch()
	err := s.saveBlockToBlockStore(block)
	if err != nil {
		return fmt.Errorf("save to block store height:%d error:%s", blockHeight, err)
	}
	err = s.saveBlockToStateStore(block, result)
	if err != nil {
		return fmt.Errorf("save to state store height:%d error:%s", blockHeight, err)
	}
	err = s.saveBlockToEventStore(block)
	if err != nil {
		return fmt.Errorf("save to event store height:%d error:%s", blockHeight, err)
	}
	err = s.blockStore.CommitTo()
	if err != nil {
		return fmt.Errorf("blockStore.CommitTo height:%d error %s", blockHeight, err)
	}
	// event store is idempotent to re-save when in recovering process, so save first before stateStore
	err = s.eventStore.CommitTo()
	if err != nil {
		return fmt.Errorf("eventStore.CommitTo height:%d error %s", blockHeight, err)
	}
	err = s.stateStore.CommitTo()
	if err != nil {
		return fmt.Errorf("stateStore.CommitTo height:%d error %s", blockHeight, err)
	}
	s.setCurrentBlock(blockHeight, blockHash)

	// if events.DefActorPublisher != nil {
	//	events.DefActorPublisher.Publish(
	//		message.TOPIC_SAVE_BLOCK_COMPLETE,
	//		&message.SaveBlockCompleteMsg{
	//			Block: block,
	//		})
	// }
	return nil
}

// saveBlock do the job of execution samrt contract and commit block to store.
func (s *LedgerStoreImp) saveBlock(block *types.Block, stateMerkleRoot common.Uint256) error {
	blockHeight := block.Header.Height
	if s.tryGetSavingBlockLock() {
		// hash already saved or is saving
		return nil
	}
	defer s.releaseSavingBlockLock()
	if blockHeight > 0 && blockHeight != (s.GetCurrentBlockHeight()+1) {
		return nil
	}

	result, err := s.executeBlock(block)
	if err != nil {
		return err
	}

	if result.MerkleRoot != stateMerkleRoot {
		return errors.New("state merkle root mismatch")
	}

	return s.submitBlock(block, result)
}

func (s *LedgerStoreImp) saveHeaderIndexList() error {
	s.lock.RLock()
	storeCount := s.storedIndexCount
	currHeight := s.currBlockHeight
	if currHeight-storeCount < HEADER_INDEX_BATCH_SIZE {
		s.lock.RUnlock()
		return nil
	}

	headerList := make([]common.Uint256, HEADER_INDEX_BATCH_SIZE)
	for i := uint64(0); i < HEADER_INDEX_BATCH_SIZE; i++ {
		height := storeCount + i
		headerList[i] = s.headerIndex[height]
	}
	s.lock.RUnlock()

	err := s.blockStore.SaveHeaderIndexList(storeCount, headerList)
	if err != nil {
		return fmt.Errorf("SaveHeaderIndexList start %d error %s", storeCount, err)
	}

	s.lock.Lock()
	s.storedIndexCount += HEADER_INDEX_BATCH_SIZE
	s.lock.Unlock()
	return nil
}

func (s *LedgerStoreImp) PreExecuteContract(tx payload.Payload) (*cstates.PreExecResult, error) {
	result := &sstate.PreExecResult{State: event.CONTRACT_STATE_FAIL, Result: nil}
	if _, ok := tx.(*payload.InvokeCode); !ok {
		return result, fmt.Errorf("transaction payload type error")
	}
	hash := s.GetCurrentBlockHash()
	block, err := s.GetBlockByHash(hash)
	if err != nil {
		return result, fmt.Errorf("get current block error")
	}
	overlay := s.stateStore.NewOverlayDB()
	cache := storage.NewCacheDB(overlay)

	service, err := native.NewNativeService(cache, tx, block.Header.Height,
		hash, block.Header.ChainID, tx.(*payload.InvokeCode).Code, true)
	if err != nil {
		return result, fmt.Errorf("PreExecuteContract Error: %+v\n", err)
	}
	res, err := service.Invoke()
	if err != nil {
		return result, err
	}
	return &sstate.PreExecResult{State: event.CONTRACT_STATE_SUCCESS, Result: common.ToHexString(res.([]byte)), Notify: service.GetNotify()}, nil
}

// IsContainBlock return whether the block is in store
func (s *LedgerStoreImp) IsContainBlock(blockHash common.Uint256) (bool, error) {
	return s.blockStore.ContainBlock(blockHash)
}

// IsContainTransaction return whether the transaction is in store. Wrap function of BlockStore.ContainTransaction
func (s *LedgerStoreImp) IsContainTransaction(txHash common.Uint256) (bool, error) {
	return s.blockStore.ContainTransaction(txHash)
}

func (s *LedgerStoreImp) GetBlockRootWithPreBlockHashes(startHeight uint64, preBlockHashes []common.Uint256) common.Uint256 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	// the block height in consensus is far behind ledger, this case should be rare
	if s.currBlockHeight > startHeight+uint64(len(preBlockHashes))-1 {
		// or return error?
		logrus.Errorf("s.currBlockHeight= %d, startHeight= %d, len(preBlockHashes)= %d\n", s.currBlockHeight, startHeight, len(preBlockHashes))
		return common.UINT256_EMPTY
	}

	needs := preBlockHashes[s.currBlockHeight+1-startHeight:]
	return s.stateStore.GetBlockRootWithPreBlockHashes(needs)
}

// GetBlockHash return the block hash by block height
func (s *LedgerStoreImp) GetBlockHash(height uint64) common.Uint256 {
	return s.getHeaderIndex(height)
}

// GetHeaderByHash return the block header by block hash
func (s *LedgerStoreImp) GetHeaderByHash(blockHash common.Uint256) (*types.Header, error) {
	header := s.getHeaderCache(blockHash)
	if header != nil {
		return header, nil
	}
	return s.blockStore.GetHeader(blockHash)
}

// GetHeaderByHeight return the block header by block height
func (s *LedgerStoreImp) GetHeaderByHeight(height uint64) (*types.Header, error) {
	blockHash := s.GetBlockHash(height)
	var empty common.Uint256
	if blockHash == empty {
		return nil, nil
	}
	return s.GetHeaderByHash(blockHash)
}

// GetTransaction return transaction by transaction hash. Wrap function of BlockStore.GetTransaction
func (s *LedgerStoreImp) GetTransaction(txHash common.Uint256) (payload.Payload, uint64, error) {
	return s.blockStore.GetTransaction(txHash)
}

// GetTransactionByReqId return transaction by request id. Wrap function of BlockStore.GetTransactionByReqId
func (s *LedgerStoreImp) GetTransactionByReqId(reqId [32]byte) (payload.Payload, uint64, error) {
	return s.blockStore.GetTransactionByReqId(reqId)
}

// GetRequestState return request state by request id. Wrap function of BlockStore.GetRequestState
func (s *LedgerStoreImp) GetRequestState(reqId [32]byte) (payload.ReqState, error) {
	return s.blockStore.GetRequestState(reqId)
}

// GetBlockByHash return block by block hash. Wrap function of BlockStore.GetBlockByHash
func (s *LedgerStoreImp) GetBlockByHash(blockHash common.Uint256) (*types.Block, error) {
	return s.blockStore.GetBlock(blockHash)
}

// GetBlockByHeight return block by height.
func (s *LedgerStoreImp) GetBlockByHeight(height uint64) (*types.Block, error) {
	blockHash := s.GetBlockHash(height)
	var empty common.Uint256
	if blockHash == empty {
		return nil, nil
	}
	return s.GetBlockByHash(blockHash)
}

// GetEpochState return the bookkeeper state. Wrap function of StateStore.GetEpochState
func (s *LedgerStoreImp) GetEpochState() (*states.EpochState, error) {
	return s.stateStore.GetEpochState()
}

// GetMerkleProof return the block merkle proof. Wrap function of StateStore.GetMerkleProof
func (s *LedgerStoreImp) GetMerkleProof(raw []byte, proofHeight, rootHeight uint64) ([]byte, error) {
	return s.stateStore.GetMerkleProof(raw, proofHeight, rootHeight)
}

// GetStorageItem return the storage value of the key in smart contract. Wrap function of StateStore.GetStorageState
func (s *LedgerStoreImp) GetStorageItem(key *states.StorageKey) (*states.StorageItem, error) {
	return s.stateStore.GetStorageState(key)
}

// GetEventNotifyByTx return the events notify gen by executing of smart contract.  Wrap function of EventStore.GetEventNotifyByTx
func (s *LedgerStoreImp) GetEventNotifyByTx(tx common.Uint256) (*event.ExecuteNotify, error) {
	return s.eventStore.GetEventNotifyByTx(tx)
}

// GetEventNotifyByBlock return the transaction hash which have event notice after execution of smart contract. Wrap function of EventStore.GetEventNotifyByBlock
func (s *LedgerStoreImp) GetEventNotifyByBlock(height uint64) ([]*event.ExecuteNotify, error) {
	return s.eventStore.GetEventNotifyByBlock(height)
}

// GetProcessedHeight return source chain processed height
// Note: processed height is not last stored block types.Header.SourceHeight
func (s *LedgerStoreImp) GetProcessedHeight() uint64 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.processedHeight
}

// SetProcessedHeight set source chain processed height to ledger
func (s *LedgerStoreImp) SetProcessedHeight(srcBlockHeight uint64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if srcBlockHeight > s.processedHeight {
		s.processedHeight = srcBlockHeight
	}
}

// saveProcessedHeight save processed height to ledger state store
func (s *LedgerStoreImp) saveProcessedHeight() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.stateStore.NewBatch()
	if err := s.stateStore.SaveProcessedHeight(s.processedHeight); err != nil {
		return fmt.Errorf("save processed height [%d] error: %w", s.processedHeight, err)
	}

	return s.stateStore.CommitTo()
}

// Close ledger store.
func (s *LedgerStoreImp) Close() error {
	if err := s.saveProcessedHeight(); err != nil {
		logrus.Error(err)
	}

	s.lock.RLock()
	logrus.Infof("gracefull shutdown ledger.  processed height: %d", s.processedHeight)
	s.lock.RUnlock()
	err := s.blockStore.Close()
	if err != nil {
		return fmt.Errorf("blockStore close error %s", err)
	}
	err = s.stateStore.Close()
	if err != nil {
		return fmt.Errorf("stateStore close error %s", err)
	}
	err = s.eventStore.Close()
	if err != nil {
		return fmt.Errorf("eventStore close error %s", err)
	}
	return nil
}
