package dht

import "math/big"

const K = 20
const B = 20

type Bucket struct {
	lower, upper *big.Int
	nodes        []Node
	replacements []Node
}

func NewBucket() *Bucket {
	lower := big.NewInt(0)
	upper := new(big.Int).Exp(big.NewInt(2), big.NewInt(B*8), nil)
	return &Bucket{
		lower:        lower,
		upper:        upper,
		nodes:        make([]Node, 0),
		replacements: make([]Node, 0),
	}
}

func NewBucketWithRange(lower, upper *big.Int) *Bucket {
	return &Bucket{
		lower:        lower,
		upper:        upper,
		nodes:        make([]Node, 0),
		replacements: make([]Node, 0),
	}
}
