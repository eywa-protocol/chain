package leveldbstore

import (
	"github.com/ethereum/go-ethereum/common/fdlimit"
	"github.com/eywa-protocol/chain/core/store/common"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// LevelDBStore level DB store
type LevelDBStore struct {
	db    *leveldb.DB // LevelDB instance
	batch *leveldb.Batch
}

// BITSPERKEY used to compute the size of bloom filter bits array .
// too small will lead to high false positive rate.
const BITSPERKEY = 10

// NewLevelDBStore return LevelDBStore instance
func NewLevelDBStore(file string) (*LevelDBStore, error) {
	openFileCache := opt.DefaultOpenFilesCacheCapacity
	maxOpenFiles, err := fdlimit.Current()
	if err == nil && maxOpenFiles < openFileCache*5 {
		openFileCache = maxOpenFiles / 5
	}

	if openFileCache < 16 {
		openFileCache = 16
	}

	// default Options
	o := opt.Options{
		NoSync:                 false,
		OpenFilesCacheCapacity: openFileCache,
		Filter:                 filter.NewBloomFilter(BITSPERKEY),
	}

	db, err := leveldb.OpenFile(file, &o)

	if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
		db, err = leveldb.RecoverFile(file, nil)
	}

	if err != nil {
		return nil, err
	}

	return &LevelDBStore{
		db:    db,
		batch: nil,
	}, nil
}

func NewMemLevelDBStore() (*LevelDBStore, error) {
	store := storage.NewMemStorage()
	// default Options
	o := opt.Options{
		NoSync: false,
		Filter: filter.NewBloomFilter(BITSPERKEY),
	}
	db, err := leveldb.Open(store, &o)
	if err != nil {
		return nil, err
	}

	return &LevelDBStore{
		db:    db,
		batch: nil,
	}, nil
}

// Put a key-value pair to leveldb
func (s *LevelDBStore) Put(key []byte, value []byte) error {
	return s.db.Put(key, value, nil)
}

// Get the value of a key from leveldb
func (s *LevelDBStore) Get(key []byte) ([]byte, error) {
	dat, err := s.db.Get(key, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	return dat, nil
}

// Has return whether the key is exist in leveldb
func (s *LevelDBStore) Has(key []byte) (bool, error) {
	return s.db.Has(key, nil)
}

// Delete the the in leveldb
func (s *LevelDBStore) Delete(key []byte) error {
	return s.db.Delete(key, nil)
}

// NewBatch start commit batch
func (s *LevelDBStore) NewBatch() {
	s.batch = new(leveldb.Batch)
}

// BatchPut put a key-value pair to leveldb batch
func (s *LevelDBStore) BatchPut(key []byte, value []byte) {
	s.batch.Put(key, value)
}

// BatchDelete delete a key to leveldb batch
func (s *LevelDBStore) BatchDelete(key []byte) {
	s.batch.Delete(key)
}

// BatchCommit commit batch to leveldb
func (s *LevelDBStore) BatchCommit() error {
	err := s.db.Write(s.batch, nil)
	if err != nil {
		return err
	}
	s.batch = nil
	return nil
}

// Close leveldb
func (s *LevelDBStore) Close() error {
	err := s.db.Close()
	return err
}

// NewIterator return a iterator of leveldb with the key prefix
func (s *LevelDBStore) NewIterator(prefix []byte) common.StoreIterator {

	iter := s.db.NewIterator(util.BytesPrefix(prefix), nil)

	return iter
}
