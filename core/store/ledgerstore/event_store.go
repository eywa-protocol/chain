package ledgerstore

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/common/serialization"
	scom "github.com/eywa-protocol/chain/core/store/common"
	"github.com/eywa-protocol/chain/core/store/leveldbstore"
	"github.com/eywa-protocol/chain/native/event"
	"github.com/sirupsen/logrus"
)

// EventStore saving event notifies gen by smart contract execution
type EventStore struct {
	dbDir string                     // Store path
	store *leveldbstore.LevelDBStore // Store handler
}

// NewEventStore return event store instance
func NewEventStore(dbDir string) (*EventStore, error) {
	store, err := leveldbstore.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	return &EventStore{
		dbDir: dbDir,
		store: store,
	}, nil
}

// NewBatch start event commit batch
func (s *EventStore) NewBatch() {
	s.store.NewBatch()
}

// SaveEventNotifyByTx persist event notify by transaction hash
func (s *EventStore) SaveEventNotifyByTx(txHash common.Uint256, notify *event.ExecuteNotify) error {
	result, err := json.Marshal(notify)
	if err != nil {
		return fmt.Errorf("json.Marshal error %s", err)
	}
	key := s.getEventNotifyByTxKey(txHash)
	s.store.BatchPut(key, result)
	return nil
}

// SaveEventNotifyByBlock persist transaction hash which have event notify to store
func (s *EventStore) SaveEventNotifyByBlock(height uint64, txHashs []common.Uint256) error {
	key, err := s.getEventNotifyByBlockKey(height)
	if err != nil {
		return err
	}

	values := bytes.NewBuffer(nil)
	err = serialization.WriteUint32(values, uint32(len(txHashs)))
	if err != nil {
		return err
	}
	for _, txHash := range txHashs {
		err = txHash.Serialize(values)
		if err != nil {
			return err
		}
	}
	s.store.BatchPut(key, values.Bytes())

	return nil
}

// GetEventNotifyByTx return event notify by trasanction hash
func (s *EventStore) GetEventNotifyByTx(txHash common.Uint256) (*event.ExecuteNotify, error) {
	key := s.getEventNotifyByTxKey(txHash)
	data, err := s.store.Get(key)
	if err != nil {
		return nil, err
	}
	var notify event.ExecuteNotify
	if err = json.Unmarshal(data, &notify); err != nil {
		return nil, fmt.Errorf("json.Unmarshal error %s", err)
	}
	return &notify, nil
}

// GetEventNotifyByBlock return all event notify of transaction in block
func (s *EventStore) GetEventNotifyByBlock(height uint64) ([]*event.ExecuteNotify, error) {
	key, err := s.getEventNotifyByBlockKey(height)
	if err != nil {
		return nil, err
	}
	data, err := s.store.Get(key)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewBuffer(data)
	size, err := serialization.ReadUint64(reader)
	if err != nil {
		return nil, fmt.Errorf("ReadUint32 error %s", err)
	}
	evtNotifies := make([]*event.ExecuteNotify, 0)
	for i := uint64(0); i < size; i++ {
		var txHash common.Uint256
		err = txHash.Deserialize(reader)
		if err != nil {
			return nil, fmt.Errorf("txHash.Deserialize error %s", err)
		}
		evtNotify, err := s.GetEventNotifyByTx(txHash)
		if err != nil {
			logrus.Errorf("getEventNotifyByTx Height:%d by txhash:%s error:%s", height, txHash.ToHexString(), err)
			continue
		}
		evtNotifies = append(evtNotifies, evtNotify)
	}
	return evtNotifies, nil
}

// CommitTo event store batch to store
func (s *EventStore) CommitTo() error {
	return s.store.BatchCommit()
}

// Close event store
func (s *EventStore) Close() error {
	return s.store.Close()
}

// ClearAll all data in event store
func (s *EventStore) ClearAll() error {
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

// SaveCurrentBlock persist current block height and block hash to event store
func (s *EventStore) SaveCurrentBlock(height uint64, blockHash common.Uint256) error {
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

// GetCurrentBlock return current block hash, and block height
func (s *EventStore) GetCurrentBlock() (common.Uint256, uint32, error) {
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
	height, err := serialization.ReadUint32(reader)
	if err != nil {
		return common.Uint256{}, 0, err
	}
	return blockHash, height, nil
}

func (s *EventStore) getCurrentBlockKey() []byte {
	return []byte{byte(scom.SYS_CURRENT_BLOCK)}
}

func (s *EventStore) getEventNotifyByBlockKey(height uint64) ([]byte, error) {
	key := make([]byte, 9, 9)
	key[0] = byte(scom.EVENT_NOTIFY)
	binary.LittleEndian.PutUint64(key[1:], height)
	return key, nil
}

func (s *EventStore) getEventNotifyByTxKey(txHash common.Uint256) []byte {
	data := txHash.ToArray()
	key := make([]byte, 1+len(data))
	key[0] = byte(scom.EVENT_NOTIFY)
	copy(key[1:], data)
	return key
}
