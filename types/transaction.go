package types

import (
	//"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/sha3"

	"github.com/janrockdev/darkblock/crypto"
	"github.com/janrockdev/darkblock/proto"
	"github.com/janrockdev/darkblock/util"
	pb "google.golang.org/protobuf/proto"
)

// VerifyBlock verifies the signature of a block.
func SignTransaction(pk *crypto.PrivateKey, tx *proto.Transaction) *crypto.Signature {
	return pk.Sign(HashTransaction(tx))
}

// HashTransactions returns the sha hash of a transaction.
func HashTransaction(tx *proto.Transaction) []byte {
	b, err := pb.Marshal(tx)
	if err != nil {
		util.Logger.Error().Msgf("error marshalling transaction [%s]", err)
		panic(err)
	}
	//hash := sha256.Sum256(b)
	hash := sha3.New512()
	hash.Write(b)
	hashBytes := hash.Sum(nil)

	return hashBytes[:]
}

// HashTransactions returns the sha hash of a transaction.
func HashTransactionNoSigPuK(tx *proto.Transaction) []byte {

	ctx := CopyTransaction(tx)

	ctx.Inputs[0].Signature = nil
	ctx.Inputs[0].PublicKey = nil

	b, err := pb.Marshal(ctx)
	if err != nil {
		util.Logger.Error().Msgf("error marshalling transaction [%s]", err)
		panic(err)
	}
	//hash := sha256.Sum256(b)
	hash := sha3.New512()
	hash.Write(b)
	hashBytes := hash.Sum(nil)

	return hashBytes[:]
}

// VerifyTransaction verifies the signature of a transaction.
func VerifyTransaction(tx *proto.Transaction) bool {
	for _, input := range tx.Inputs {
		if len(input.Signature) == 0 {
			util.Logger.Error().Msgf("empty transaction signature found [%s]", input.String())
			panic("empty transaction signature")
		}

		var (
			sig    = crypto.SignatureFromBytes(input.Signature)
			pubKey = crypto.PublicKeyFromBytes(input.PublicKey)
		)

		// make sure the signature is nil before verifying
		tmpSig := input.Signature
		tmpPubKey := input.PublicKey
		input.Signature = nil
		input.PublicKey = nil
		util.Logger.Trace().Msgf("before verification tx: [%s] signature: [%s] public key: [%s]", hex.EncodeToString(HashTransaction(tx))[:3], hex.EncodeToString(tmpSig)[:3], hex.EncodeToString(tmpPubKey)[:3])
		if !sig.Verify(pubKey, HashTransaction(tx)) {
			return false
		}
		input.Signature = tmpSig
		input.PublicKey = tmpPubKey
	}

	return true
}

// CopyTransaction creates a deep copy of a transaction.
func CopyTransaction(tx *proto.Transaction) *proto.Transaction {
	// Copy inputs
	inputs := make([]*proto.TxInput, len(tx.Inputs))
	for i, input := range tx.Inputs {
		inputs[i] = &proto.TxInput{
			PrevTxHash:   append([]byte(nil), input.PrevTxHash...),
			PrevOutIndex: input.PrevOutIndex,
			PublicKey:    append([]byte(nil), input.PublicKey...),
			Signature:    append([]byte(nil), input.Signature...),
		}
	}

	// Copy outputs
	outputs := make([]*proto.TxOutput, len(tx.Outputs))
	for i, output := range tx.Outputs {
		outputs[i] = &proto.TxOutput{
			Amount:  output.Amount,
			Address: append([]byte(nil), output.Address...),
			Payload: append([]byte(nil), output.Payload...),
		}
	}

	// Create a new transaction with copied inputs and outputs
	return &proto.Transaction{
		Version: tx.Version,
		Inputs:  inputs,
		Outputs: outputs,
	}
}
