package services

import (
	"encoding/hex"
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/janrockdev/darkblock/proto"
	"github.com/janrockdev/darkblock/types"
	"github.com/stretchr/testify/assert"
)

func TestRecoveryFromLastBlock(t *testing.T) {
	DB, err := ConnectBadgerDB("../db")
	if err != nil {
		panic(err)
	}
	prevBlockHash, lastBlockHeight, lastTxHash, lastSignature, lastPublicKey, err := DB.GetRecoveryFromCache([]byte("blockStore"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("lastBlockHash: %s, lastBlockHeight: %d, lastTxHash: %s, lastSignature: %s, lastPublicKey: %s", hex.EncodeToString(prevBlockHash), lastBlockHeight, hex.EncodeToString(lastTxHash), hex.EncodeToString(lastSignature), hex.EncodeToString(lastPublicKey))
}

func previewBlock(lastBlock []byte) {
	// Unmarshal the serialized block into a Block struct
	var block *proto.Block
	block, _ = types.UnmarshalBlock(lastBlock)

	// Parse and log the Header
	header := block.Header
	log.Printf("Block Hash: %x", types.HashBlock(block))
	// log.Printf("Parsed Block Header:")
	// log.Printf("  Version: %d", header.Version)
	// log.Printf("  Height: %d", header.Height)
	log.Printf("  PrevHash: %x", header.PrevHash)
	// log.Printf("  RootHash: %x", header.RootHash)
	log.Printf("  Timestamp: %s", time.Unix(0, header.Timestamp))

	// // Parse and log the Public Key and Signature
	// log.Printf("Block Public Key: %x", block.PublicKey)
	// log.Printf("Block Signature: %x", block.Signature)

	// Iterate through Transactions
	log.Printf("Parsed Transactions:")
	for i, tx := range block.Transactions {
		log.Printf("Transaction %d:", i)
		log.Printf("  Version: %d", tx.Version)
		log.Printf("  Hash: %x", types.HashTransactionNoSigPuK(tx))

		// Log Transaction Inputs
		for j, input := range tx.Inputs {
			log.Printf("    Input %d:", j)
			log.Printf("      PrevTxHash: %s", input.PrevTxHash)
			log.Printf("      PrevOutIndex: %d", input.PrevOutIndex)
			log.Printf("      PublicKey: %x", input.PublicKey)
			log.Printf("      Signature: %x", input.Signature)
		}

		// Log Transaction Outputs
		for k, output := range tx.Outputs {
			log.Printf("    Output %d:", k)
			log.Printf("      Amount: %d", output.Amount)
			log.Printf("      Address: %x", output.Address)
			log.Printf("      Payload: %s", string(output.Payload)) // Assuming payload is string-like
		}
	}
}

func TestGetLatestTwoBlocks(t *testing.T) {
	// Connect to the database
	DB, err := ConnectBadgerDB("../db")
	assert.Nil(t, err)

	// Get the length of the blockStore namespace
	length, err := DB.Len([]byte("blockStore"))
	assert.Nil(t, err)
	fmt.Println("Length of blockStore namespace:", length)

	// Retrieve the last block by its number
	prevLastBlock, _, _, err := DB.GetLatestRecord() //GetByNumber([]byte("blockStore"), 0)
	assert.Nil(t, err)
	previewBlock(prevLastBlock)

	DB.Close()

	// // Retrieve the last block by its number
	// prevPrevLastBlock, err := DB.GetByNumber([]byte("blockStore"), length-2)
	// assert.Nil(t, err)
	// previewBlock(prevPrevLastBlock)

	// // Retrieve the last block by its number
	// prevPrevLastBlock2, err := DB.GetByNumber([]byte("blockStore"), length-3)
	// assert.Nil(t, err)
	// previewBlock(prevPrevLastBlock2)
}

func TestPrintAllTransactions(t *testing.T) {
	// Open the database in read-only mode
	opts := badger.DefaultOptions("../db").WithReadOnly(true).WithLogger(nil)
	bdb, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer bdb.Close()

	// Start a read-only transaction
	var prefix string
	err = bdb.View(func(txn *badger.Txn) error {
		// Create an iterator for reading through all keys
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		// Iterate through all keys in the database
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()

			//Retrieve the value for each key
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			// Print the key and value
			//fmt.Printf("Key: %s, Value: %s\n", key, val)
			// Extract the prefix using the regex
			prefixRegex := regexp.MustCompile(`^blockStore/(\d+)_`) // regex to match the prefix before the first underscore
			matches := prefixRegex.FindStringSubmatch(string(key))
			if len(matches) > 1 {
				// Set the prefix as the first captured match
				prefix = matches[1]
			}
			fmt.Printf("Key: %s Prefix: %s\n", key, prefix)
			previewBlock(val)
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}
