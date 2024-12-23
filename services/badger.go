package services

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/janrockdev/darkblock/proto"
	"github.com/janrockdev/darkblock/types"
	"github.com/janrockdev/darkblock/util"
)

const (
	badgerDiscardRatio = 0.5
	badgerGCInterval   = 5 * time.Second
)

type (
	DB interface {
		Get(namespace, key []byte) (value []byte, err error)
		GetByNumber(namespace []byte, number int64) (value []byte, err error)
		GetLatestRecord() (value []byte, prefix int64, hash []byte, err error)
		GetRecoveryFromCache(nameSpace []byte) (lastBlockHash []byte, lastBlockHeight int32, lastTxHash []byte, lastSignature []byte, lastPublicKey []byte, err error)
		Set(namespace, keyHash []byte, keyHeight int64, value []byte) error
		Has(namespace, key []byte) (bool, error)
		Size(namespace []byte) (int64, error)
		Len(namespace []byte) (int64, error)
		RecordExists() (bool, error)
		Close() error
	}

	BadgerDB struct {
		db         *badger.DB
		ctx        context.Context
		cancelFunc context.CancelFunc
		logger     badger.Logger
	}
)

func NewBadgerDB(dataDir string) DB {
	if err := os.MkdirAll(dataDir, 0774); err != nil {
		return nil
	}

	opts := badger.DefaultOptions("")
	opts.SyncWrites = true
	opts.Dir, opts.ValueDir = dataDir, dataDir

	opts.NumVersionsToKeep = 0
	opts.CompactL0OnClose = true
	opts.NumLevelZeroTables = 1
	opts.NumLevelZeroTablesStall = 2
	opts.ValueLogFileSize = 1024 * 1024 * 10

	badgerDB, err := badger.Open(opts)
	if err != nil {
		return nil
	}
	//defer badgerDB.Close()

	bdb := &BadgerDB{
		db: badgerDB,
	}
	bdb.ctx, bdb.cancelFunc = context.WithCancel(context.Background())

	go bdb.runGC()
	return bdb
}

func ConnectBadgerDB(dataDir string) (DB, error) {
	opts := badger.DefaultOptions("")
	opts.SyncWrites = true
	opts.WithLogger(nil)
	opts.Logger = nil
	opts.Dir, opts.ValueDir = dataDir, dataDir

	badgerDB, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	//defer badgerDB.Close()

	bdb := &BadgerDB{
		db: badgerDB,
	}
	bdb.ctx, bdb.cancelFunc = context.WithCancel(context.Background())

	//go bdb.runGC()
	return bdb, nil
}

func ConnectBadgerDBReadOnly(dataDir string) (DB, error) {
	opts := badger.DefaultOptions(dataDir).
		WithReadOnly(true).WithLogger(nil)

	badgerDB, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	//defer badgerDB.Close()

	bdb := &BadgerDB{
		db: badgerDB,
	}
	bdb.ctx, bdb.cancelFunc = context.WithCancel(context.Background())

	//go bdb.runGC()
	return bdb, nil
}

func FolderExists(dataDir string) bool {
	info, err := os.Stat(dataDir)
	if err != nil {
		return false
	}

	return info.IsDir()
}

func CacheExists(dataDir string) bool {
	opts := badger.DefaultOptions(dataDir).
		WithReadOnly(true).WithLogger(nil)

	bdb, err := badger.Open(opts)
	if err != nil {
		return false
	}
	defer bdb.Close()

	err = bdb.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			return nil
		}
		return badger.ErrKeyNotFound
	})

	return err == nil
}

func (bdb *BadgerDB) RecordExists() (bool, error) {
	err := bdb.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			return nil
		}
		return badger.ErrKeyNotFound
	})
	if err == badger.ErrKeyNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

func (bdb *BadgerDB) Get(namespace, key []byte) (value []byte, err error) {
	err = bdb.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(badgerNamespaceKey(namespace, key))
		if err != nil {
			return err
		}

		value, err = item.ValueCopy(value)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return value, nil
}

func (bdb *BadgerDB) Set(namespace, keyHash []byte, keyHeight int64, value []byte) error {
	err := bdb.db.Update(func(txn *badger.Txn) error {
		return txn.Set(badgerNamespaceKey(namespace, []byte(fmt.Sprintf("%d_%s", keyHeight, keyHash))), value)
	})

	if err != nil {
		bdb.logger.Debugf("failed to set key %s for namespace %s: %v", fmt.Sprintf("%d_%s", keyHeight, keyHash), namespace, err)
		return err
	}

	return nil
}

func (bdb *BadgerDB) Has(namespace, key []byte) (ok bool, err error) {
	_, err = bdb.Get(namespace, key)
	switch err {
	case badger.ErrKeyNotFound:
		ok, err = false, nil
	case nil:
		ok, err = true, nil
	}

	return
}

// get record by number
func (bdb *BadgerDB) GetByNumber(namespace []byte, index int64) (value []byte, err error) {
	err = bdb.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false // Only keys are prefetched
		it := txn.NewIterator(opts)
		defer it.Close()
		currentIndex := int64(0) // Counter to track the index during iteration
		for it.Seek(namespace); it.ValidForPrefix(namespace); it.Next() {
			if currentIndex == index {
				// Key found at the desired index, retrieve its value
				item := it.Item()
				value, err = item.ValueCopy(nil)
				return err
			}
			currentIndex++
		}

		return fmt.Errorf("index %d not found in namespace %s", index, namespace)
	})

	return value, err
}

func (bdb *BadgerDB) GetLatestRecord() (value []byte, prefix int64, hash []byte, err error) {

	err = bdb.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Reverse = true
		it := txn.NewIterator(opts)
		defer it.Close()

		// Move to the end of the key range.
		it.Rewind()

		if !it.Valid() {
			return badger.ErrKeyNotFound
		}

		item := it.Item()
		key := item.Key()
		var err error
		value, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}

		keyStr := string(key)

		hashRegex := `^blockStore/\d+_(\w+)$`
		re := regexp.MustCompile(hashRegex)
		matches := re.FindStringSubmatch(keyStr)
		if len(matches) > 1 {
			if len(matches[1]) > 0 {
				hashMatch := matches[1]
				hash, err = hex.DecodeString(hashMatch)
				if err != nil {
					return err
				}
			}
		}

		prefixRegex := regexp.MustCompile(`^blockStore/(\d+)_`)
		matches = prefixRegex.FindStringSubmatch(string(key))
		if len(matches) > 1 {
			prefix, err = strconv.ParseInt(matches[1], 10, 64)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return value, prefix, hash, err
}

func (bdb *BadgerDB) Size(namespace []byte) (int64, error) {
	var size int64
	err := bdb.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false // We only need keys for size calculation
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(namespace); it.ValidForPrefix(namespace); it.Next() {
			item := it.Item()

			// Add the size of the key
			size += int64(len(item.Key()))

			// Add the size of the value
			valSize := item.EstimatedSize()
			size += int64(valSize)
		}
		return nil
	})

	return size, err
}

func (bdb *BadgerDB) Len(namespace []byte) (int64, error) {
	var keyCount int64
	err := bdb.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(namespace); it.ValidForPrefix(namespace); it.Next() {
			// Increment the key count
			keyCount++
		}
		return nil
	})

	return keyCount, err
}

// Get the last block hash, height, and last transaction hash from the cache
func (bdb *BadgerDB) GetRecoveryFromCache(nameSpace []byte) (lastBlockHash []byte, lastBlockHeight int32, lastTxHash []byte, lastSignature []byte, lastPublicKey []byte, err error) {
	length, err := bdb.Len(nameSpace)
	if err != nil {
		util.Logger.Error().Msgf("error getting length of namespace %s: %v", nameSpace, err)
		panic(err)
	}

	lastBlockBytes, err := bdb.GetByNumber(nameSpace, length-1)
	if err != nil {
		util.Logger.Error().Msgf("error getting last block from namespace %s: %v", nameSpace, err)
		panic(err)
	}

	var lastBlock *proto.Block
	lastBlock, err = types.UnmarshalBlock(lastBlockBytes)
	if err != nil {
		util.Logger.Error().Msgf("error unmarshalling last block: %v", err)
		panic(err)
	}

	last := lastBlock.Transactions[len(lastBlock.Transactions)-1]
	lastTxHash = types.HashTransactionNoSigPuK(last)
	lastSignature = last.Inputs[0].Signature
	lastPublicKey = last.Inputs[0].PublicKey

	return types.HashBlock(lastBlock), lastBlock.Header.Height, lastTxHash, lastSignature, lastPublicKey, nil
}

func (bdb *BadgerDB) Close() error {
	bdb.cancelFunc()
	return bdb.db.Close()
}

func (bdb *BadgerDB) runGC() {
	ticker := time.NewTicker(badgerGCInterval)
	for {
		select {
		case <-ticker.C:
			err := bdb.db.RunValueLogGC(badgerDiscardRatio)
			if err != nil {
				// don't report error when GC didn't result in any cleanup
				if err == badger.ErrNoRewrite {
					util.Logger.Trace().Msgf("no badgerDB GC occurred: %v", err)
				} else {
					util.Logger.Error().Msgf("failed to GC BadgerDB: %v", err)
				}
			}

		case <-bdb.ctx.Done():
			return
		}
	}
}

func badgerNamespaceKey(namespace, key []byte) []byte {
	prefix := []byte(fmt.Sprintf("%s/", namespace))
	return append(prefix, key...)
}
