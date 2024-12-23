package types

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/janrockdev/darkblock/crypto"
	"github.com/janrockdev/darkblock/proto"
	"github.com/janrockdev/darkblock/util"
	"github.com/stretchr/testify/assert"
)

func TestCalculateRootHash(t *testing.T) {
	var (
		privKey = crypto.GeneratePrivateKey()
		block   = util.RandomBlock()
		tx      = &proto.Transaction{
			Version: 1,
		}
	)
	block.Transactions = append(block.Transactions, tx)
	SignBlock(privKey, block)

	assert.True(t, VerifyRootHash(block))
	assert.Equal(t, len(block.Header.RootHash), 32)
}

// TestSignVerifyBlock tests the signing and verification of a block.
func TestSignVerifyBlock(t *testing.T) {
	var (
		block   = util.RandomBlock()
		privKey = crypto.GeneratePrivateKey()
		pubKey  = privKey.Public()
	)

	// Sign and verify the block
	sig := SignBlock(privKey, block)
	assert.Equal(t, 64, len(sig.Bytes()))
	assert.True(t, sig.Verify(pubKey, HashBlock(block)))

	// Verify the block
	assert.Equal(t, block.PublicKey, pubKey.Bytes())
	assert.Equal(t, block.Signature, sig.Bytes())
	assert.True(t, VerifyBlock(block))

	// Change the public key
	invalidPrivKey := crypto.GeneratePrivateKey()
	block.PublicKey = invalidPrivKey.Public().Bytes()
	assert.False(t, VerifyBlock(block))
}

// TestHashHeader tests the hashing of a header.
func TestHashBlock(t *testing.T) {
	block := util.RandomBlock()
	hash := HashBlock(block)
	fmt.Println(hex.EncodeToString(hash)) // hash of the block
	assert.Equal(t, 32, len(hash))
}
