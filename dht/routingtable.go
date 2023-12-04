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

func (rt *RoutingTable) AddNode(node *Node) bool {
	bucket := rt.getBucketFor(node.Key)
	// This will succeed unless the bucket is full.
	if bucket.Add(node) {
		return true
	}
	// Per section 4.2 of the paper, split if the bucket has the own node in its range or
	// if the depth is not congruent to 0 mod 5.
	if bucket.FitsInRange(rt.root.Key) || bucket.Depth()%5 != 0 {
		return rt.splitAndAddNode(bucket, node)
	}
	// TODO: Section 4.1 of the Kademlia paper!
	panic(errors.New("TODO: Section 4.1 of the Kademlia paper!"))
}

func (rt *RoutingTable) RemoveNode(node *Node) bool {
	return rt.getBucketFor(node.Key).Remove(node)
}

func (rt *RoutingTable) FindNodeByKey(key *Key) *Node {
	return rt.getBucketFor(key).FindNodeByKey(key)
}

func (rt *RoutingTable) FindNearbyNodesByKey(key *Key) []*Node {
	nodes := rt.getBucketFor(key).toSlice()
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

func (rt *RoutingTable) getBucketFor(key *Key) *Bucket {
	for _, bucket := range rt.buckets {
		if bucket.Fits(key) {
			return bucket
		}
	}
	// This should not happen if the routing table is initialized correctly.
	panic(errors.New("no suitable bucket found"))
}

func (rt *RoutingTable) splitAndAddNode(bucket *Bucket, node *Node) bool {
	lowerIndex := -1
	for idx, buc := range rt.buckets {
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
	rt.buckets[lowerIndex] = lowerBucket
	rt.buckets = append(rt.buckets[:upperIndex], upperBucket)
	rt.buckets = append(rt.buckets, &Bucket{}) // Placeholder for the new bucket.
	return rt.AddNode(node)
}
