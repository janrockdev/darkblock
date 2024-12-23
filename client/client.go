package main

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/janrockdev/darkblock/crypto"
	"github.com/janrockdev/darkblock/proto"
	"github.com/janrockdev/darkblock/types"
	"github.com/janrockdev/darkblock/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var logger = util.Logger

func main() {
	start := time.Now()
	for i := 1; i < 10001; i++ {

		sendTransaction(i)
		//time.Sleep(10 * time.Millisecond)
	}
	end := time.Now()
	logger.Info().Msgf("time taken to send 100 transactions: %s", end.Sub(start))
	//time.Sleep(5 * time.Second)
	//util.Logger.Debug().Msgf("%s", searchTransaction(1))
}

// first just block
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

func sendTransaction(i int) {
	// create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// create a new grpc client
	client, err := grpc.DialContext(ctx, ":4000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal().Msgf("did not connect to %s: %v", ":3000", err)
	}
	defer client.Close()

	// create a new node client
	c := proto.NewNodeClient(client)
	privKey, err := crypto.LoadPrivateKeyFromFile("private_key.txt")
	if err != nil {
		logger.Fatal().Msgf("failed to load private key: %s", err)
	}

	rno := uuid.New().String()

	tx := &proto.Transaction{
		Version: 1,
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
				Payload: []byte("{\"metadata\": \"sims_" + rno + "\"}"),
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
		logger.Fatal().Msgf("handshake failed while sending transaction to node at :3000: %s", err)
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
