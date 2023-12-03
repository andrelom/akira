package dht

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"math/big"
)

type Key struct {
	data *big.Int
}

func NewKey() *Key {
	data := getRandomBytes()
	return &Key{
		data: new(big.Int).Abs(new(big.Int).SetBytes(data)),
	}
}

func NewKeyFromBytes(bytes []byte) (*Key, error) {
	if !isValid(bytes) {
		return nil, fmt.Errorf("invalid key length")
	}
	return &Key{
		data: new(big.Int).SetBytes(bytes),
	}, nil
}

func (key *Key) DistanceTo(target *Key) *big.Int {
	// XOR can produce negative numbers when dealing with BigIntegers,
	// so we use Abs to ensure the distance is positive.
	return new(big.Int).Abs(new(big.Int).Xor(key.data, target.data))
}

func (key *Key) ToBinaryString() string {
	return fmt.Sprintf("%b", key.data)
}

func (key *Key) ToBigInt() *big.Int {
	return key.data
}

func isValid(bytes []byte) bool {
	return len(bytes) == K
}

func getRandomBytes() []byte {
	hash := sha1.New()
	bytes := make([]byte, 16)
	rand.Read(bytes)
	hash.Write(bytes)
	return hash.Sum(nil)
}