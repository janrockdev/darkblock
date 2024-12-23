package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"io"
	"io/ioutil"

	"github.com/janrockdev/darkblock/util"
)

const (
	PrivKeyLen   = 64
	SignatureLen = 64
	PubKeyLen    = 32
	SeedLen      = 32
	AddressLen   = 20
)

type PrivateKey struct {
	key ed25519.PrivateKey
}

func NewPrivateKeyFromString(s string) *PrivateKey {
	b, err := hex.DecodeString(s)
	if err != nil {
		util.Logger.Error().Msgf("error decoding private key [%s]", err)
		panic(err)
	}
	return NewPrivateKeyFromSeed(b)
}

func NewPrivateKeyFromSeedStr(seed string) *PrivateKey {
	seedBytes, err := hex.DecodeString(seed)
	if err != nil {
		util.Logger.Error().Msgf("error decoding seed [%s]", err)
		panic(err)
	}
	return NewPrivateKeyFromSeed(seedBytes)
}

func NewPrivateKeyFromSeed(seed []byte) *PrivateKey {
	if len(seed) != SeedLen {
		panic("invalid seed length")
	}
	return &PrivateKey{
		key: ed25519.NewKeyFromSeed(seed),
	}
}

func GeneratePrivateKey() *PrivateKey {
	seed := make([]byte, SeedLen)
	_, err := io.ReadFull(rand.Reader, seed)
	if err != nil {
		util.Logger.Error().Msgf("error generating private key [%s]", err)
		panic(err)
	}

	return &PrivateKey{
		key: ed25519.NewKeyFromSeed(seed)}
}

// NewPrivateKeyFromBytes creates a PrivateKey from a byte slice.
func NewPrivateKeyFromBytes(b []byte) *PrivateKey {
	if len(b) != PrivKeyLen {
		panic("invalid private key length")
	}
	return &PrivateKey{
		key: ed25519.PrivateKey(b),
	}
}

// SavePrivateKeyToFile saves the private key to a file.
func SavePrivateKeyToFile(privKey *PrivateKey, filename string) error {
	keyBytes := privKey.Bytes()
	keyHex := hex.EncodeToString(keyBytes)
	return ioutil.WriteFile(filename, []byte(keyHex), 0600)
}

// LoadPrivateKeyFromFile loads the private key from a file.
func LoadPrivateKeyFromFile(filename string) (*PrivateKey, error) {
	keyHex, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	keyBytes, err := hex.DecodeString(string(keyHex))
	if err != nil {
		return nil, err
	}
	return NewPrivateKeyFromBytes(keyBytes), nil
}

func (p *PrivateKey) Bytes() []byte {
	return p.key
}

func (p *PrivateKey) Sign(msg []byte) *Signature {
	return &Signature{
		value: ed25519.Sign(p.key, msg),
	}
}

func (p *PrivateKey) Public() *PublicKey {
	b := make([]byte, PubKeyLen)
	copy(b, p.key[32:])

	return &PublicKey{key: b}
}

type PublicKey struct {
	key ed25519.PublicKey
}

func PublicKeyFromBytes(b []byte) *PublicKey {
	if len(b) != PubKeyLen {
		panic("invalid public key length from bytes")
	}
	return &PublicKey{
		key: ed25519.PublicKey(b),
	}
}

func (p *PublicKey) Address() *Address {
	return &Address{
		value: p.key[len(p.key)-AddressLen:],
	}
}

func (p *PublicKey) Bytes() []byte {
	return p.key
}

type Signature struct {
	value []byte
}

func SignTransaction(pk *PrivateKey, tx []byte) *Signature {
	return pk.Sign(tx)
}

func SignatureFromBytes(b []byte) *Signature {
	if len(b) != SignatureLen {
		panic("invalid signature length")
	}
	return &Signature{value: b}
}

func (s *Signature) Bytes() []byte {
	return s.value
}

func (s *Signature) Verify(pubKey *PublicKey, msg []byte) bool {
	return ed25519.Verify(pubKey.key, msg, s.value)
}

type Address struct {
	value []byte
}

func AddressFromBytes(b []byte) Address {
	if len(b) != AddressLen {
		panic("invalid address length")
	}
	return Address{
		value: b,
	}
}

func (a *Address) Bytes() []byte {
	return a.value
}

func (a *Address) String() string {
	return hex.EncodeToString(a.value)
}
