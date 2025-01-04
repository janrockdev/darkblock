package node

import (
	"context"
	"encoding/hex"
	"net"
	"sync"
	"time"

	"github.com/janrockdev/darkblock/consensus"
	"github.com/janrockdev/darkblock/crypto"
	"github.com/janrockdev/darkblock/proto"
	"github.com/janrockdev/darkblock/services"
	"github.com/janrockdev/darkblock/types"
	"github.com/janrockdev/darkblock/util"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
)

var (
	logger            = util.Logger
	blockTime         = time.Second * time.Duration(util.LoadConfig().NETWORK.Tick)
	globalDialedAddrs = make(map[string]string)
	globalDialedLock  sync.Mutex
	red               = "\x1b[32m"
	reset             = "\x1b[0m"
)

// Mempool struct.
type Mempool struct {
	lock sync.RWMutex
	txx  map[string]*proto.Transaction
}

// NewMempool creates a new mempool.
func NewMempool() *Mempool {
	return &Mempool{
		txx: make(map[string]*proto.Transaction),
	}
}

// Clear clears the mempool.
func (pool *Mempool) Clear() []*proto.Transaction {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	txx := make([]*proto.Transaction, len(pool.txx))
	it := 0
	for k, v := range pool.txx {
		delete(pool.txx, k)
		txx[it] = v
		it++
	}
	return txx
}

// Len returns the length of the mempool.
func (pool *Mempool) Len() int {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	return len(pool.txx)
}

// Has checks if a transaction is in the mempool.
func (pool *Mempool) Has(tx *proto.Transaction) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	hash := hex.EncodeToString(types.HashTransaction(tx))
	_, ok := pool.txx[hash]

	return ok
}

// Add adds a transaction to the mempool.
func (pool *Mempool) Add(tx *proto.Transaction) bool {
	if pool.Has(tx) {
		return false
	}

	pool.lock.Lock()
	defer pool.lock.Unlock()

	hash := hex.EncodeToString(types.HashTransaction(tx))
	pool.txx[hash] = tx
	return true
}

// ServerConfig struct.
type ServerConfig struct {
	Version    string
	ListenAddr string
	PrivateKey *crypto.PrivateKey
}

// Node struct.
type Node struct {
	ServerConfig
	Logger *zerolog.Logger

	peerLock sync.RWMutex
	peers    map[proto.NodeClient]*proto.Version
	mempool  *Mempool
	chain    *Chain
	//cache       services.DB
	dialedAddrs map[string]string // Comment: This map is used to keep track of the addresses that have been dialed by this node

	ConsensusEngine consensus.Consensus //consensus.Consensus

	proto.UnimplementedNodeServer
}

// NewNode creates a new node.
func NewNode(cfg ServerConfig, bootstrapNodes []string) *Node {
	logger := util.Logger

	rpbft := consensus.NewPBFTPoA(bootstrapNodes)

	return &Node{
		peers:       make(map[proto.NodeClient]*proto.Version),
		dialedAddrs: make(map[string]string), // Comment: Initialize the map
		Logger:      &logger,
		mempool:     NewMempool(),
		chain:       NewChain(NewMemoryBlockStore(), NewMemoryTXStore()),
		//cache:           &services.BadgerDB{}, // <---- review
		ConsensusEngine: rpbft, // <---- review
		ServerConfig:    cfg,
	}
}

// Start starts the node.
func (n *Node) Start(listenAddr string, bootstrapNodes []string) error {
	n.ListenAddr = listenAddr

	var (
		opts       = []grpc.ServerOption{}
		grpcServer = grpc.NewServer(opts...)
	)
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	proto.RegisterNodeServer(grpcServer, n)

	// Comment: Initialize cache
	//n.cache = services.NewBadgerDB(db_dir)

	n.Logger.Info().Msgf("node running on port: [%s]", n.ListenAddr)

	if len(bootstrapNodes) > 0 {
		go n.bootstrapNetwork(bootstrapNodes)
	}

	if n.PrivateKey != nil {
		go n.validatorLoop()
		go n.ConsensusEngine.Start()
	}

	return grpcServer.Serve(ln)
}

// GetTransaction returns a transaction by hash.
func (n *Node) GetBlock(ctx context.Context, v *proto.BlockSearch) (*proto.BlockSearchResult, error) {
	block, err := n.chain.GetBlockByHeight(int(v.BlockHeight))
	if err != nil {
		return nil, err
	}
	return &proto.BlockSearchResult{Block: block}, nil
}

// Handshake performs a handshake with a remote node.
func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	c, err := makeNodeClient(v.ListenAddr)
	if err != nil {
		return nil, err
	}

	n.addPeer(c, v)

	return n.getVersion(), nil
}

// HandleTransaction handles incoming transaction.
func (n *Node) HandleTransaction(ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {
	peer, _ := peer.FromContext(ctx)

	// verify the transaction signature using public key
	if !types.VerifyTransaction(tx) {
		n.Logger.Error().Msgf("invalid initial transaction signature check")
		return nil, nil
	}

	if n.mempool.Add(tx) {
		baseTx := types.CopyTransaction(tx)
		baseTx.Inputs[0].Signature = nil
		baseTx.Inputs[0].PublicKey = nil
		hash := hex.EncodeToString(types.HashTransaction(baseTx)) // <---- without signature and pubkey
		n.Logger.Debug().Msgf("received transaction from [%s] [%s] with hash [%s%s%s] sign [%s] pk [%s]",
			peer.Addr, n.ListenAddr, red, hash[:3], reset, hex.EncodeToString(tx.Inputs[0].Signature)[:3], hex.EncodeToString(tx.Inputs[0].PublicKey)[:3])
		n.Logger.Debug().Msgf("payload: [%s]", string(tx.Outputs[0].Payload))
		go func() {
			if err := n.broadcast(tx); err != nil {
				n.Logger.Error().Msgf("failed to broadcast transaction [%s]", err)
			}
		}()
	}

	return &proto.Ack{}, nil
}

// HandleBlock handles incoming block.
func (n *Node) HandleBlock(ctx context.Context, bk *proto.Block) (*proto.Ack, error) {
	//peer, _ := peer.FromContext(ctx)
	hash := hex.EncodeToString(types.HashBlock(bk))
	height := bk.Header.Height
	size := len(bk.GetTransactions())

	n.Logger.Info().Msgf("received block [%s] with height [%d] and [%d] transaction/s",
		hash[:3], height, size)

	n.chain.AddBlock(bk)

	return &proto.Ack{}, nil
}

func initBlock(chain *Chain) *proto.Block {
	var (
		hash       []byte
		prevHash   []byte
		prevHeight int64
		db_dir     = util.LoadConfig().BADGER.DataDir
	)

	if services.FolderExists(db_dir) {
		DB, err := services.ConnectBadgerDB(db_dir)
		if err != nil {
			logger.Error().Msgf("failed to connect to badgerDB: [%s]", err)
		}
		_, prevHeight, hash, err = DB.GetLatestRecord()
		if err != nil {
			logger.Error().Msgf("failed to get latest block index: [%s]", err)
		}
		prevHash = []byte(hash)
		DB.Close()
	} else {
		var err error
		prevBlock, err := chain.GetBlockByHeight(chain.Height())
		prevHash = types.HashBlock(prevBlock)
		if err != nil {
			logger.Panic().Msgf("failed to get previous block height: [%s]", err)
		}
	}

	header := &proto.Header{
		Version:   1,                     // from config file
		Height:    int32(prevHeight) + 1, // int32(chain.Height() + 1),  // current size of blockStore + 1
		PrevHash:  prevHash,              // types.HashBlock(prevBlock), // previous full block hash
		RootHash:  nil,                   // merkle root hash, to be calculated
		Timestamp: time.Now().UnixNano(),
	}

	return &proto.Block{
		Header:       header,
		PublicKey:    nil,                    // public key of the validator, to be calculated <--- node public key
		Signature:    nil,                    // signature of the block, to be calculated <--- added later
		Transactions: []*proto.Transaction{}, // to be calculated
	}
}

// ValidatorLoop is the main loop for the validator node.
func (n *Node) validatorLoop() {
	n.Logger.Debug().Msgf("validator loop started with blocktime [%s] - waiting for transactions...", blockTime)
	ticker := time.NewTicker(blockTime)
	//refactoring
	var (
		privKey   = crypto.NewPrivateKeyFromSeedStr(util.LoadConfig().KEYS.GodSeed) // <---- this has to be refactored
		recipient = []byte{}
		db_dir    = util.LoadConfig().BADGER.DataDir // <---- this has to be refactored
	)

	privKey, err := crypto.LoadPrivateKeyFromFile("private_key.txt")
	if err != nil {
		logger.Fatal().Msgf("failed to load private key: %s", err)
	}
	recipient = privKey.Public().Address().Bytes()

	for {
		<-ticker.C

		txx := n.mempool.Clear() // Load all transactions to txx clean the mempool
		//n.Logger.Debug().Msgf("memPool [%d] txStore [%d] blockStore [%d]", len(txx), n.chain.txStore.Size(), n.chain.blockStore.Size())

		// check if transactions are available
		if len(txx) > 0 {
			// create a new block
			block := initBlock(n.chain)

			n.Logger.Debug().Msgf("(1) building a new block height [%d] with [%d] transactions", n.chain.Height()+1, len(txx))

			blockTemplateHash := hex.EncodeToString(types.HashBlock(block)) // <---- this has to be refactored
			// if n.chain.Height() != 0 {
			// 	prevBlock, err := n.chain.GetBlockByHeight(n.chain.Height())
			// 	if err != nil {
			// 		logger.Panic().Msgf("failed to get previous block height: [%s]", err)
			// 	}
			// 	block.Header.PrevHash = types.HashBlock(prevBlock)
			// }

			// store all transactions from mempool to txStore
			for _, tx := range txx {
				n.chain.txStore.Put(tx)
				baseTx := types.CopyTransaction(tx)
				baseTx.Inputs[0].Signature = nil
				baseTx.Inputs[0].PublicKey = nil
				// baseTx.Inputs[0].PrevTxHash = nil
				util.Logger.Debug().Msgf("baseTx: %v", baseTx)
				logger.Debug().Msgf("(2) transaction [%s%s%s] stored to mempool",
					red, hex.EncodeToString(types.HashTransaction(baseTx))[:3], reset)
			}

			// get all transactions from txStore
			txs := n.chain.txStore.GetAll()
			logger.Debug().Msgf("(3) getting all transactions from txStore [%d]", len(txs))

			// keep hash of previous tranaction

			//prevTx := txs[0]

			// iterate over transactions
			for _, tx := range txs {
				// build transaction
				baseTx := types.CopyTransaction(tx)
				if string(baseTx.Outputs[0].Payload) != "genesis" {
					baseTx.Inputs[0].Signature = nil
					baseTx.Inputs[0].PublicKey = nil
					logger.Debug().Msgf("(4) preparing transaction [%s%s%s] --------------", red, hex.EncodeToString(types.HashTransaction(baseTx))[:3], reset)
				} else {
					logger.Debug().Msgf("(4) building transaction [%sgenesis%s] -----------------------", red, reset)
				}
				prevTx, _ := n.chain.txStore.Get("b451bea77c3e4255abc7e021a66c636b417a48b89a23f271c0dabb25a01042b7") //genesis tx
				var inputs = []*proto.TxInput{}
				if tx.Inputs != nil {
					inputs = []*proto.TxInput{
						{
							PrevTxHash:   nil,
							PrevOutIndex: tx.Inputs[0].PrevOutIndex,
							PublicKey:    nil,
							Signature:    nil,
						},
					}
				} else {
					inputs = []*proto.TxInput{
						{
							PrevTxHash:   []byte(hex.EncodeToString(types.HashTransaction(prevTx))),
							PrevOutIndex: uint32(0),
							PublicKey:    nil,
							Signature:    nil,
						},
					}
				}
				outputs := []*proto.TxOutput{
					{
						Amount:  1,                     // <---- this has to be refactored
						Address: recipient,             // <---- this has to be refactored
						Payload: tx.Outputs[0].Payload, // <---- this has to be refactored c:metadata
					},
				}
				tx := &proto.Transaction{
					Version:   1,
					Timestamp: tx.Timestamp,
					Inputs:    inputs,
					Outputs:   outputs,
				}

				originalTx := hex.EncodeToString(types.HashTransaction(baseTx))
				tx.Inputs[0].PrevTxHash = []byte(originalTx)

				sig := types.SignTransaction(privKey, tx) // <---- this has to be refactored
				tx.Inputs[0].Signature = sig.Bytes()      // <---- this has to be refactored
				pubKey := privKey.Public().Bytes()
				tx.Inputs[0].PublicKey = pubKey

				logger.Debug().Msgf("--- tx: sign [%s] pubKey [%s] payload [%s%s%s]",
					hex.EncodeToString(tx.Inputs[0].GetSignature())[:3], hex.EncodeToString(tx.Inputs[0].GetPublicKey())[:3], red, string(tx.Outputs[0].Payload), reset)

				// append transaction to block
				logger.Debug().Msgf("(5) adding updated transaction [%s%s%s] is now [%s%s%s] and appended to block template [%s]",
					red, originalTx[:3], reset, red, hex.EncodeToString(types.HashTransactionNoSigPuK(tx))[:3], reset, blockTemplateHash[:3])
				block.Transactions = append(block.Transactions, tx)

				prevTx = tx
			}

			// build merkle tree
			tree, err := types.GetMerkleTree(block)
			if err != nil {
				logger.Panic().Msgf("failed to build merkle tree: [%s]", err)
			}
			block.Header.RootHash = tree.MerkleRoot()

			logger.Debug().Msgf("(6) block template [%s] built with [%d] transactions, height [%d], prevHash [%s] and merkle [%s]",
				blockTemplateHash[:3],
				len(block.Transactions),
				block.GetHeader().Height,
				hex.EncodeToString(block.GetHeader().PrevHash)[:3],
				hex.EncodeToString(block.GetHeader().RootHash)[:3])

			//logger.Debug().Msgf("pubKey: [%s]", hex.EncodeToString(privKey.Public().Bytes()))

			types.SignBlock(privKey, block)
			logger.Debug().Msgf("(7) new block [%s] (from template [%s]) has been created and signed", hex.EncodeToString(types.HashBlock(block))[:3], blockTemplateHash[:3])

			// // add validation here (remove chain.AddBlock(block)) <---- this has to be refactored
			// ver := types.VerifyBlock(block)
			// if !ver {
			// 	logger.Error().Msgf("(Err) block verification failed: [%s]", hex.EncodeToString(types.HashBlock(block)))
			// 	continue
			// }

			n.chain.txStore.Clear()

			var lastBlockHeight int64 = 0

			// BadgerDB
			bdb, err := services.ConnectBadgerDB(db_dir)
			if err != nil {
				logger.Error().Msgf("failed to connect to badgerDB: [%s]", err)
			}
			if services.FolderExists(db_dir) {
				_, lastBlockHeight, _, _ = bdb.GetLatestRecord() //ignore error
			}
			bdb.Set([]byte("blockStore"), []byte(hex.EncodeToString(types.HashBlock(block))), lastBlockHeight+1, types.BlockBytes(block))
			keys, err := bdb.Len([]byte("blockStore"))
			if err != nil {
				logger.Error().Msgf("badger access error (Len)")
			}

			// Couchbase
			cs, err := services.NewCouchbaseService("couchbase://localhost", "Administrator", "password", "blocks", "transactions")
			if err != nil {
				logger.Error().Msgf("failed to create Couchbase service: %v", err)
			}

			err = cs.StoreBlock(block)
			if err != nil {
				logger.Error().Msgf("failed to store block: %v", err)
			}
			cs.Close()

			// probably the place to plug consensus
			n.Logger.Debug().Msgf("(8) proposing block [%s] with [%d] transactions to consensus", hex.EncodeToString(types.HashBlock(block))[:3], len(block.Transactions))
			n.ConsensusEngine.ProposeBlock(block)

			// validate and append to chain (header + merkle + signature, no transactions)
			n.chain.AddBlock(block)

			n.Logger.Info().Msgf("(10) block height [%d] blockStore(M) size [%d] blockStore(P) size [%d] txStore size [%d] headers [%d]",
				n.chain.Height(), n.chain.blockStore.Size(), keys, n.chain.txStore.Size(), n.chain.headers.Height())

			// broadcast the block
			n.broadcast(block)

			bdb.Close()
		}
	}
}

// Broadcast sends a message to all connected peers.
func (n *Node) broadcast(msg any) error {
	for peer := range n.peers {
		switch v := msg.(type) {
		case *proto.Transaction:
			_, err := peer.HandleTransaction(context.Background(), v)
			if err != nil {
				logger.Warn().Msgf("removing peer [%s] from list", peer)
				n.deletePeer(peer)
				return err
			}
		case *proto.Block:
			_, err := peer.HandleBlock(context.Background(), v)
			if err != nil {
				//logger.Warn().Msgf("removing peer [%s] from list", peer)
				//n.deletePeer(peer)
				return err
			}
		}
		// Comment: Add more cases for other message types - block broadcast, etc.
	}
	return nil
}

func (n *Node) addPeer(c proto.NodeClient, v *proto.Version) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()

	// Check if the peer already exists
	if _, exists := n.peers[c]; exists {
		n.Logger.Debug().Msgf("peer already exists: [%s]", v.ListenAddr)
		return
	}

	n.peers[c] = v

	// Connect to all the peers in the peer list
	if len(v.PeerList) > 0 {
		go n.bootstrapNetwork(v.PeerList)
	}

	n.Logger.Info().Msgf("[%s] node successfully connected to [%s] with height [%d]", n.ListenAddr, v.ListenAddr, v.Height)
}

func (n *Node) deletePeer(c proto.NodeClient) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	delete(n.peers, c)
}

func (n *Node) bootstrapNetwork(addrs []string) error {
	//n.Logger.Trace().Msgf("[%s] bootstrap nodes: [%v]", n.listenAddr, addrs)
	for i, addr := range addrs {
		if !n.canConnectWith(addr) {
			continue
		}

		// Comment: Check if the address has already been dialed by this node
		if dialedBy, dialed := n.dialedAddrs[addr]; dialed && dialedBy == n.ListenAddr {
			n.Logger.Debug().Msgf("[%s][%d] already dialed: [%s]", n.ListenAddr, i, addr)
			continue
		}

		// Comment: Check if the address has already been dialed globally
		globalDialedLock.Lock()
		if dialedBy, dialed := globalDialedAddrs[addr]; dialed && dialedBy == n.ListenAddr {
			globalDialedLock.Unlock()
			//n.Logger.Trace().Msgf("[%s][%d] already dialed globally: [%s]", n.listenAddr, i, addr)
			continue
		}
		globalDialedAddrs[addr] = n.ListenAddr
		globalDialedLock.Unlock()

		n.Logger.Debug().Msgf("[%s] node is dialing to: [%s]", n.ListenAddr, addr)

		c, v, err := n.dialRemoteNode(addr)
		if err != nil {
			n.Logger.Error().Msgf("failed to dial remote node: [%s]", err)
			continue
		}

		n.addPeer(c, v)
		n.dialedAddrs[addr] = n.ListenAddr // Mark the address as dialed by this node

		// pull the blockchain from the peer

	}

	return nil
}

func (n *Node) dialRemoteNode(addr string) (proto.NodeClient, *proto.Version, error) {
	c, err := makeNodeClient(addr)
	if err != nil {
		return nil, nil, err
	}

	v, err := c.Handshake(context.Background(), n.getVersion())
	if err != nil {
		return nil, nil, err
	}

	return c, v, nil
}

func (n *Node) getVersion() *proto.Version {
	return &proto.Version{
		Version:    "darkblock-0.1",
		Height:     0,
		ListenAddr: n.ListenAddr,
		PeerList:   n.getPeerList(),
	}
}

func (n *Node) canConnectWith(addr string) bool {
	if n.ListenAddr == addr {
		return false
	}

	connectedPeers := n.getPeerList()
	for _, connectAddr := range connectedPeers {
		if connectAddr == addr {
			return false
		}
	}

	return true
}

func (n *Node) getPeerList() []string {
	n.peerLock.RLock()
	defer n.peerLock.RUnlock()

	peers := []string{}
	for _, version := range n.peers {
		peers = append(peers, version.ListenAddr)
	}

	return peers
}

func makeNodeClient(listenAddr string) (proto.NodeClient, error) {
	c, err := grpc.Dial(listenAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return proto.NewNodeClient(c), nil
}
