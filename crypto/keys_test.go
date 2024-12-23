package crypto

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePrivateKey(t *testing.T) {
	privKey := GeneratePrivateKey()
	assert.Equal(t, len(privKey.Bytes()), PrivKeyLen)

	pubKey := privKey.Public()
	assert.Equal(t, len(pubKey.Bytes()), PubKeyLen)
}

func TestPrivKeySign(t *testing.T) {
	privKey := GeneratePrivateKey()
	pubKey := privKey.Public()
	msg := []byte("darkblock")

	sig := privKey.Sign(msg)
	assert.True(t, sig.Verify(pubKey, msg))

	// different message
	assert.False(t, sig.Verify(pubKey, []byte("darkblock2")))

	// different public key
	invPrivKey := GeneratePrivateKey()
	invPubKey := invPrivKey.Public()
	assert.False(t, sig.Verify(invPubKey, msg))
}

func TestStoreAndLoadKey(t *testing.T) {
	privKey := GeneratePrivateKey()
	filename := "../private_key.txt"
	err := SavePrivateKeyToFile(privKey, filename)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Private key saved to %s", filename)

	// Load the private key from the file
	loadedPrivKey, err := LoadPrivateKeyFromFile(filename)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Private key loaded from %s", filename)

	// Verify that the loaded key matches the original key
	if !assert.Equal(t, privKey.Bytes(), loadedPrivKey.Bytes()) {
		t.Errorf("Loaded private key does not match the original key")
	}
	fmt.Println("Loaded private key matches the original key")
}

func TestPrivateKeyFromString(t *testing.T) {
	var (
		seed       = "4a9fb8494f467fd001fad589342a3d63c4ddc148a119b76b0d14f4655fbb09f7"
		privKey    = NewPrivateKeyFromString(seed)
		addressStr = "04a4d41f57569fc850c6bba317a623fdefba61c0"
	)
	assert.Equal(t, PrivKeyLen, len(privKey.Bytes()))
	address := privKey.Public().Address()
	assert.Equal(t, addressStr, address.String())
}

func TestPubKeyToAddress(t *testing.T) {
	privKey := GeneratePrivateKey()
	pubKey := privKey.Public()
	addr := pubKey.Address()
	assert.Equal(t, AddressLen, len(addr.Bytes()))
}
