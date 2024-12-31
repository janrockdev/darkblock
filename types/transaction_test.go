package types

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/janrockdev/darkblock/crypto"
	"github.com/stretchr/testify/assert"
)

// TestNewTransaction tests the creation of a new transaction.
func TestNewTransaction(t *testing.T) {
	fromPrivKey := crypto.GeneratePrivateKey()
	fromAddress := fromPrivKey.Public().Address().Bytes()
	toPrivKey := crypto.GeneratePrivateKey()
	toAddress := toPrivKey.Public().Address().Bytes()

	fmt.Printf("From key: %s, from address: %s, to key: %s, to address: %s \n", hex.EncodeToString(fromPrivKey.Bytes()), hex.EncodeToString(fromAddress), hex.EncodeToString(toPrivKey.Bytes()), hex.EncodeToString(toAddress))

	assert.NotNil(t, fromPrivKey)
	assert.NotNil(t, toPrivKey)
	assert.NotNil(t, fromAddress)
	assert.NotNil(t, toAddress)

	// // Create a new transaction
	// input := &proto.TxInput{
	// 	PrevTxHash:   util.RandomHash(),
	// 	PrevOutIndex: 0,
	// 	PublicKey:    fromPrivKey.Public().Bytes(),
	// }

	// // Create a new transaction output
	// output1 := &proto.TxOutput{
	// 	Amount:  5,
	// 	Address: toAddress,
	// }

	// // Create a new transaction output
	// output2 := &proto.TxOutput{
	// 	Amount:  95,
	// 	Address: fromAddress,
	// }

	// // Create a new transaction
	// tx := &proto.Transaction{
	// 	Version: 1,
	// 	Inputs:  []*proto.TxInput{input},
	// 	Outputs: []*proto.TxOutput{output1, output2},
	// }

	// // Sign the transaction
	// sig := SignTransaction(fromPrivKey, tx)
	// input.Signature = sig.Bytes()

	// assert.True(t, VerifyTransaction(tx))
}
