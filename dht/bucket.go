package dht

import "math/big"

type Bucket struct {
	lower, upper *big.Int
	nodes        []Node
	replacements []Node
}
