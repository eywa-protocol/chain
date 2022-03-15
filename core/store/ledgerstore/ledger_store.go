package ledgerstore

import (
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/eywa-protocol/chain/core/payload"
	"github.com/eywa-protocol/chain/core/states"
	scom "github.com/eywa-protocol/chain/core/store/common"
	"github.com/eywa-protocol/chain/native"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/common/log"
	"github.com/eywa-protocol/chain/core/store"
	"github.com/eywa-protocol/chain/core/store/overlaydb"
	"github.com/eywa-protocol/chain/core/types"
	"github.com/eywa-protocol/chain/merkle"
	"github.com/eywa-protocol/chain/native/event"
	cstates "github.com/eywa-protocol/chain/native/states"
	sstate "github.com/eywa-protocol/chain/native/states"
	"github.com/eywa-protocol/chain/native/storage"
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
func (this *LedgerStoreImp) InitLedgerStoreWithGenesisBlock(genesisBlock *types.Block) error {
	hasInit, err := this.hasAlreadyInitGenesisBlock()
	if err != nil {
		return fmt.Errorf("hasAlreadyInit error %s", err)
	}
	if !hasInit {
		err = this.blockStore.ClearAll()
		if err != nil {
			return fmt.Errorf("blockStore.ClearAll error %s", err)
		}
		err = this.stateStore.ClearAll()
		if err != nil {
			return fmt.Errorf("stateStore.ClearAll error %s", err)
		}
		err = this.eventStore.ClearAll()
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

		result, err := this.executeBlock(genesisBlock)
		if err != nil {
			return err
		}
		err = this.submitBlock(genesisBlock, result)
		if err != nil {
			return fmt.Errorf("save genesis block error %s", err)
		}
		err = this.initGenesisBlock()
		if err != nil {
			return fmt.Errorf("init error %s", err)
		}
		genHash := genesisBlock.Hash()
		log.Infof("GenesisBlock init success. GenesisBlock hash:%s\n", genHash.ToHexString())
		this.currBlockHash = genesisBlock.Hash()
	} else {
		genesisHash := genesisBlock.Hash()
		exist, err := this.blockStore.ContainBlock(genesisHash)
		if err != nil {
			return fmt.Errorf("HashBlockExist error %s", err)
		}
		if !exist {
			return fmt.Errorf("GenesisBlock is not inited correctly")
		}
		err = this.init()
		if err != nil {
			return fmt.Errorf("init error %s", err)
		}
	}

	return err
}

func (this *LedgerStoreImp) hasAlreadyInitGenesisBlock() (bool, error) {
	version, err := this.blockStore.GetVersion()
	if err != nil && err != scom.ErrNotFound {
		return false, fmt.Errorf("GetVersion error %s", err)
	}
	return version == SYSTEM_VERSION, nil
}

func (this *LedgerStoreImp) initGenesisBlock() error {
	return this.blockStore.SaveVersion(SYSTEM_VERSION)
}

func (this *LedgerStoreImp) init() error {
	err := this.loadCurrentBlock()
	if err != nil {
		return fmt.Errorf("loadCurrentBlock error %s", err)
	}
	err = this.loadHeaderIndexList()
	if err != nil {
		return fmt.Errorf("loadHeaderIndexList error %s", err)
	}
	err = this.recoverStore()
	if err != nil {
		return fmt.Errorf("recoverStore error %s", err)
	}
	return nil
}

func (this *LedgerStoreImp) loadCurrentBlock() error {
	currentBlockHash, currentBlockHeight, err := this.blockStore.GetCurrentBlock()
	if err != nil {
		return fmt.Errorf("LoadCurrentBlock error %s", err)
	}
	log.Infof("InitCurrentBlock currentBlockHash %s currentBlockHeight %d", currentBlockHash.ToHexString(), currentBlockHeight)
	this.currBlockHash = currentBlockHash
	this.currBlockHeight = currentBlockHeight
	return nil
}

func (this *LedgerStoreImp) loadHeaderIndexList() error {
	currBlockHeight := this.GetCurrentBlockHeight()
	headerIndex, err := this.blockStore.GetHeaderIndexList()
	if err != nil {
		return fmt.Errorf("LoadHeaderIndexList error %s", err)
	}
	storeIndexCount := uint64(len(headerIndex))
	this.headerIndex = headerIndex
	this.storedIndexCount = storeIndexCount

	for i := storeIndexCount; i <= currBlockHeight; i++ {
		height := i
		blockHash, err := this.blockStore.GetBlockHash(height)
		if err != nil {
			return fmt.Errorf("LoadBlockHash height %d error %s", height, err)
		}
		if blockHash == common.UINT256_EMPTY {
			return fmt.Errorf("LoadBlockHash height %d hash nil", height)
		}
		this.headerIndex[height] = blockHash
	}
	return nil
}

func (this *LedgerStoreImp) recoverStore() error {
	blockHeight := this.GetCurrentBlockHeight()

	_, stateHeight, err := this.stateStore.GetCurrentBlock()
	if err != nil {
		return fmt.Errorf("stateStore.GetCurrentBlock error %s", err)
	}
	for i := stateHeight; i < blockHeight; i++ {
		blockHash, err := this.blockStore.GetBlockHash(i)
		if err != nil {
			return fmt.Errorf("blockStore.GetBlockHash height:%d error:%s", i, err)
		}
		block, err := this.blockStore.GetBlock(blockHash)
		if err != nil {
			return fmt.Errorf("blockStore.GetBlock height:%d error:%s", i, err)
		}
		this.eventStore.NewBatch()
		this.stateStore.NewBatch()
		result, err := this.executeBlock(block)
		if err != nil {
			return err
		}
		err = this.saveBlockToStateStore(block, result)
		if err != nil {
			return fmt.Errorf("save to state store height:%d error:%s", i, err)
		}
		err = this.saveBlockToEventStore(block)
		if err != nil {
			return fmt.Errorf("save to event store height:%d error:%s", i, err)
		}
		err = this.eventStore.CommitTo()
		if err != nil {
			return fmt.Errorf("eventStore.CommitTo height:%d error %s", i, err)
		}
		err = this.stateStore.CommitTo()
		if err != nil {
			return fmt.Errorf("stateStore.CommitTo height:%d error %s", i, err)
		}
	}
	return nil
}

func (this *LedgerStoreImp) setHeaderIndex(height uint64, blockHash common.Uint256) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.headerIndex[height] = blockHash
}

func (this *LedgerStoreImp) getHeaderIndex(height uint64) common.Uint256 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	blockHash, ok := this.headerIndex[height]
	if !ok {
		return common.Uint256{}
	}
	return blockHash
}

// GetCurrentHeaderHeight return the current header height.
// In block sync states, Header height is usually higher than block height that is has already committed to storage
func (this *LedgerStoreImp) GetCurrentHeaderHeight() uint64 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	size := len(this.headerIndex)
	if size == 0 {
		return 0
	}
	return uint64(size) - 1
}

// GetCurrentHeaderHash return the current header hash. The current header means the latest header.
func (this *LedgerStoreImp) GetCurrentHeaderHash() common.Uint256 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	size := len(this.headerIndex)
	if size == 0 {
		return common.Uint256{}
	}
	return this.headerIndex[uint64(size)-1]
}

func (this *LedgerStoreImp) setCurrentBlock(height uint64, blockHash common.Uint256) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.currBlockHash = blockHash
	this.currBlockHeight = height
	return
}

// GetCurrentBlock return the current block height, and block hash.
// Current block means the latest block in store.
func (this *LedgerStoreImp) GetCurrentBlock() (uint64, common.Uint256) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.currBlockHeight, this.currBlockHash
}

// GetCurrentBlockHash return the current block hash
func (this *LedgerStoreImp) GetCurrentBlockHash() common.Uint256 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.currBlockHash
}

// GetCurrentBlockHeight return the current block height
func (this *LedgerStoreImp) GetCurrentBlockHeight() uint64 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.currBlockHeight
}

func (this *LedgerStoreImp) addHeaderCache(header *types.Header) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.headerCache[header.Hash()] = header
}

func (this *LedgerStoreImp) delHeaderCache(blockHash common.Uint256) {
	this.lock.Lock()
	defer this.lock.Unlock()
	delete(this.headerCache, blockHash)
}

func (this *LedgerStoreImp) getHeaderCache(blockHash common.Uint256) *types.Header {
	this.lock.RLock()
	defer this.lock.RUnlock()
	header, ok := this.headerCache[blockHash]
	if !ok {
		return nil
	}
	return header
}

func (this *LedgerStoreImp) verifyHeader(header *types.Header) error {
	if header.Height == 0 {
		return nil
	}
	var prevHeader *types.Header
	prevHeaderHash := header.PrevBlockHash
	prevHeader, err := this.GetHeaderByHash(prevHeaderHash)
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
func (this *LedgerStoreImp) AddHeader(header *types.Header) error {
	nextHeaderHeight := this.GetCurrentHeaderHeight() + 1
	if header.Height != nextHeaderHeight {
		return fmt.Errorf("header height %d not equal next header height %d", header.Height, nextHeaderHeight)
	}
	err := this.verifyHeader(header)
	if err != nil {
		return fmt.Errorf("AddHeader verifyHeader error %s", err)
	}
	this.addHeaderCache(header)
	this.setHeaderIndex(header.Height, header.Hash())
	return nil
}

// AddHeaders bath add header.
func (this *LedgerStoreImp) AddHeaders(headers []*types.Header) error {
	sort.Slice(headers, func(i, j int) bool {
		return headers[i].Height < headers[j].Height
	})
	var err error
	for _, header := range headers {
		err = this.AddHeader(header)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *LedgerStoreImp) GetStateMerkleRoot(height uint64) (common.Uint256, error) {
	return this.stateStore.GetStateMerkleRoot(height)
}

func (this *LedgerStoreImp) GetCrossStateRoot(height uint64) (common.Uint256, error) {
	return this.stateStore.GetCrossStateRoot(height)
}

func (this *LedgerStoreImp) ExecuteBlock(block *types.Block) (result store.ExecuteResult, err error) {
	this.getSavingBlockLock()
	defer this.releaseSavingBlockLock()
	currBlockHeight := this.GetCurrentBlockHeight()
	blockHeight := block.Header.Height
	if blockHeight <= currBlockHeight {
		result.MerkleRoot, err = this.GetStateMerkleRoot(blockHeight)
		return
	}
	nextBlockHeight := currBlockHeight + 1
	if blockHeight != nextBlockHeight {
		err = fmt.Errorf("block height %d not equal next block height %d", blockHeight, nextBlockHeight)
		return
	}
	result, err = this.executeBlock(block)
	return
}

func (this *LedgerStoreImp) SubmitBlock(block *types.Block, result store.ExecuteResult) error {
	this.getSavingBlockLock()
	defer this.releaseSavingBlockLock()
	currBlockHeight := this.GetCurrentBlockHeight()
	blockHeight := block.Header.Height
	if blockHeight <= currBlockHeight {
		return nil
	}
	nextBlockHeight := currBlockHeight + 1
	if blockHeight != nextBlockHeight {
		return fmt.Errorf("block height %d not equal next block height %d", blockHeight, nextBlockHeight)
	}
	err := this.verifyHeader(block.Header)
	if err != nil {
		return fmt.Errorf("SubmitBlock verifyHeader error %s", err)
	}

	err = block.VerifyIntegrity()
	if err != nil {
		return fmt.Errorf("SubmitBlock block integrity error %s", err)
	}

	err = this.submitBlock(block, result)
	if err != nil {
		return fmt.Errorf("saveBlock error %s", err)
	}
	this.delHeaderCache(block.Hash())
	return nil
}

// AddBlock add the block to store.
// When the block is not the next block, it will be cache. until the missing block arrived
func (this *LedgerStoreImp) AddBlock(block *types.Block, stateMerkleRoot common.Uint256) error {
	currBlockHeight := this.GetCurrentBlockHeight()
	blockHeight := block.Header.Height
	if blockHeight <= currBlockHeight {
		return nil
	}
	nextBlockHeight := currBlockHeight + 1
	if blockHeight != nextBlockHeight {
		return fmt.Errorf("block height %d not equal next block height %d", blockHeight, nextBlockHeight)
	}
	err := this.verifyHeader(block.Header)
	if err != nil {
		return fmt.Errorf("AddBlock stateMerkleRoot verifyHeader error %s", err)
	}

	err = this.saveBlock(block, stateMerkleRoot)
	if err != nil {
		return fmt.Errorf("saveBlock error %s", err)
	}
	this.delHeaderCache(block.Hash())
	return nil
}

func (this *LedgerStoreImp) GetCrossStatesProof(height uint64, key []byte) ([]byte, error) {
	hashes, err := this.stateStore.GetCrossStates(height)
	if err != nil {
		return nil, fmt.Errorf("GetCrossStates:%s", err)
	}
	state, err := this.stateStore.GetStorageValue(key)
	if err != nil {
		return nil, fmt.Errorf("GetStorageState key:%x", key)
	}
	path, err := merkle.MerkleLeafPath(state, hashes)
	if err != nil {
		return nil, err
	}
	return path, nil
}

func (this *LedgerStoreImp) saveBlockToBlockStore(block *types.Block) error {
	blockHash := block.Hash()
	blockHeight := block.Header.Height

	this.setHeaderIndex(blockHeight, blockHash)
	err := this.saveHeaderIndexList()
	if err != nil {
		return fmt.Errorf("saveHeaderIndexList error %s", err)
	}
	err = this.blockStore.SaveCurrentBlock(blockHeight, blockHash)
	if err != nil {
		return fmt.Errorf("SaveCurrentBlock error %s", err)
	}
	this.blockStore.SaveBlockHash(blockHeight, blockHash)
	err = this.blockStore.SaveBlock(block)
	if err != nil {
		return fmt.Errorf("SaveBlock height %d hash %s error %s", blockHeight, blockHash.ToHexString(), err)
	}
	return nil
}

func (this *LedgerStoreImp) executeBlock(block *types.Block) (result store.ExecuteResult, err error) {
	overlay := this.stateStore.NewOverlayDB()

	cache := storage.NewCacheDB(overlay)
	for _, tx := range block.Transactions {
		cache.Reset()
		notify, crossHashes, e := this.handleTransaction(overlay, cache, block, tx.Payload)
		if e != nil {
			err = e
			return
		}
		result.Notify = append(result.Notify, notify)
		result.CrossHashes = append(result.CrossHashes, crossHashes...)
	}
	if len(result.CrossHashes) != 0 {
		result.CrossStatesRoot = merkle.TreeHasher{}.HashFullTreeWithLeafHash(result.CrossHashes)
	} else {
		result.CrossStatesRoot = common.UINT256_EMPTY
	}
	result.Hash = overlay.ChangeHash()
	result.WriteSet = overlay.GetWriteSet()
	result.MerkleRoot = this.stateStore.GetStateMerkleRootWithNewHash(result.Hash)
	return
}

func (this *LedgerStoreImp) saveBlockToStateStore(block *types.Block, result store.ExecuteResult) error {
	blockHash := block.Hash()
	blockHeight := block.Header.Height

	for _, notify := range result.Notify {
		err := SaveNotify(this.eventStore, notify.TxHash, notify)
		if err != nil {
			return fmt.Errorf("SaveNotify error %s", err)
		}
	}

	err := this.stateStore.AddStateMerkleTreeRoot(blockHeight, result.Hash)
	if err != nil {
		return fmt.Errorf("AddBlockMerkleTreeRoot error %s", err)
	}

	err = this.stateStore.AddBlockMerkleTreeRoot(block.Header.PrevBlockHash)
	if err != nil {
		return fmt.Errorf("AddBlockMerkleTreeRoot error %s", err)
	}

	err = this.stateStore.SaveCurrentBlock(blockHeight, blockHash)
	if err != nil {
		return fmt.Errorf("SaveCurrentBlock error %s", err)
	}

	err = this.stateStore.AddCrossStates(blockHeight, result.CrossHashes, result.CrossStatesRoot)
	if err != nil {
		return err
	}

	log.Debugf("the state transition hash of block %d is:%s", blockHeight, result.Hash.ToHexString())

	result.WriteSet.ForEach(func(key, val []byte) {
		if len(val) == 0 {
			this.stateStore.BatchDeleteRawKey(key)
		} else {
			this.stateStore.BatchPutRawKeyVal(key, val)
		}
	})

	return nil
}

func (this *LedgerStoreImp) saveBlockToEventStore(block *types.Block) error {
	blockHash := block.Hash()
	blockHeight := block.Header.Height
	txs := make([]common.Uint256, 0)
	for _, tx := range block.Transactions {
		txHash := tx.Hash()
		txs = append(txs, txHash)
	}
	if len(txs) > 0 {
		err := this.eventStore.SaveEventNotifyByBlock(block.Header.Height, txs)
		if err != nil {
			return fmt.Errorf("SaveEventNotifyByBlock error %s", err)
		}
	}
	err := this.eventStore.SaveCurrentBlock(blockHeight, blockHash)
	if err != nil {
		return fmt.Errorf("SaveCurrentBlock error %s", err)
	}
	return nil
}

func (this *LedgerStoreImp) tryGetSavingBlockLock() (hasLocked bool) {
	select {
	case this.savingBlockSemaphore <- true:
		return false
	default:
		return true
	}
}

func (this *LedgerStoreImp) getSavingBlockLock() {
	this.savingBlockSemaphore <- true
}

func (this *LedgerStoreImp) releaseSavingBlockLock() {
	select {
	case <-this.savingBlockSemaphore:
		return
	default:
		panic("can not release in unlocked state")
	}
}

// saveBlock do the job of execution samrt contract and commit block to store.
func (this *LedgerStoreImp) submitBlock(block *types.Block, result store.ExecuteResult) error {
	blockHash := block.Hash()
	blockHeight := block.Header.Height

	// blockRoot := this.GetBlockRootWithPreBlockHashes(block.Header.Height, []common.Uint256{block.Header.PrevBlockHash})
	// if block.Header.Height != 0 && blockRoot != block.Header.BlockRoot {
	// 	return fmt.Errorf("wrong block root at height:%d, expected:%s, got:%s",
	// 		block.Header.Height, blockRoot.ToHexString(), block.Header.BlockRoot.ToHexString())
	// }

	this.blockStore.NewBatch()
	this.stateStore.NewBatch()
	this.eventStore.NewBatch()
	err := this.saveBlockToBlockStore(block)
	if err != nil {
		return fmt.Errorf("save to block store height:%d error:%s", blockHeight, err)
	}
	err = this.saveBlockToStateStore(block, result)
	if err != nil {
		return fmt.Errorf("save to state store height:%d error:%s", blockHeight, err)
	}
	err = this.saveBlockToEventStore(block)
	if err != nil {
		return fmt.Errorf("save to event store height:%d error:%s", blockHeight, err)
	}
	err = this.blockStore.CommitTo()
	if err != nil {
		return fmt.Errorf("blockStore.CommitTo height:%d error %s", blockHeight, err)
	}
	// event store is idempotent to re-save when in recovering process, so save first before stateStore
	err = this.eventStore.CommitTo()
	if err != nil {
		return fmt.Errorf("eventStore.CommitTo height:%d error %s", blockHeight, err)
	}
	err = this.stateStore.CommitTo()
	if err != nil {
		return fmt.Errorf("stateStore.CommitTo height:%d error %s", blockHeight, err)
	}
	this.setCurrentBlock(blockHeight, blockHash)

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
func (this *LedgerStoreImp) saveBlock(block *types.Block, stateMerkleRoot common.Uint256) error {
	blockHeight := block.Header.Height
	if this.tryGetSavingBlockLock() {
		// hash already saved or is saving
		return nil
	}
	defer this.releaseSavingBlockLock()
	if blockHeight > 0 && blockHeight != (this.GetCurrentBlockHeight()+1) {
		return nil
	}

	result, err := this.executeBlock(block)
	if err != nil {
		return err
	}

	if result.MerkleRoot != stateMerkleRoot {
		return fmt.Errorf("state merkle root mismatch!")
	}

	return this.submitBlock(block, result)
}

func (this *LedgerStoreImp) handleTransaction(overlay *overlaydb.OverlayDB, cache *storage.CacheDB, block *types.Block, txPayload payload.Payload) (*event.ExecuteNotify, []common.Uint256, error) {
	tx := types.ToTransaction(txPayload)
	txHash := tx.Hash()
	notify := &event.ExecuteNotify{TxHash: txHash, State: event.CONTRACT_STATE_FAIL}
	if tx.TxType() == payload.InvokeType {
		crossHashes, err := this.stateStore.HandleInvokeTransaction(this, overlay, cache, txPayload, block, notify)
		if overlay.Error() != nil {
			return nil, nil, fmt.Errorf("HandleInvokeTransaction tx %s error %s", txHash.ToHexString(), overlay.Error())
		}
		if err != nil {
			log.Debugf("HandleInvokeTransaction tx %s error %s", txHash.ToHexString(), err)
		}
		return notify, crossHashes, nil
	} else if tx.TxType() == payload.EpochType {
		crossHashes, err := this.stateStore.HandleEpochTransaction(this, overlay, cache, txPayload, block, notify)
		if overlay.Error() != nil {
			return nil, nil, fmt.Errorf("HandleInvokeTransaction tx %s error %s", txHash.ToHexString(), overlay.Error())
		}
		if err != nil {
			log.Debugf("HandleInvokeTransaction tx %s error %s", txHash.ToHexString(), err)
		}
		return notify, crossHashes, nil
	} else if tx.TxType() == payload.BridgeEventType || tx.TxType() == payload.BridgeEventSolanaType ||
		tx.TxType() == payload.SolanaToEVMEventType || tx.TxType() == payload.ReceiveRequestEventType {
		crossHashes, err := this.stateStore.HandleBridgeTransaction(this, overlay, cache, txPayload, block, notify)
		if overlay.Error() != nil {
			return nil, nil, fmt.Errorf("HandleBridgeTransaction tx %s error %s", txHash.ToHexString(), overlay.Error())
		}
		if err != nil {
			return nil, nil, fmt.Errorf("HandleBridgeTransaction tx %s error %s", txHash.ToHexString(), err)
		}
		return notify, crossHashes, nil
	} else {
		return nil, nil, fmt.Errorf("Unsupported transaction type! type=%v payload=%v", tx.TxType(), tx)
	}
}

func (this *LedgerStoreImp) saveHeaderIndexList() error {
	this.lock.RLock()
	storeCount := this.storedIndexCount
	currHeight := this.currBlockHeight
	if currHeight-storeCount < HEADER_INDEX_BATCH_SIZE {
		this.lock.RUnlock()
		return nil
	}

	headerList := make([]common.Uint256, HEADER_INDEX_BATCH_SIZE)
	for i := uint64(0); i < HEADER_INDEX_BATCH_SIZE; i++ {
		height := storeCount + i
		headerList[i] = this.headerIndex[height]
	}
	this.lock.RUnlock()

	err := this.blockStore.SaveHeaderIndexList(storeCount, headerList)
	if err != nil {
		return fmt.Errorf("SaveHeaderIndexList start %d error %s", storeCount, err)
	}

	this.lock.Lock()
	this.storedIndexCount += HEADER_INDEX_BATCH_SIZE
	this.lock.Unlock()
	return nil
}

func (this *LedgerStoreImp) PreExecuteContract(tx payload.Payload) (*cstates.PreExecResult, error) {
	result := &sstate.PreExecResult{State: event.CONTRACT_STATE_FAIL, Result: nil}
	if _, ok := tx.(*payload.InvokeCode); !ok {
		return result, fmt.Errorf("transaction payload type error")
	}
	hash := this.GetCurrentBlockHash()
	block, err := this.GetBlockByHash(hash)
	if err != nil {
		return result, fmt.Errorf("get current block error")
	}
	overlay := this.stateStore.NewOverlayDB()
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
func (this *LedgerStoreImp) IsContainBlock(blockHash common.Uint256) (bool, error) {
	return this.blockStore.ContainBlock(blockHash)
}

// IsContainTransaction return whether the transaction is in store. Wrap function of BlockStore.ContainTransaction
func (this *LedgerStoreImp) IsContainTransaction(txHash common.Uint256) (bool, error) {
	return this.blockStore.ContainTransaction(txHash)
}

func (this *LedgerStoreImp) GetBlockRootWithPreBlockHashes(startHeight uint64, preBlockHashes []common.Uint256) common.Uint256 {
	this.lock.RLock()
	defer this.lock.RUnlock()

	// the block height in consensus is far behind ledger, this case should be rare
	if this.currBlockHeight > startHeight+uint64(len(preBlockHashes))-1 {
		// or return error?
		log.Errorf("this.currBlockHeight= %d, startHeight= %d, len(preBlockHashes)= %d\n", this.currBlockHeight, startHeight, len(preBlockHashes))
		return common.UINT256_EMPTY
	}

	needs := preBlockHashes[this.currBlockHeight+1-startHeight:]
	return this.stateStore.GetBlockRootWithPreBlockHashes(needs)
}

// GetBlockHash return the block hash by block height
func (this *LedgerStoreImp) GetBlockHash(height uint64) common.Uint256 {
	return this.getHeaderIndex(height)
}

// GetHeaderByHash return the block header by block hash
func (this *LedgerStoreImp) GetHeaderByHash(blockHash common.Uint256) (*types.Header, error) {
	header := this.getHeaderCache(blockHash)
	if header != nil {
		return header, nil
	}
	return this.blockStore.GetHeader(blockHash)
}

// GetHeaderByHash return the block header by block height
func (this *LedgerStoreImp) GetHeaderByHeight(height uint64) (*types.Header, error) {
	blockHash := this.GetBlockHash(height)
	var empty common.Uint256
	if blockHash == empty {
		return nil, nil
	}
	return this.GetHeaderByHash(blockHash)
}

// GetTransaction return transaction by transaction hash. Wrap function of BlockStore.GetTransaction
func (this *LedgerStoreImp) GetTransaction(txHash common.Uint256) (payload.Payload, uint64, error) {
	return this.blockStore.GetTransaction(txHash)
}

// GetBlockByHash return block by block hash. Wrap function of BlockStore.GetBlockByHash
func (this *LedgerStoreImp) GetBlockByHash(blockHash common.Uint256) (*types.Block, error) {
	return this.blockStore.GetBlock(blockHash)
}

// GetBlockByHeight return block by height.
func (this *LedgerStoreImp) GetBlockByHeight(height uint64) (*types.Block, error) {
	blockHash := this.GetBlockHash(height)
	var empty common.Uint256
	if blockHash == empty {
		return nil, nil
	}
	return this.GetBlockByHash(blockHash)
}

// GetEpochState return the bookkeeper state. Wrap function of StateStore.GetEpochState
func (this *LedgerStoreImp) GetEpochState() (*states.EpochState, error) {
	return this.stateStore.GetEpochState()
}

// GetMerkleProof return the block merkle proof. Wrap function of StateStore.GetMerkleProof
func (this *LedgerStoreImp) GetMerkleProof(raw []byte, proofHeight, rootHeight uint64) ([]byte, error) {
	return this.stateStore.GetMerkleProof(raw, proofHeight, rootHeight)
}

// GetStorageItem return the storage value of the key in smart contract. Wrap function of StateStore.GetStorageState
func (this *LedgerStoreImp) GetStorageItem(key *states.StorageKey) (*states.StorageItem, error) {
	return this.stateStore.GetStorageState(key)
}

// GetEventNotifyByTx return the events notify gen by executing of smart contract.  Wrap function of EventStore.GetEventNotifyByTx
func (this *LedgerStoreImp) GetEventNotifyByTx(tx common.Uint256) (*event.ExecuteNotify, error) {
	return this.eventStore.GetEventNotifyByTx(tx)
}

// GetEventNotifyByBlock return the transaction hash which have event notice after execution of smart contract. Wrap function of EventStore.GetEventNotifyByBlock
func (this *LedgerStoreImp) GetEventNotifyByBlock(height uint64) ([]*event.ExecuteNotify, error) {
	return this.eventStore.GetEventNotifyByBlock(height)
}

// Close ledger store.
func (this *LedgerStoreImp) Close() error {
	err := this.blockStore.Close()
	if err != nil {
		return fmt.Errorf("blockStore close error %s", err)
	}
	err = this.stateStore.Close()
	if err != nil {
		return fmt.Errorf("stateStore close error %s", err)
	}
	err = this.eventStore.Close()
	if err != nil {
		return fmt.Errorf("eventStore close error %s", err)
	}
	return nil
}
