package node

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/janrockdev/darkblock/crypto"
	"github.com/janrockdev/darkblock/proto"
	"github.com/janrockdev/darkblock/services"
	"github.com/janrockdev/darkblock/types"
	"github.com/janrockdev/darkblock/util"
	pb "google.golang.org/protobuf/proto"
)

type HeaderList struct {
	headers []*proto.Header
}

func NewHeaderList() *HeaderList {
	return &HeaderList{
		headers: []*proto.Header{},
	}
}

func (list *HeaderList) Add(h *proto.Header) {
	list.headers = append(list.headers, h)
}

func (list *HeaderList) Get(index int) *proto.Header {

	if index == -1 || index > list.Height() {
		logger.Warn().Msgf("index out of range in HeaderList")
		return nil
	}
	return list.headers[index]
}

func (list *HeaderList) Height() int {
	return list.Len() - 1
}

func (list *HeaderList) Len() int {
	return len(list.headers)
}

// type UTXO struct {
// 	Hash     string
// 	OutIndex int
// 	Amount   int64
// 	Spend    bool
// }

type Chain struct {
	txStore    TXStorer
	blockStore BlockStorer
	// utxoStore  UTXOStorer
	headers *HeaderList
}

// func NewChain(bs BlockStorer, txStore TXStorer) *Chain {
// 	chain := &Chain{
// 		blockStore: bs,
// 		txStore:    txStore,
// 		//utxoStore:  NewMemoryUTXOStore(),
// 		headers: NewHeaderList(),
// 	}
// 	chain.addBlock(createGenesisBlock())
// 	return chain
// }

func NewChain(bs BlockStorer, txStore TXStorer) *Chain {
	db_dir := util.LoadConfig().BADGER.DataDir
	chain := &Chain{
		blockStore: bs,
		txStore:    txStore,
		//utxoStore:  NewMemoryUTXOStore(),
		headers: NewHeaderList(),
	}
	// check badger db for existing blocks
	// if there is no block, create a genesis block
	if !services.CacheExists(db_dir) {
		chain.addBlock(createGenesisBlock())
	} else {
		// connect to badger db
		bdb, err := services.ConnectBadgerDB(db_dir)
		if err != nil {
			util.Logger.Error().Msgf("error connecting to badger db: [%s]", err.Error())
			panic(err)
		}
		prevLastBlockBytes, _, _, err := bdb.GetLatestRecord() //prevLastPrefix
		if err != nil {
			util.Logger.Error().Msgf("error getting latest record from badger db: [%s]", err.Error())
			panic(err)
		}
		prevLastBlock := &proto.Block{}
		if err := pb.Unmarshal(prevLastBlockBytes, prevLastBlock); err != nil {
			util.Logger.Error().Msgf("error unmarshalling block: [%s]", err.Error())
			panic(err)
		}
		if err := chain.AddBlock(prevLastBlock); err != nil {
			util.Logger.Error().Msgf("error adding block to chain: [%s]", err.Error())
			panic(err)
		}
		bdb.Close()
	}

	return chain
}

func (c *Chain) Height() int {
	return c.headers.Height()
}

func (c *Chain) AddBlock(b *proto.Block) error {
	if err := c.ValidateBlock(b); err != nil {
		return err
	}
	return c.addBlock(b)
}

func (c *Chain) addBlock(b *proto.Block) error {

	if hex.EncodeToString(types.HashBlock(b))[:3] != "c95" {
		util.Logger.Debug().Msgf("(9) Adding block [%s] to local blockchain, height [%d], headers [%d]", hex.EncodeToString(types.HashBlock(b))[:3], c.Height(), b.Header.Height)
	} else {
		util.Logger.Debug().Msgf("adding genesis block to local blockchain")
	}

	c.headers.Add(b.Header)

	//for _, tx := range b.Transactions {
	///	if err := c.txStore.Put(tx); err != nil {
	//		return err
	//	}
	// hash := hex.EncodeToString(types.HashTransactions(tx))

	// // store UTXOs
	// for it, output := range tx.Outputs {
	// 	utxo := &UTXO{
	// 		Hash:     hash,
	// 		Amount:   output.Amount,
	// 		OutIndex: it,
	// 		Spend:    false,
	// 	}
	// 	if err := c.utxoStore.Put(utxo); err != nil {
	// 		return err
	// 	}
	// }

	// // double spend check
	// for _, input := range tx.Inputs {
	// 	key := fmt.Sprintf("%s_%d", hex.EncodeToString(input.PrevTxHash), input.PrevOutIndex)
	// 	utxo, err := c.utxoStore.Get(key)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	utxo.Spend = true
	// 	if err := c.utxoStore.Put(utxo); err != nil {
	// 		return err
	// 	}
	// }
	//}

	//util.Logger.Debug().Msgf("blockchain height: %s", c.headers.headers)

	return c.blockStore.Put(b)
}

func (c *Chain) GetBlockByHash(hash []byte) (*proto.Block, error) {
	hashHex := hex.EncodeToString(hash)
	return c.blockStore.Get(hashHex)
}

func (c *Chain) GetBlockByHeight(height int) (*proto.Block, error) {
	if c.Height() < height {
		return nil, fmt.Errorf("block height %d is greater than chain height %d", height, c.Height())
	}
	header := c.headers.Get(height)
	hash := types.HashHeader(header)
	return c.GetBlockByHash(hash)
}

func (c *Chain) ValidateBlock(b *proto.Block) error {
	// validate signature of the block
	if !types.VerifyBlock(b) {
		return fmt.Errorf("invalid block signature")
	}

	// currentBlock := &proto.Block{}
	// if c.Height() > 0 {
	// 	_, err := c.GetBlockByHeight(c.Height())
	// 	if err != nil {
	// 		util.Logger.Error().Msgf("error getting block by height: [%s]", err.Error())
	// 	}
	// } else {
	// 	bdb, err := services.ConnectBadgerDBReadOnly(db_dir)
	// 	if err != nil {
	// 		util.Logger.Error().Msgf("error connecting to badger db: [%s]", err.Error())
	// 		panic(err)
	// 	}
	// 	prevLastBlockBytes, _, _, err := bdb.GetLatestRecord() //prevLastPrefix
	// 	if err != nil {
	// 		util.Logger.Error().Msgf("error getting latest record from badger db: [%s]", err.Error())
	// 		panic(err)
	// 	}
	// 	currentBlock := &proto.Block{}
	// 	if err := pb.Unmarshal(prevLastBlockBytes, currentBlock); err != nil {
	// 		util.Logger.Error().Msgf("error unmarshalling block: [%s]", err.Error())
	// 		panic(err)
	// 	}
	// }

	if c.Height() > 0 {
		currentBlock, err := c.GetBlockByHeight(c.Height())
		if err != nil {
			util.Logger.Error().Msgf("error getting block by height: [%s]", err.Error())
		}

		hash := types.HashBlock(currentBlock)
		if !bytes.Equal(hash, b.Header.PrevHash) && hex.EncodeToString(hash)[:3] != "c95" {
			util.Logger.Error().Msgf("invalid previous block hash: [%s] expected: [%s]", hex.EncodeToString(hash)[:3], hex.EncodeToString(b.Header.PrevHash)[:3])
		}

		//validate transactions (double validation)
		for _, tx := range b.Transactions {
			if err := c.ValidateTransaction(tx); err != nil {
				util.Logger.Error().Msgf("validate transaction error: [%s]", err.Error())
				return err
			}
		}
	}

	return nil
}

func (c *Chain) ValidateTransaction(tx *proto.Transaction) error {
	// verify signature of the transaction
	if !types.VerifyTransaction(tx) {
		return fmt.Errorf("invalid transaction signature: [%s]", hex.EncodeToString(types.HashTransaction(tx))[:3])
	}

	// check if all the inputs are unspent
	// if tx.GetInputs() != nil {

	// 	var (
	// 		nInputs = len(tx.Inputs)
	// 		hash    = hex.EncodeToString(types.HashTransactions(tx))
	// 	)
	// 	sumInputs := 0
	// 	for i := 0; i < nInputs; i++ {
	// 		prevHash := hex.EncodeToString(tx.Inputs[i].PrevTxHash)
	// 		key := fmt.Sprintf("%s_%d", prevHash, i)
	// 		utxo, err := c.utxoStore.Get(key)
	// 		sumInputs += int(utxo.Amount)
	// 		if err != nil {
	// 			util.Logger.Error().Msgf("validate transaction error: [%s]", err.Error())
	// 			return err
	// 		}
	// 		if utxo.Spend {
	// 			util.Logger.Error().Msgf("input %d of tx %s is already spent", i, hash)
	// 			return fmt.Errorf("input %d of tx %s is already spent", i, hash) // <---- this need to be refactored
	// 		}
	// 	}

	// 	sumOutputs := 0
	// 	for _, output := range tx.Outputs {
	// 		sumOutputs += int(output.Amount)
	// 	}

	// 	if sumInputs < sumOutputs {
	// 		util.Logger.Error().Msgf("insufficient balance got (%d) speding (%d)", sumInputs, sumOutputs)
	// 		return fmt.Errorf("insufficient balance got (%d) speding (%d)", sumInputs, sumOutputs) // <---- this need to be refactored
	// 	}
	//}

	return nil
}

func createGenesisBlock() *proto.Block {
	privKey := crypto.NewPrivateKeyFromSeedStr(util.LoadConfig().KEYS.GodSeed)
	block := &proto.Block{
		Header: &proto.Header{
			Version: 1,
		},
	}

	tx := &proto.Transaction{
		Version: 1,
		Outputs: []*proto.TxOutput{
			{
				Amount:  1000,
				Address: privKey.Public().Address().Bytes(),
				Payload: []byte("genesis"),
			},
		},
	}

	block.Transactions = append(block.Transactions, tx)
	types.SignBlock(privKey, block)

	return block
}
