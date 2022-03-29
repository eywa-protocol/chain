package ledgerstore

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/common/serialization"
	"github.com/eywa-protocol/chain/core/payload"
	scom "github.com/eywa-protocol/chain/core/store/common"
	"github.com/eywa-protocol/chain/core/store/leveldbstore"
	"github.com/eywa-protocol/chain/core/types"
)

// BlockStore Block store save the data of block & transaction
type BlockStore struct {
	enableCache bool                       // Is enable lru cache
	dbDir       string                     // The path of store file
	cache       *BlockCache                // The cache of block, if have.
	store       *leveldbstore.LevelDBStore // block store handler
}

// NewBlockStore return the block store instance
func NewBlockStore(dbDir string, enableCache bool) (*BlockStore, error) {
	var cache *BlockCache
	var err error
	if enableCache {
		cache, err = NewBlockCache()
		if err != nil {
			return nil, fmt.Errorf("NewBlockCache error %s", err)
		}
	}

	store, err := leveldbstore.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	blockStore := &BlockStore{
		dbDir:       dbDir,
		enableCache: enableCache,
		store:       store,
		cache:       cache,
	}
	return blockStore, nil
}

// NewBatch start a commit batch
func (s *BlockStore) NewBatch() {
	s.store.NewBatch()
}

// SaveBlock persist block to store
func (s *BlockStore) SaveBlock(block *types.Block) error {
	if s.enableCache {
		s.cache.AddBlock(block)
	}

	blockHeight := block.Header.Height
	err := s.SaveHeader(block)
	if err != nil {
		return fmt.Errorf("SaveHeader error %s", err)
	}
	for _, tx := range block.Transactions {
		err = s.SaveTransaction(tx.Payload, blockHeight)
		if err != nil {
			txHash := tx.Hash()
			return fmt.Errorf("SaveTransaction block height %d tx %s err %s", blockHeight, txHash.ToHexString(), err)
		}
	}
	return nil
}

// ContainBlock return the block specified by block hash save in store
func (s *BlockStore) ContainBlock(blockHash common.Uint256) (bool, error) {
	if s.enableCache {
		if s.cache.ContainBlock(blockHash) {
			return true, nil
		}
	}
	key := s.getHeaderKey(blockHash)
	_, err := s.store.Get(key)
	if err != nil {
		if err == scom.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetBlock return block by block hash
func (s *BlockStore) GetBlock(blockHash common.Uint256) (*types.Block, error) {
	var block *types.Block
	if s.enableCache {
		block = s.cache.GetBlock(blockHash)
		if block != nil {
			return block, nil
		}
	}
	header, txHashes, err := s.loadHeaderWithTx(blockHash)
	if err != nil {
		return nil, err
	}
	txList := make(types.Transactions, 0, len(txHashes))
	for _, txHash := range txHashes {

		tx, _, err := s.GetTransaction(txHash)
		if err != nil {
			return nil, fmt.Errorf("GetTransaction %s error %s", txHash.ToHexString(), err)
		}
		if tx == nil {
			return nil, fmt.Errorf("cannot get transaction %s", txHash.ToHexString())
		}
		txList = append(txList, types.ToTransaction(tx))
	}
	block = &types.Block{
		Header:       header,
		Transactions: txList,
	}
	return block, nil
}

func (s *BlockStore) loadHeaderWithTx(blockHash common.Uint256) (*types.Header, []common.Uint256, error) {
	key := s.getHeaderKey(blockHash)
	value, err := s.store.Get(key)
	if err != nil {
		return nil, nil, err
	}
	source := common.NewZeroCopySource(value)
	header := new(types.Header)
	err = header.Deserialization(source)
	if err != nil {
		return nil, nil, err
	}
	txSize, eof := source.NextUint32()
	if eof {
		return nil, nil, io.ErrUnexpectedEOF
	}
	txHashes := make([]common.Uint256, 0, int(txSize))
	for i := uint32(0); i < txSize; i++ {
		txHash, eof := source.NextHash()
		if eof {
			return nil, nil, io.ErrUnexpectedEOF
		}
		txHashes = append(txHashes, txHash)
	}
	return header, txHashes, nil
}

// SaveHeader persist block header to store
func (s *BlockStore) SaveHeader(block *types.Block) error {
	blockHash := block.Hash()
	key := s.getHeaderKey(blockHash)
	sink := common.NewZeroCopySink(nil)
	if err := block.Header.Serialization(sink); err != nil {
		return err
	}
	sink.WriteUint32(uint32(len(block.Transactions)))
	for _, tx := range block.Transactions {
		txHash := tx.Hash()
		sink.WriteHash(txHash)
	}
	s.store.BatchPut(key, sink.Bytes())
	return nil
}

// GetHeader return the header specified by block hash
func (s *BlockStore) GetHeader(blockHash common.Uint256) (*types.Header, error) {
	if s.enableCache {
		block := s.cache.GetBlock(blockHash)
		if block != nil {
			return block.Header, nil
		}
	}
	return s.loadHeader(blockHash)
}

func (s *BlockStore) loadHeader(blockHash common.Uint256) (*types.Header, error) {
	key := s.getHeaderKey(blockHash)
	value, err := s.store.Get(key)
	if err != nil {
		return nil, err
	}
	source := common.NewZeroCopySource(value)
	header := new(types.Header)
	err = header.Deserialization(source)
	if err != nil {
		return nil, err
	}
	return header, nil
}

// GetCurrentBlock return the current block hash and current block height
func (s *BlockStore) GetCurrentBlock() (common.Uint256, uint64, error) {
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

// SaveCurrentBlock persist the current block height and current block hash to store
func (s *BlockStore) SaveCurrentBlock(height uint64, blockHash common.Uint256) error {
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

// GetHeaderIndexList return the head index store in header index list
func (s *BlockStore) GetHeaderIndexList() (map[uint64]common.Uint256, error) {
	result := make(map[uint64]common.Uint256)
	iter := s.store.NewIterator([]byte{byte(scom.IX_HEADER_HASH_LIST)})
	defer iter.Release()
	for iter.Next() {
		startCount, err := s.getStartHeightByHeaderIndexKey(iter.Key())
		if err != nil {
			return nil, fmt.Errorf("getStartHeightByHeaderIndexKey error %s", err)
		}
		reader := bytes.NewReader(iter.Value())
		count, err := serialization.ReadUint64(reader)
		if err != nil {
			return nil, fmt.Errorf("serialization.ReadUint64 count error %s", err)
		}
		for i := uint64(0); i < count; i++ {
			height := startCount + i
			blockHash := common.Uint256{}
			err = blockHash.Deserialize(reader)
			if err != nil {
				return nil, fmt.Errorf("blockHash.Deserialize error %s", err)
			}
			result[height] = blockHash
		}
	}
	if err := iter.Error(); err != nil {
		return nil, err
	}
	return result, nil
}

// SaveHeaderIndexList persist header index list to store
func (s *BlockStore) SaveHeaderIndexList(startIndex uint64, indexList []common.Uint256) error {
	indexKey, err := s.getHeaderIndexListKey(startIndex)
	if err != nil {
		return err
	}
	indexSize := uint64(len(indexList))
	value := bytes.NewBuffer(nil)
	if err := serialization.WriteUint64(value, indexSize); err != nil {
		return err
	}
	for _, hash := range indexList {
		if err := hash.Serialize(value); err != nil {
			return err
		}
	}

	s.store.BatchPut(indexKey, value.Bytes())
	return nil
}

// GetBlockHash return block hash by block height
func (s *BlockStore) GetBlockHash(height uint64) (common.Uint256, error) {
	key := s.getBlockHashKey(height)
	value, err := s.store.Get(key)
	if err != nil {
		return common.UINT256_EMPTY, err
	}
	blockHash, err := common.Uint256ParseFromBytes(value)
	if err != nil {
		return common.UINT256_EMPTY, err
	}
	return blockHash, nil
}

// SaveBlockHash persist block height and block hash to store
func (s *BlockStore) SaveBlockHash(height uint64, blockHash common.Uint256) {
	key := s.getBlockHashKey(height)
	s.store.BatchPut(key, blockHash.ToArray())
}

// SaveTransaction persist transaction to store
func (s *BlockStore) SaveTransaction(tx payload.Payload, height uint64) error {
	if s.enableCache {
		s.cache.AddTransaction(tx, height)
	}
	return s.putTransaction(tx, height)
}

func (s *BlockStore) putTransaction(payload payload.Payload, height uint64) error {
	tx := types.ToTransaction(payload)
	txHash := tx.Hash()

	key, err := s.getTransactionKey(txHash)
	if err != nil {
		return err
	}
	value := bytes.NewBuffer(nil)

	if err := serialization.WriteUint64(value, height); err != nil {
		return err
	}

	if err := serialization.WriteBytes(value, tx.ToArray()); err != nil {
		return err
	}

	s.store.BatchPut(key, value.Bytes())

	// put request id  to batch
	if reqState := payload.RequestState(); reqState > 0 {
		rKey := s.getRequestIdKey(payload.RequestId())
		rVal := make([]byte, 1+len(txHash))
		rVal[0] = byte(reqState)
		copy(rVal[1:], txHash[:])
		s.store.BatchPut(rKey, rVal)
	}

	return nil
}

// GetTransaction return transaction by transaction hash
func (s *BlockStore) GetTransaction(txHash common.Uint256) (payload.Payload, uint64, error) {
	if s.enableCache {
		tx, height := s.cache.GetTransaction(txHash)
		if tx != nil {
			return tx, height, nil
		}
	}
	return s.loadTransaction(txHash)
}

// GetTransactionByReqId return transaction by request id
func (s *BlockStore) GetTransactionByReqId(reqId [32]byte) (payload.Payload, uint64, error) {
	if s.enableCache {
		tx, height := s.cache.GetTransactionByReqId(reqId)
		if tx != nil {
			return tx, height, nil
		}
	}
	return s.loadTransactionByReqId(reqId)
}

func (s *BlockStore) GetRequestState(reqId [32]byte) (payload.ReqState, error) {
	if s.enableCache {
		tx, _ := s.cache.GetTransactionByReqId(reqId)
		if tx != nil {
			return tx.RequestState(), nil
		}
	}
	return s.loadReqIdState(reqId)
}

func (s *BlockStore) loadReqIdState(reqId [32]byte) (payload.ReqState, error) {
	key := s.getRequestIdKey(reqId)
	value, err := s.store.Get(key)
	if err != nil {
		return 0, err
	}
	return payload.ReqState(value[0]), nil
}

func (s *BlockStore) loadTransactionByReqId(reqId [32]byte) (payload.Payload, uint64, error) {
	key := s.getRequestIdKey(reqId)
	value, err := s.store.Get(key)
	if err != nil {
		return nil, 0, err
	}
	var txHash common.Uint256
	copy(txHash[:], value[1:])
	return s.loadTransaction(txHash)
}

func (s *BlockStore) loadTransaction(txHash common.Uint256) (payload.Payload, uint64, error) {
	key, err := s.getTransactionKey(txHash)
	if err != nil {
		return nil, 0, err
	}
	var height uint64
	if s.enableCache {
		tx, height := s.cache.GetTransaction(txHash)
		if tx != nil {
			return tx, height, nil
		}
	}
	value, err := s.store.Get(key)
	if err != nil {
		return nil, 0, err
	}
	source := common.NewZeroCopySource(value)
	var eof bool
	height, eof = source.NextUint64()
	if eof {
		return nil, 0, io.ErrUnexpectedEOF
	}
	tx, err := types.TransactionDeserialization(source)
	if err != nil {
		return nil, 0, fmt.Errorf("transaction deserialize error %s", err)
	}
	return tx.Payload, height, nil
}

// ContainTransaction return whether the transaction is in store
func (s *BlockStore) ContainTransaction(txHash common.Uint256) (bool, error) {
	key, err := s.getTransactionKey(txHash)
	if err != nil {
		return false, err
	}
	if s.enableCache {
		if s.cache.ContainTransaction(txHash) {
			return true, nil
		}
	}
	_, err = s.store.Get(key)
	if err != nil {
		if err == scom.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetVersion return the version of store
func (s *BlockStore) GetVersion() (byte, error) {
	key := s.getVersionKey()
	value, err := s.store.Get(key)
	if err != nil {
		return 0, err
	}
	reader := bytes.NewReader(value)
	return reader.ReadByte()
}

// SaveVersion persist version to store
func (s *BlockStore) SaveVersion(ver byte) error {
	key := s.getVersionKey()
	return s.store.Put(key, []byte{ver})
}

// ClearAll clear all the data of block store
func (s *BlockStore) ClearAll() error {
	s.NewBatch()
	iter := s.store.NewIterator(nil)
	for iter.Next() {
		s.store.BatchDelete(iter.Key())
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return err
	}
	return s.CommitTo()
}

// CommitTo commit the batch to store
func (s *BlockStore) CommitTo() error {
	return s.store.BatchCommit()
}

// Close block store
func (s *BlockStore) Close() error {
	return s.store.Close()
}

func (s *BlockStore) getTransactionKey(txHash common.Uint256) ([]byte, error) {
	key := bytes.NewBuffer(nil)
	if err := key.WriteByte(byte(scom.DATA_TRANSACTION)); err != nil {
		return nil, err
	}
	if err := txHash.Serialize(key); err != nil {
		return nil, err
	}
	return key.Bytes(), nil
}

// getRequestIdKey return request id key
func (s *BlockStore) getRequestIdKey(requestId [32]byte) []byte {
	key := make([]byte, 33)
	key[0] = byte(scom.DATA_REQUEST_ID)
	copy(key[1:], requestId[:])
	return key
}

func (s *BlockStore) getHeaderKey(blockHash common.Uint256) []byte {
	data := blockHash.ToArray()
	key := make([]byte, 1+len(data))
	key[0] = byte(scom.DATA_HEADER)
	copy(key[1:], data)
	return key
}

func (s *BlockStore) getBlockHashKey(height uint64) []byte {
	key := make([]byte, 9, 9)
	key[0] = byte(scom.DATA_BLOCK)
	binary.LittleEndian.PutUint64(key[1:], height)
	return key
}

func (s *BlockStore) getCurrentBlockKey() []byte {
	return []byte{byte(scom.SYS_CURRENT_BLOCK)}
}

func (s *BlockStore) getBlockMerkleTreeKey() []byte {
	return []byte{byte(scom.SYS_BLOCK_MERKLE_TREE)}
}

func (s *BlockStore) getVersionKey() []byte {
	return []byte{byte(scom.SYS_VERSION)}
}

func (s *BlockStore) getHeaderIndexListKey(startHeight uint64) ([]byte, error) {
	key := bytes.NewBuffer(nil)
	if err := key.WriteByte(byte(scom.IX_HEADER_HASH_LIST)); err != nil {
		return nil, err
	}
	if err := serialization.WriteUint64(key, startHeight); err != nil {
		return nil, err
	}
	return key.Bytes(), nil
}

func (s *BlockStore) getStartHeightByHeaderIndexKey(key []byte) (uint64, error) {
	reader := bytes.NewReader(key[1:])
	height, err := serialization.ReadUint64(reader)
	if err != nil {
		return 0, err
	}
	return height, nil
}
