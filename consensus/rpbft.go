package consensus

import (
	"encoding/hex"
	"sync"

	"github.com/janrockdev/darkblock/proto"
	"github.com/janrockdev/darkblock/types"
	"github.com/janrockdev/darkblock/util"
)

type PBFTPoA struct {
	mu          sync.Mutex
	validators  []string // <---- review list of node public keys / ports
	quorumSize  int
	currentView int
	isValidator bool
	nodeID      string

	// Buffers for messages
	prePrepareMsgs map[int]*proto.Block
	prepareVotes   map[int]map[string]bool
	commitVotes    map[int]map[string]bool

	blockProposals chan *proto.Block
	incomingMsgs   chan ConsensusMessage

	stopCh chan struct{}
}

func NewPBFTPoA(validators []string) *PBFTPoA {
	return &PBFTPoA{
		validators:     validators,
		quorumSize:     len(validators), // better option 2*len(validators))/3 + 1, // PBFT quorum
		currentView:    0,
		isValidator:    true,         // <---- review hardcoded value for a single node instance - development only
		nodeID:         "validator1", // <---- review hardcoded value for a single node instance - development only
		prePrepareMsgs: make(map[int]*proto.Block),
		prepareVotes:   make(map[int]map[string]bool),
		commitVotes:    make(map[int]map[string]bool),
		blockProposals: make(chan *proto.Block),
		incomingMsgs:   make(chan ConsensusMessage, 100),
		stopCh:         make(chan struct{}),
	}
}

func (p *PBFTPoA) Start() {
	util.Logger.Info().Msg("starting RpBFT consensus protocol")
	go p.run()
}

func (p *PBFTPoA) Stop() {
	close(p.stopCh)
}

func (p *PBFTPoA) run() {
	util.Logger.Info().Msg("incomming message")
	for {
		select {
		case <-p.stopCh:
			return
		case b := <-p.blockProposals:
			util.Logger.Info().Msgf("received block proposal [%s]", hex.EncodeToString(types.HashBlock(b))[:3])
			if p.isLeader() {
				p.handleProposeBlock(b)
			}
		case msg := <-p.incomingMsgs:
			util.Logger.Info().Msgf("received incomming message [%s]", hex.EncodeToString(types.HashBlock(msg.Block)[:3]))
			p.handleMessage(msg)
		}
	}
}

func (p *PBFTPoA) isLeader() bool {
	return true // <---- review p.validators[p.currentView%len(p.validators)] == p.nodeID
}

func (p *PBFTPoA) ProposeBlock(b *proto.Block) error {
	// Non-leaders just queue block proposals. They won't actually propose.
	if !p.isLeader() {
		return nil
	}
	p.blockProposals <- b
	return nil
}

func (p *PBFTPoA) handleProposeBlock(b *proto.Block) {
	// Create a Pre-Prepare message and broadcast it
	msg := ConsensusMessage{
		Type:   "PrePrepare",
		Block:  b,
		View:   p.currentView,
		NodeID: p.nodeID,
	}
	util.Logger.Info().Msgf("proposing block [%s] msg [%d]\n", hex.EncodeToString(types.HashBlock(b))[:3], p.currentView)
	p.broadcast(msg)
}

func (p *PBFTPoA) OnReceiveMessage(msg ConsensusMessage) {
	p.incomingMsgs <- msg
}

func (p *PBFTPoA) handleMessage(msg ConsensusMessage) {
	switch msg.Type {
	case "PrePrepare":
		p.handlePrePrepare(msg)
	case "Prepare":
		p.handlePrepare(msg)
	case "Commit":
		p.handleCommit(msg)
	}
}

func (p *PBFTPoA) handlePrePrepare(msg ConsensusMessage) {
	// Validate leader
	if p.validators[p.currentView%len(p.validators)] != msg.NodeID {
		return
	}
	// Validate block
	if !p.ValidateBlock(msg.Block) {
		return
	}

	p.prePrepareMsgs[msg.View] = msg.Block
	p.prepareVotes[msg.View] = map[string]bool{p.nodeID: true}

	// Broadcast Prepare
	prepareMsg := ConsensusMessage{
		Type:   "Prepare",
		Block:  msg.Block,
		View:   msg.View,
		NodeID: p.nodeID,
	}
	p.broadcast(prepareMsg)
}

func (p *PBFTPoA) handlePrepare(msg ConsensusMessage) {
	// Must have a pre-prepare for this view
	if p.prePrepareMsgs[msg.View] == nil {
		return
	}
	// Validate block matches pre-prepare
	if hex.EncodeToString(types.HashBlock(p.prePrepareMsgs[msg.View])) != hex.EncodeToString(types.HashBlock(msg.Block)) {
		return
	}

	if p.prepareVotes[msg.View] == nil {
		p.prepareVotes[msg.View] = make(map[string]bool)
	}
	p.prepareVotes[msg.View][msg.NodeID] = true

	// Check if we have quorum
	if len(p.prepareVotes[msg.View]) >= p.quorumSize {
		// Broadcast Commit
		if p.commitVotes[msg.View] == nil {
			p.commitVotes[msg.View] = make(map[string]bool)
		}
		p.commitVotes[msg.View][p.nodeID] = true
		commitMsg := ConsensusMessage{
			Type:   "Commit",
			Block:  p.prePrepareMsgs[msg.View],
			View:   msg.View,
			NodeID: p.nodeID,
		}
		p.broadcast(commitMsg)
	}
}

func (p *PBFTPoA) handleCommit(msg ConsensusMessage) {
	// Validate block
	if p.prePrepareMsgs[msg.View] == nil {
		return
	}
	if hex.EncodeToString(types.HashBlock(p.prePrepareMsgs[msg.View])) != hex.EncodeToString(types.HashBlock(msg.Block)) {
		return
	}

	if p.commitVotes[msg.View] == nil {
		p.commitVotes[msg.View] = make(map[string]bool)
	}
	p.commitVotes[msg.View][msg.NodeID] = true

	// Check for quorum of commits
	if len(p.commitVotes[msg.View]) >= p.quorumSize {
		// Finalize the block
		finalizedBlock := p.prePrepareMsgs[msg.View]
		// Add to blockchain storage

		//n.chain.AddBlock(finalizedBlock) <---------------------------------------------
		util.Logger.Warn().Msgf("Finalized block [%s] in view [%d]", hex.EncodeToString(types.HashBlock(finalizedBlock))[:8], msg.View)

		// Move to next view
		p.currentView++
	}
}

func (p *PBFTPoA) broadcast(msg ConsensusMessage) {
	// This depends on how the existing P2P layer is implemented.
	// For example:
	// node.BroadcastConsensusMessage(msg)
}

func (p *PBFTPoA) ValidateBlock(b *proto.Block) bool {
	// Implement block validation logic, similar to what the chain currently does
	// e.g. validate transactions, previous hash, etc.
	return true
}

// // node/node.go
// package node

// import (
//     "github.com/janrockdev/darkblock/block"
//     "github.com/janrockdev/darkblock/consensus"
// )

// type Node struct {
//     ConsensusEngine consensus.Consensus
//     // ... other fields ...
// }

// func NewNode(validators []string, nodeID string, isValidator bool) *Node {
//     pbft := consensus.NewPBFTPoA(validators, nodeID, isValidator)
//     return &Node{
//         ConsensusEngine: pbft,
//         // ... other initialization ...
//     }
// }

// func (n *Node) Start() {
//     n.ConsensusEngine.Start()
//     // ... start network, etc.
// }

// func (n *Node) AddBlockToChain(b *block.Block) {
//     // Add the block to the chain’s storage
//     // ...
// }
