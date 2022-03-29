package ledgerstore

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/common/serialization"
	"github.com/eywa-protocol/chain/core/states"
	scom "github.com/eywa-protocol/chain/core/store/common"
	"github.com/eywa-protocol/chain/core/store/leveldbstore"
	"github.com/eywa-protocol/chain/core/store/overlaydb"
	"github.com/eywa-protocol/chain/merkle"
	"github.com/sirupsen/logrus"
)

var (
	BOOKKEEPER = []byte("Epoch") // Epoch store key
)

// StateStore saving the data of ledger states. Like balance of account, and the execution result of smart contract
type StateStore struct {
	dbDir                string                    // Store file path
	store                scom.PersistStore         // Store handler
	merklePath           string                    // Merkle tree store path
	merkleTree           *merkle.CompactMerkleTree // Merkle tree of block root
	deltaMerkleTree      *merkle.CompactMerkleTree // Merkle tree of delta state root
	merkleHashStore      merkle.HashStore
	stateHashCheckHeight uint64
}

// NewStateStore return state store instance
func NewStateStore(dbDir, merklePath string) (*StateStore, error) {
	var err error
	store, err := leveldbstore.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	stateStore := &StateStore{
		dbDir:      dbDir,
		store:      store,
		merklePath: merklePath,
	}
	_, height, err := stateStore.GetCurrentBlock()
	if err != nil && err != scom.ErrNotFound {
		return nil, fmt.Errorf("GetCurrentBlock error %s", err)
	}
	err = stateStore.init(height)
	if err != nil {
		return nil, fmt.Errorf("init error %s", err)
	}
	return stateStore, nil
}

// NewMemStateStore for test
func NewMemStateStore(stateHashHeight uint64) *StateStore {
	store, _ := leveldbstore.NewMemLevelDBStore()
	stateStore := &StateStore{
		store:                store,
		merkleTree:           merkle.NewTree(0, nil, nil),
		deltaMerkleTree:      merkle.NewTree(0, nil, nil),
		stateHashCheckHeight: stateHashHeight,
	}

	return stateStore
}

// NewBatch start new commit batch
func (s *StateStore) NewBatch() {
	s.store.NewBatch()
}

func (s *StateStore) BatchPutRawKeyVal(key, val []byte) {
	s.store.BatchPut(key, val)
}

func (s *StateStore) BatchDeleteRawKey(key []byte) {
	s.store.BatchDelete(key)
}

func (s *StateStore) init(currBlockHeight uint64) error {
	treeSize, hashes, err := s.GetBlockMerkleTree()
	if err != nil && err != scom.ErrNotFound {
		return err
	}
	if treeSize > 0 && treeSize != currBlockHeight+1 {
		return fmt.Errorf("merkle tree size is inconsistent with blockheight: %d", currBlockHeight+1)
	}
	s.merkleHashStore, err = merkle.NewFileHashStore(s.merklePath, treeSize)
	if err != nil {
		logrus.Warn("merkle store is inconsistent with ChainStore. persistence will be disabled")
	}
	s.merkleTree = merkle.NewTree(treeSize, hashes, s.merkleHashStore)

	if currBlockHeight >= s.stateHashCheckHeight {
		treeSize, hashes, err := s.GetStateMerkleTree()
		if err != nil && err != scom.ErrNotFound {
			return err
		}
		if treeSize > 0 && treeSize != currBlockHeight-s.stateHashCheckHeight+1 {
			return fmt.Errorf("merkle tree size is inconsistent with blockheight: %d", currBlockHeight+1)
		}
		s.deltaMerkleTree = merkle.NewTree(treeSize, hashes, nil)
	}
	return nil
}

// GetStateMerkleTree return merkle tree size an tree node
func (s *StateStore) GetStateMerkleTree() (uint64, []common.Uint256, error) {
	key := s.genStateMerkleTreeKey()
	return s.getMerkleTree(key)
}

// GetBlockMerkleTree return merkle tree size an tree node
func (s *StateStore) GetBlockMerkleTree() (uint64, []common.Uint256, error) {
	key := s.genBlockMerkleTreeKey()
	return s.getMerkleTree(key)
}
func (s *StateStore) getMerkleTree(key []byte) (uint64, []common.Uint256, error) {
	data, err := s.store.Get(key)
	if err != nil {
		return 0, nil, err
	}
	value := bytes.NewBuffer(data)
	treeSize, err := serialization.ReadUint64(value)
	if err != nil {
		return 0, nil, err
	}
	hashCount := (len(data) - 8) / common.UINT256_SIZE
	hashes := make([]common.Uint256, 0, hashCount)
	for i := 0; i < hashCount; i++ {
		var hash = new(common.Uint256)
		err = hash.Deserialize(value)
		if err != nil {
			return 0, nil, err
		}
		hashes = append(hashes, *hash)
	}
	return treeSize, hashes, nil
}

func (s *StateStore) GetStateMerkleRoot(height uint64) (result common.Uint256, err error) {
	if height < s.stateHashCheckHeight {
		return
	}
	key := s.genStateMerkleRootKey(height)
	var value []byte
	value, err = s.store.Get(key)
	if err != nil {
		return
	}
	source := common.NewZeroCopySource(value)
	_, eof := source.NextHash()
	result, eof = source.NextHash()
	if eof {
		err = io.ErrUnexpectedEOF
	}
	return
}

func (s *StateStore) AddStateMerkleTreeRoot(blockHeight uint64, writeSetHash common.Uint256) error {
	if blockHeight < s.stateHashCheckHeight {
		return nil
	} else if blockHeight == s.stateHashCheckHeight {
		s.deltaMerkleTree = merkle.NewTree(0, nil, nil)
	}
	key := s.genStateMerkleTreeKey()

	s.deltaMerkleTree.Append(writeSetHash.ToArray())
	treeSize := s.deltaMerkleTree.TreeSize()
	hashes := s.deltaMerkleTree.Hashes()
	value := common.NewZeroCopySink(make([]byte, 0, 8+len(hashes)*common.UINT256_SIZE))
	value.WriteUint64(treeSize)
	for _, hash := range hashes {
		value.WriteHash(hash)
	}
	s.store.BatchPut(key, value.Bytes())

	key = s.genStateMerkleRootKey(blockHeight)
	value.Reset()
	value.WriteHash(writeSetHash)
	value.WriteHash(s.deltaMerkleTree.Root())
	s.store.BatchPut(key, value.Bytes())

	return nil
}

func (s *StateStore) AddCrossStates(height uint64, crossStates []common.Uint256, crossStatesHash common.Uint256) error {
	if len(crossStates) == 0 {
		return nil
	}
	key := genCrossStatesKey(height)
	sink := common.NewZeroCopySink(make([]byte, 0, len(crossStates)*common.UINT256_SIZE))
	for _, v := range crossStates {
		sink.WriteHash(v)
	}
	s.store.BatchPut(key, sink.Bytes())

	buf := bytes.NewBuffer(nil)
	err := crossStatesHash.Serialize(buf)
	if err != nil {
		return err
	}
	s.store.BatchPut(genCrossStatesRootKey(height), buf.Bytes())
	return nil
}

func (s *StateStore) GetCrossStateRoot(height uint64) (common.Uint256, error) {
	var hash common.Uint256
	key := genCrossStatesRootKey(height)
	value, err := s.store.Get(key)
	if err != nil && err != scom.ErrNotFound {
		return common.UINT256_EMPTY, err
	}
	if err == scom.ErrNotFound {
		return common.UINT256_EMPTY, nil
	}
	buf := bytes.NewBuffer(value)
	err = hash.Deserialize(buf)
	if err != nil {
		return common.UINT256_EMPTY, err
	}
	return hash, nil
}

func (s *StateStore) GetCrossStates(height uint64) (hashes []common.Uint256, err error) {
	key := genCrossStatesKey(height)

	var value []byte
	value, err = s.store.Get(key)
	if err != nil {
		return
	}

	source := common.NewZeroCopySource(value)

	l := int(source.Size() / common.UINT256_SIZE)

	hashes = make([]common.Uint256, 0, l)

	for i := 0; i < l; i++ {
		u256, eof := source.NextHash()
		if eof {
			err = io.ErrUnexpectedEOF
			return
		}
		hashes = append(hashes, u256)
	}
	return
}

// AddBlockMerkleTreeRoot add a new tree root
func (s *StateStore) AddBlockMerkleTreeRoot(preBlockHash common.Uint256) error {
	key := s.genBlockMerkleTreeKey()

	s.merkleTree.Append(preBlockHash.ToArray())
	treeSize := s.merkleTree.TreeSize()
	hashes := s.merkleTree.Hashes()
	value := common.NewZeroCopySink(make([]byte, 0, 8+len(hashes)*common.UINT256_SIZE))
	value.WriteUint64(treeSize)
	for _, hash := range hashes {
		value.WriteHash(hash)
	}
	s.store.BatchPut(key, value.Bytes())
	return nil
}

// GetMerkleProof return merkle proof of block hash
func (s *StateStore) GetMerkleProof(raw []byte, proofHeight, rootHeight uint64) ([]byte, error) {
	return s.merkleTree.MerkleInclusionLeafPath(raw, proofHeight, rootHeight+1)
}

func (s *StateStore) NewOverlayDB() *overlaydb.OverlayDB {
	return overlaydb.NewOverlayDB(s.store)
}

// CommitTo commit state batch to state store
func (s *StateStore) CommitTo() error {
	return s.store.BatchCommit()
}

// GetEpochState return current book keeper states
func (s *StateStore) GetEpochState() (*states.EpochState, error) {
	key, err := s.getEpochKey()
	if err != nil {
		return nil, err
	}

	value, err := s.store.Get(key)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(value)
	bookkeeperState := new(states.EpochState)
	err = bookkeeperState.Deserialize(reader)
	if err != nil {
		return nil, err
	}
	return bookkeeperState, nil
}

// SaveEpochState persist book keeper state to store
func (s *StateStore) SaveEpochState(bookkeeperState *states.EpochState) error {
	key, err := s.getEpochKey()
	if err != nil {
		return err
	}
	value := bytes.NewBuffer(nil)
	err = bookkeeperState.Serialize(value)
	if err != nil {
		return err
	}

	return s.store.Put(key, value.Bytes())
}

// GetStorageState return the storage value of the key in smart contract.
func (s *StateStore) GetStorageState(key *states.StorageKey) (*states.StorageItem, error) {
	storeKey, err := s.getStorageKey(key)
	if err != nil {
		return nil, err
	}

	data, err := s.store.Get(storeKey)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(data)
	storageState := new(states.StorageItem)
	err = storageState.Deserialize(reader)
	if err != nil {
		return nil, err
	}
	return storageState, nil
}

func (s *StateStore) GetStorageValue(key []byte) ([]byte, error) {
	data, err := s.store.Get(append([]byte{byte(scom.ST_STORAGE)}, key...))
	if err != nil {
		return nil, err
	}
	reader := bytes.NewBuffer(data)
	storageState := new(states.StorageItem)
	err = storageState.Deserialize(reader)
	if err != nil {
		return nil, err
	}
	return storageState.Value, nil
}

// GetCurrentBlock return current block height and current hash in state store
func (s *StateStore) GetCurrentBlock() (common.Uint256, uint64, error) {
	key := s.getCurrentBlockKey()
	data, err := s.store.Get(key)
	if err != nil {
		return common.Uint256{}, 0, err
	}
	reader := bytes.NewReader(data)
	blockHash := common.Uint256{}
	err = blockHash.Deserialize(reader)
	if err != nil {
		return common.Uint256{}, 0, err
	}
	height, err := serialization.ReadUint64(reader)
	if err != nil {
		return common.Uint256{}, 0, err
	}
	return blockHash, height, nil
}

// SaveCurrentBlock persist current block to state store
func (s *StateStore) SaveCurrentBlock(height uint64, blockHash common.Uint256) error {
	key := s.getCurrentBlockKey()
	value := bytes.NewBuffer(nil)
	if err := blockHash.Serialize(value); err != nil {
		return err
	}
	if err := serialization.WriteUint64(value, height); err != nil {
		return err
	}
	s.store.BatchPut(key, value.Bytes())
	return nil
}

func (s *StateStore) SaveProcessedHeight(height uint64) error {
	key := s.getProcessedHeightKey()
	value := bytes.NewBuffer(nil)
	if err := serialization.WriteUint64(value, height); err != nil {
		return err
	}
	s.store.BatchPut(key, value.Bytes())
	return nil
}

func (s *StateStore) getProcessedHeight() (uint64, error) {
	key := s.getProcessedHeightKey()
	if value, err := s.store.Get(key); err != nil {
		return 0, err
	} else if height, err := serialization.ReadUint64(bytes.NewReader(value)); err != nil {
		return 0, err
	} else {
		return height, nil
	}
}

func (s *StateStore) getCurrentBlockKey() []byte {
	return []byte{byte(scom.SYS_CURRENT_BLOCK)}
}

func (s *StateStore) getProcessedHeightKey() []byte {
	return []byte{byte(scom.SYS_PROCESSED_SRC_HEIGHT)}
}

func (s *StateStore) getEpochKey() ([]byte, error) {
	key := make([]byte, 1+len(BOOKKEEPER))
	key[0] = byte(scom.ST_BOOKKEEPER)
	copy(key[1:], BOOKKEEPER)
	return key, nil
}

func (s *StateStore) getContractStateKey(contractHash common.Address) ([]byte, error) {
	data := contractHash[:]
	key := make([]byte, 1+len(data))
	key[0] = byte(scom.ST_CONTRACT)
	copy(key[1:], data)
	return key, nil
}

func (s *StateStore) getStorageKey(key *states.StorageKey) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(byte(scom.ST_STORAGE))
	buf.Write(key.ContractAddress[:])
	buf.Write(key.Key)
	return buf.Bytes(), nil
}

func (s *StateStore) GetStateMerkleRootWithNewHash(writeSetHash common.Uint256) common.Uint256 {
	return s.deltaMerkleTree.GetRootWithNewLeaf(writeSetHash)
}

func (s *StateStore) GetBlockRootWithPreBlockHashes(preBlockHashes []common.Uint256) common.Uint256 {
	return s.merkleTree.GetRootWithNewLeaves(preBlockHashes)
}

func (s *StateStore) genBlockMerkleTreeKey() []byte {
	return []byte{byte(scom.SYS_BLOCK_MERKLE_TREE)}
}

func (s *StateStore) genStateMerkleTreeKey() []byte {
	return []byte{byte(scom.SYS_STATE_MERKLE_TREE)}
}

func genCrossStatesKey(height uint64) []byte {
	key := make([]byte, 9, 9)
	key[0] = byte(scom.SYS_CROSS_STATES)
	binary.LittleEndian.PutUint64(key[1:], height)
	return key
}

func genCrossStatesRootKey(height uint64) []byte {
	key := make([]byte, 9, 9)
	key[0] = byte(scom.SYS_CROSS_STATES_HASH)
	binary.LittleEndian.PutUint64(key[1:], height)
	return key
}

func (s *StateStore) genStateMerkleRootKey(height uint64) []byte {
	key := make([]byte, 9, 9)
	key[0] = byte(scom.DATA_STATE_MERKLE_ROOT)
	binary.LittleEndian.PutUint64(key[1:], height)
	return key
}

// ClearAll clear all data in state store
func (s *StateStore) ClearAll() error {
	s.store.NewBatch()
	iter := s.store.NewIterator(nil)
	for iter.Next() {
		s.store.BatchDelete(iter.Key())
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		s.store.NewBatch() // reset the batch
		return err
	}
	return s.store.BatchCommit()
}

// Close state store
func (s *StateStore) Close() error {
	s.merkleHashStore.Close()
	return s.store.Close()
}
