package main

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/janrockdev/darkblock/crypto"
	"github.com/janrockdev/darkblock/proto"
	"github.com/janrockdev/darkblock/services"
	"github.com/janrockdev/darkblock/types"
	"github.com/janrockdev/darkblock/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var logger = util.Logger

func main() {
	// port := flag.String("port", ":4000", "port to connect to the node")
	flag.Parse()
	metadata := "hello2"

	// // insert
	// start := time.Now()
	// var i int
	// for i = 1; i < 2; i++ {
	// 	sendTransaction(i, *port, metadata)
	// 	//time.Sleep(1 * time.Second)
	// }
	// end := time.Now()
	// logger.Info().Msgf("time taken to send %d transactions: %s", i-1, end.Sub(start))

	// time.Sleep(2 * time.Second)

	// search
	metadataObject := fmt.Sprintf("{\"metadata\": \"sims_%s\"}", metadata)
	str := base64.StdEncoding.EncodeToString([]byte(metadataObject))
	cs, err := services.NewCouchbaseService("couchbase://localhost", "Administrator", "password", "blocks", "transactions")
	if err != nil {
		logger.Error().Msgf("failed to create Couchbase service: %v", err)
	}
	transaction := ""
	transaction, err = cs.SearchTransactionByPayload([]byte(str))
	if err != nil {
		logger.Error().Msgf("failed to search transaction: %v", err)
	}
	util.Logger.Info().Msg(transaction)
	block := ""
	block, err = cs.SearchBlockByPayload([]byte(str))
	if err != nil {
		logger.Error().Msgf("failed to search block: %v", err)
	}
	util.Logger.Info().Msg(block)
	cs.Close()
	if searchBlockAndValidate(strings.TrimPrefix(block, "block::"), metadataObject) {
		logger.Info().Msg("matadata validated")
	} else {
		logger.Error().Msg("matadata not found")
	}
}

// first block
func searchTransaction(index int32) []*proto.Transaction {
	// create a context with timeout,
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// create a new grpc client
	client, err := grpc.DialContext(ctx, ":3000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal().Msgf("did not connect to %s: %v", ":3000", err)
	}
	defer client.Close()

	// create a new node client
	c := proto.NewNodeClient(client)

	// search block
	block, err := c.GetBlock(ctx, &proto.BlockSearch{BlockHeight: index})
	if err != nil {
		logger.Fatal().Msgf("unable to find block index [%s]", err)
	}

	return block.Block.Transactions
}

func searchBlockAndValidate(key string, metadata string) bool {
	db_dir := util.LoadConfig().BADGER.DataDir
	DB, err := services.ConnectBadgerDBReadOnly(db_dir)
	if err != nil {
		logger.Error().Msgf("failed to create Couchbase service: %v", err)
	}
	length, err := DB.Len([]byte("blockStore"))
	if err != nil {
		logger.Error().Msgf("failed to open Couchbase bucket: %v", err)
	}
	util.Logger.Debug().Msgf("length of blockStore namespace: %d", length)
	block, err := DB.Get([]byte("blockStore"), []byte(key))
	if err != nil {
		logger.Error().Msgf("failed to validate metadata: %v", err)
	}
	res := services.ValidatePayload(block, metadata)
	DB.Close()

	return res
}

func sendTransaction(i int, port string, v string) {
	// create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// create a new grpc client
	client, err := grpc.DialContext(ctx, port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal().Msgf("did not connect to %s: %v", port, err)
	}
	defer client.Close()

	// create a new node client
	c := proto.NewNodeClient(client)
	privKey, err := crypto.LoadPrivateKeyFromFile("private_key.txt")
	if err != nil {
		logger.Fatal().Msgf("failed to load private key: %s", err)
	}

	if v == "" {
		v = uuid.New().String()
	}

	tx := &proto.Transaction{
		Version:   1,
		Timestamp: time.Now().UnixNano(),
		Inputs: []*proto.TxInput{
			{
				PrevTxHash:   nil, // this will be filled in later after tx update
				PrevOutIndex: uint32(i),
				Signature:    nil, // this will be filled in later after signing
				PublicKey:    nil, // this will be filled in later after signing
			},
		},
		Outputs: []*proto.TxOutput{
			{
				Amount:  1,
				Address: privKey.Public().Address().Bytes(),
				Payload: []byte("{\"metadata\": \"sims_" + v + "\"}"),
			},
		},
	}

	// hash the transaction
	hashTx := types.HashTransaction(tx)

	// sign the transaction
	sig := crypto.SignTransaction(privKey, hashTx)
	tx.Inputs[0].Signature = sig.Bytes()

	// add the signature and public key to the transaction
	tx.Inputs[0].Signature = sig.Bytes()
	tx.Inputs[0].PublicKey = privKey.Public().Bytes()

	// send the transaction
	_, err = c.HandleTransaction(ctx, tx) // receiver side will have different hash bacause added signature and pubkey
	if err != nil {
		logger.Fatal().Msgf("handshake failed while sending transaction to node at %s: %s", port, err)
	}

	// log what I sent <---- this need to be refactored
	var (
		prevOutIndex = tx.Inputs[0].PrevOutIndex
		signature    = tx.Inputs[0].Signature
		pubKey       = tx.Inputs[0].PublicKey
	)

	red := "\x1b[32m"
	reset := "\x1b[0m"
	logger.Debug().Msgf("sent transaction [%s%s%s], version [%d], prevOutIndex [%d], signature [%s], publicKey[%s]",
		red, hex.EncodeToString(hashTx)[:3], reset, tx.Version, prevOutIndex, hex.EncodeToString(signature)[:3], hex.EncodeToString(pubKey)[:3])
}
