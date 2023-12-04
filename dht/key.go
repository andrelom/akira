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
	return &Key{
		data: getRandomKey(),
	}
}

func NewKeyFromBytes(bytes []byte) (*Key, error) {
	if len(bytes) != K {
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

func (key *Key) BigInt() *big.Int {
	return key.data
}

func getRandomKey() *big.Int {
	hash := sha1.New()
	bytes := make([]byte, 16)
	rand.Read(bytes)
	hash.Write(bytes)
	data := hash.Sum(nil)
	return new(big.Int).Abs(new(big.Int).SetBytes(data))
}
