package consensus

import "github.com/janrockdev/darkblock/proto"

type Consensus interface {
	Start()
	Stop()
	ProposeBlock(b *proto.Block) error
	ValidateBlock(b *proto.Block) bool
	OnReceiveMessage(msg ConsensusMessage)
}

type ConsensusMessage struct {
	Type      string
	Block     *proto.Block
	Signature []byte
	NodeID    string
	View      int
}
