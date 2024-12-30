package main

import (
	"flag"

	"github.com/janrockdev/darkblock/crypto"
	"github.com/janrockdev/darkblock/node"
	"github.com/janrockdev/darkblock/util"
)

var logger = util.Logger

func main() {
	port := flag.String("port", ":3000", "port to run the node on")
	flag.Parse()
	if *port == "" {
		logger.Fatal().Msg("port is required")
	}
	if *port == ":3000" {
		logger.Info().Msg("starting bootstrap & validator node on port [:3000]")
		makeNode(*port, []string{}, true)
	} else {
		logger.Info().Msg("starting discovery, contacting bootstrap & validator node on port [:3000]")
		makeNode(*port, []string{":3000"}, false)
	}

	select {} // block main thread forever
}

// makeNode creates a new node with the given listen address and bootstrap nodes
func makeNode(listenAddr string, bootstrapNodes []string, isValidator bool) *node.Node {
	cfg := node.ServerConfig{
		Version:    "darkblock-1",
		ListenAddr: listenAddr,
	}
	if isValidator {
		privKey, err := crypto.LoadPrivateKeyFromFile("private_key.txt")
		if err != nil {
			logger.Fatal().Msgf("failed to load private key: %s", err)
		}
		cfg.PrivateKey = privKey
	}
	n := node.NewNode(cfg, bootstrapNodes) // again bootstrapNodes for consensus
	go n.Start(listenAddr, bootstrapNodes)

	return n
}

// makeTransaction creates a new transaction with random inputs and outputs,
// signs it with a generated private key, and sends it to the node running at :3000
// func makeTransaction() {
// 	// create a context with timeout
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	// create a new grpc client
// 	client, err := grpc.DialContext(ctx, ":3000", grpc.WithTransportCredentials(insecure.NewCredentials()))
// 	if err != nil {
// 		logger.Fatal().Msgf("did not connect to %s: %v", ":3000", err)
// 	}
// 	defer client.Close()

// 	// create a new node client
// 	c := proto.NewNodeClient(client)
// 	privKey := crypto.GeneratePrivateKey()
// 	tx := &proto.Transaction{
// 		Version: 1,
// 		Inputs: []*proto.TxInput{
// 			{
// 				PrevTxHash:   util.RandomHash(),
// 				PrevOutIndex: 0,
// 				PublicKey:    privKey.Public().Bytes(),
// 			},
// 		},
// 		Outputs: []*proto.TxOutput{
// 			{
// 				Amount:  99,
// 				Address: privKey.Public().Address().Bytes(),
// 			},
// 		},
// 	}

// 	// sign the transaction
// 	_, err = c.HandleTransaction(ctx, tx)
// 	if err != nil {
// 		logger.Fatal().Msgf("handshake failed while sending transaction to node at :3000: %s", err)
// 	}
// }
