package dht

import "math/big"

const K = 20
const B = 20

type Bucket struct {
	lower, upper *big.Int
	nodes        []*Node
	replacements []*Node
}

func NewBucket() *Bucket {
	lower := big.NewInt(0)
	upper := new(big.Int).Exp(big.NewInt(2), big.NewInt(B*8), nil)
	return &Bucket{
		lower:        lower,
		upper:        upper,
		nodes:        make([]*Node, 0),
		replacements: make([]*Node, 0),
	}
}

func NewBucketWithRange(lower, upper *big.Int) *Bucket {
	return &Bucket{
		lower:        lower,
		upper:        upper,
		nodes:        make([]*Node, 0),
		replacements: make([]*Node, 0),
	}
}

func (buc *Bucket) Depth() int {
	values := make([]string, len(buc.nodes))
	for i, node := range buc.nodes {
		values[i] = node.Key.Text(2)
	}
	shared := getSharedPrefix(values)
	return len(shared)
}

func (buc *Bucket) Fits(key *big.Int) bool {
	return key.Cmp(buc.upper) < 0
}

func (buc *Bucket) FitsInRange(key *big.Int) bool {
	return key.Cmp(buc.lower) >= 0 && buc.upper.Cmp(key) >= 0
}

func (buc *Bucket) Add(node *Node) bool {
	if moved := toTailIfExists(buc.nodes, node); moved {
		return true
	}
	if isFull(buc.nodes) {
		buc.addToReplacements(node)
		return false
	}
	buc.nodes = append(buc.nodes, node)
	return true
}

func (b *Bucket) Remove(node *Node) bool {
	removed := remove(b.nodes, node)
	if removed {
		b.addFromReplacements()
	}
	return remove(b.replacements, node) || removed
}

func (buc *Bucket) FindNodeByKey(key *big.Int) *Node {
	return getNodeByKey(buc.nodes, key)
}

func (buc *Bucket) Split() (*Bucket, *Bucket) {
	nodes := append(buc.nodes, buc.replacements...)
	middle := new(big.Int).Add(buc.lower, buc.upper)
	middle.Div(middle, big.NewInt(2))
	lowerBucket := NewBucketWithRange(buc.lower, middle)
	upperBucket := NewBucketWithRange(new(big.Int).Add(middle, big.NewInt(1)), buc.upper)
	for _, node := range nodes {
		if node.Key.Cmp(middle) <= 0 {
			lowerBucket.Add(node)
		} else {
			upperBucket.Add(node)
		}
	}
	return lowerBucket, upperBucket
}

func (buc *Bucket) addToReplacements(node *Node) {
	if moved := toTailIfExists(buc.replacements, node); moved {
		return
	}
	if isFull(buc.replacements) {
		buc.replacements = buc.replacements[1:]
	}
	buc.replacements = append(buc.replacements, node)
}

func (buc *Bucket) addFromReplacements() {
	if len(buc.replacements) == 0 {
		return
	}
	node := buc.replacements[len(buc.replacements)-1]
	buc.replacements = buc.replacements[:len(buc.replacements)-1]
	buc.nodes = append(buc.nodes, node)
}

func isFull(nodes []*Node) bool {
	return K <= len(nodes)
}

func getNodeByKey(nodes []*Node, key *big.Int) *Node {
	for _, node := range nodes {
		if node.Key.Cmp(key) == 0 {
			return node
		}
	}
	return nil
}

func getSharedPrefix(values []string) string {
	if len(values) == 0 {
		return ""
	}
	prefix := values[0]
	for _, value := range values[1:] {
		idx := 0
		for idx < len(prefix) && idx < len(value) && prefix[idx] == value[idx] {
			idx++
		}
		prefix = prefix[:idx]
	}
	return prefix
}

func toTailIfExists(nodes []*Node, node *Node) bool {
	if getNodeByKey(nodes, node.Key) == nil {
		return false
	}
	for idx, val := range nodes {
		if val.Key.Cmp(node.Key) == 0 {
			nodes = append(nodes[:idx], nodes[idx+1:]...)
			nodes = append(nodes, node)
			return true
		}
	}
	return false
}

func remove(nodes []*Node, target *Node) bool {
	for idx, node := range nodes {
		if node.Key.Cmp(target.Key) == 0 {
			nodes = append(nodes[:idx], nodes[idx+1:]...)
			return true
		}
	}
	return false
}
