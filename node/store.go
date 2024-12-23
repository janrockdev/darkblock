package node

import (
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/janrockdev/darkblock/proto"
	"github.com/janrockdev/darkblock/types"
)

// UTXO storer interface.
// type UTXOStorer interface {
// 	Put(*UTXO) error
// 	Get(string) (*UTXO, error)
// }

// UIXO struct.
// type MemoryUTXOStore struct {
// 	lock sync.RWMutex
// 	data map[string]*UTXO
// }

// NewMemoryUTXOStore creates a new in-memory UTXO store.
// func NewMemoryUTXOStore() *MemoryUTXOStore {
// 	return &MemoryUTXOStore{
// 		data: make(map[string]*UTXO),
// 	}
// }

// Get retrieves a UTXO from the store.
// func (s *MemoryUTXOStore) Get(hash string) (*UTXO, error) {
// 	s.lock.RLock()
// 	defer s.lock.RUnlock()

// 	// Check if the UTXO exists.
// 	utxo, ok := s.data[hash]
// 	if !ok {
// 		return nil, fmt.Errorf("utxo [%s] not found", hash)
// 	}

// 	return utxo, nil
// }

// Put stores a UTXO in the store.
// func (s *MemoryUTXOStore) Put(utxo *UTXO) error {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	key := fmt.Sprintf("%s_%d", utxo.Hash, utxo.OutIndex)
// 	s.data[key] = utxo

// 	return nil
// }

// TX storer interface.
type TXStorer interface {
	Put(*proto.Transaction) error
	Get(string) (*proto.Transaction, error)
	Size() int
	GetAll() []*proto.Transaction
	Clear() error
}

// TX struct.
type MemoryTXStore struct {
	lock sync.RWMutex
	txx  map[string]*proto.Transaction
}

// NewMemoryTXStore creates a new in-memory TX store.
func NewMemoryTXStore() *MemoryTXStore {
	return &MemoryTXStore{
		txx: make(map[string]*proto.Transaction),
	}
}

// Get retrieves a TX from the store.
func (s *MemoryTXStore) Get(hash string) (*proto.Transaction, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	tx, ok := s.txx[hash]
	if !ok {
		return nil, fmt.Errorf("tx [%s] not found", hash)
	}

	return tx, nil
}

// Put stores a TX in the store.
func (s *MemoryTXStore) Put(tx *proto.Transaction) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	hash := hex.EncodeToString(types.HashTransaction(tx))
	s.txx[hash] = tx

	return nil
}

// Size returns the number of transactions in the store.
func (s *MemoryTXStore) Size() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return len(s.txx)
}

// GetAll returns all transactions in the store.
func (s *MemoryTXStore) GetAll() []*proto.Transaction {
	s.lock.RLock()
	defer s.lock.RUnlock()

	txs := make([]*proto.Transaction, 0, len(s.txx))
	for _, tx := range s.txx {
		txs = append(txs, tx)
	}

	return txs
}

// Clear removes all transactions from the store.
func (s *MemoryTXStore) Clear() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.txx = make(map[string]*proto.Transaction)

	return nil
}

// Block storer interface.
type BlockStorer interface {
	Put(*proto.Block) error
	Get(string) (*proto.Block, error)
	Size() int
}

// Block struct.
type MemoryBlockStore struct {
	lock   sync.RWMutex
	blocks map[string]*proto.Block
}

// NewMemoryBlockStore creates a new in-memory block store.
func NewMemoryBlockStore() *MemoryBlockStore {
	return &MemoryBlockStore{blocks: make(map[string]*proto.Block)}
}

// Put stores a block in the store.
func (s *MemoryBlockStore) Put(b *proto.Block) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	hash := hex.EncodeToString(types.HashBlock(b))
	s.blocks[hash] = b

	return nil
}

// Get retrieves a block from the store.
func (s *MemoryBlockStore) Get(hash string) (*proto.Block, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	block, ok := s.blocks[hash]
	if !ok {
		return nil, fmt.Errorf("block [%s] not found", hash)
	}

	return block, nil
}

// Size returns the number of blocks in the store.
func (s *MemoryBlockStore) Size() int {

	s.lock.Lock()
	defer s.lock.Unlock()

	return len(s.blocks)
}
