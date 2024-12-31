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
		privKey, err := crypto.LoadPrivateKeyFromFile("private_key.txt") // load private key for node from file
		if err != nil {
			logger.Fatal().Msgf("failed to load private key: %s", err)
		}
		cfg.PrivateKey = privKey
	}
	n := node.NewNode(cfg, bootstrapNodes) // bootstrapNodes for consensus
	go n.Start(listenAddr, bootstrapNodes)

	return n
}
