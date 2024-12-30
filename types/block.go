package types

import (
	"bytes"
	//"crypto/sha256"
	"golang.org/x/crypto/sha3"

	"github.com/cbergoon/merkletree"
	"github.com/janrockdev/darkblock/crypto"
	"github.com/janrockdev/darkblock/proto"
	"github.com/janrockdev/darkblock/util"
	pb "google.golang.org/protobuf/proto"
)

var logger = util.Logger

type TxHash struct {
	hash []byte
}

func NewTxHash(hash []byte) TxHash {
	return TxHash{hash: hash}
}

func (h TxHash) CalculateHash() ([]byte, error) {
	return h.hash, nil
}

func (h TxHash) Equals(other merkletree.Content) (bool, error) {
	equals := bytes.Equal(h.hash, other.(TxHash).hash)
	return equals, nil
}

// VerifyBlock verifies the signature of a block.
func VerifyBlock(b *proto.Block) bool {
	if len(b.Transactions) > 0 {
		// Check root hash
		if !VerifyRootHash(b) {
			logger.Error().Msg("invalid root hash")
			return false
		}
	}
	// Check lengths
	if len(b.PublicKey) != crypto.PubKeyLen {
		logger.Error().Msg("invalid public key length")
		return false
	}
	// Check signature
	if len(b.Signature) != crypto.SignatureLen {
		logger.Error().Msg("invalid signature length")
		return false
	}
	// Verify signature
	var (
		sig    = crypto.SignatureFromBytes(b.Signature)
		pubKey = crypto.PublicKeyFromBytes(b.PublicKey)
		hash   = HashBlock(b)
	)
	if !sig.Verify(pubKey, hash) {
		logger.Error().Msg("invalid block signature")
		return false
	}
	return true //sig.Verify(pubKey, hash)
}

// SignBlock signs a block with a private key and returns the signature.
func SignBlock(pk *crypto.PrivateKey, b *proto.Block) *crypto.Signature {
	if len(b.Transactions) > 0 {
		tree, err := GetMerkleTree(b)
		if err != nil {
			util.Logger.Error().Msgf("error creating merkle tree: %v", err)
			panic(err)
		}

		b.Header.RootHash = tree.MerkleRoot()
	}
	hash := HashBlock(b)
	sig := pk.Sign(hash)
	// logger.Debug().Msgf("hash before sign: %x", hash)
	b.PublicKey = pk.Public().Bytes()
	b.Signature = sig.Bytes()

	return sig
}

func VerifyRootHash(b *proto.Block) bool {
	tree, err := GetMerkleTree(b)
	if err != nil {
		logger.Error().Msgf("error creating merkle tree: %v", err)
		return false
	}

	verfTx, err := tree.VerifyTree()
	if err != nil {
		logger.Error().Msgf("error verifying merkle tree: %v", err)
		return false
	}

	if !verfTx {
		logger.Error().Msg("invalid merkle tree")
		return false
	}

	return bytes.Equal(b.Header.RootHash, tree.MerkleRoot())
}

func GetMerkleTree(b *proto.Block) (*merkletree.MerkleTree, error) {
	// block has to have transactions to create a merkle tree

	list := make([]merkletree.Content, len(b.Transactions))
	for i := 0; i < len(b.Transactions); i++ {
		list[i] = NewTxHash(HashTransaction(b.Transactions[i]))
	}

	// Create a new Merkle Tree from the list of content
	t, err := merkletree.NewTree(list)
	if err != nil {
		util.Logger.Error().Msgf("error creating merkle tree: %v", err)
		return nil, err
	}

	return t, nil
}

// HashBlock returns the sha256 hash of a header.
func HashBlock(block *proto.Block) []byte {
	return HashHeader(block.Header)
}

// HashHeader returns the sha256 hash of a header.
func HashHeader(header *proto.Header) []byte {
	b, err := pb.Marshal(header)
	if err != nil {
		util.Logger.Error().Msgf("error marshalling header: %v", err)
		panic(err)
	}

	//hash := sha256.Sum256(b)
	hash := sha3.New512()
	hash.Write(b)
	hashBytes := hash.Sum(nil)

	return hashBytes[:]
}

func BlockBytes(b *proto.Block) []byte {
	data, err := pb.Marshal(b)
	if err != nil {
		util.Logger.Error().Msgf("error marshalling block: %v", err)
		panic(err)
	}

	return data
}

func UnmarshalBlock(serializedBlock []byte) (*proto.Block, error) {
	var block proto.Block
	err := pb.Unmarshal(serializedBlock, &block)
	if err != nil {
		util.Logger.Error().Msgf("error unmarshalling block: %v", err)
		panic(err)
	}

	return &block, nil
}
