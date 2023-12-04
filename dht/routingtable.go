package dht

import (
	"errors"
	"sort"
)

type RoutingTable struct {
	root    *Node
	buckets []*Bucket
}

func NewRoutingTable(root *Node) *RoutingTable {
	return &RoutingTable{
		root:    root,
		buckets: []*Bucket{NewBucket()},
	}
}

func (rou *RoutingTable) AddNode(node *Node) bool {
	bucket := rou.getBucketFor(node.Key)
	// This will succeed unless the bucket is full.
	if bucket.Add(node) {
		return true
	}
	// Per section 4.2 of the paper, split if the bucket has the own node in its range or
	// if the depth is not congruent to 0 mod 5.
	if bucket.FitsInRange(rou.root.Key) || bucket.Depth()%5 != 0 {
		return rou.splitAndAddNode(bucket, node)
	}
	// TODO: Section 4.1 of the Kademlia paper!
	panic(errors.New("TODO: Section 4.1 of the Kademlia paper!"))
}

func (rou *RoutingTable) RemoveNode(node *Node) bool {
	return rou.getBucketFor(node.Key).Remove(node)
}

func (rou *RoutingTable) FindNodeByKey(key *Key) *Node {
	return rou.getBucketFor(key).FindNodeByKey(key)
}

func (rou *RoutingTable) FindNearbyNodesByKey(key *Key) []*Node {
	nodes := rou.getBucketFor(key).toSlice()
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Key.DistanceTo(key).Cmp(nodes[j].Key.DistanceTo(key)) < 0
	})
	return nodes
}

func (buc *Bucket) toSlice() []*Node {
	nodes := make([]*Node, len(buc.nodes))
	for i, node := range buc.nodes {
		nodes[i] = node
	}
	return nodes
}

func (rou *RoutingTable) getBucketFor(key *Key) *Bucket {
	for _, bucket := range rou.buckets {
		if bucket.Fits(key) {
			return bucket
		}
	}
	// This should not happen if the routing table is initialized correctly.
	panic(errors.New("no suitable bucket found"))
}

func (rou *RoutingTable) splitAndAddNode(bucket *Bucket, node *Node) bool {
	lowerIndex := -1
	for idx, buc := range rou.buckets {
		if buc == bucket {
			lowerIndex = idx
			break
		}
	}
	if lowerIndex == -1 {
		// This should not happen if the routing table is consistent.
		panic(errors.New("bucket not found in routing table"))
	}
	upperIndex := lowerIndex + 1
	lowerBucket, upperBucket := bucket.Split()
	rou.buckets[lowerIndex] = lowerBucket
	rou.buckets = append(rou.buckets[:upperIndex], upperBucket)
	rou.buckets = append(rou.buckets, &Bucket{}) // Placeholder for the new bucket.
	return rou.AddNode(node)
}
