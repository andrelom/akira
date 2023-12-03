package main

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"sort"
	"sync"
	"time"
)

const (
	bucketSize  = 20
	replication = 3
	alpha       = 3
	expiration  = 3600 // 1 hour in seconds
)

// Node represents a Kademlia node.
type Node struct {
	ID        string
	Address   string
	Neighbors []string
	Values    map[string]valueWithExpiration
	Lock      sync.Mutex
}

// valueWithExpiration represents a value with an expiration timestamp.
type valueWithExpiration struct {
	Value      string
	Expiration time.Time
}

// Kademlia implements the Kademlia DHT.
type Kademlia struct {
	Node
	Buckets  [160][]string
	Listener net.Listener
}

// RPC interface for communication between nodes.
type RPC interface {
	Ping(req struct{}, resp *struct{}) error
	Store(req StoreRequest, resp *struct{}) error
	FindNode(req FindNodeRequest, resp *FindNodeResponse) error
	FindValue(req FindValueRequest, resp *FindValueResponse) error
}

// StoreRequest represents the request for storing a value in the DHT.
type StoreRequest struct {
	Key   string
	Value string
}

// FindNodeRequest represents the request for finding the closest nodes to a given ID.
type FindNodeRequest struct {
	Key string
}

// FindNodeResponse represents the response for finding the closest nodes to a given ID.
type FindNodeResponse struct {
	Nodes []string
}

// FindValueRequest represents the request for finding a value in the DHT.
type FindValueRequest struct {
	Key string
}

// FindValueResponse represents the response for finding a value in the DHT.
type FindValueResponse struct {
	Value string
	Nodes []string
}

func main2() {
	node := NewNode()
	kademlia := NewKademlia(node)

	go func() {
		rpc.Register(kademlia)
		rpc.HandleHTTP()
		listener, err := net.Listen("tcp", node.Address)
		if err != nil {
			fmt.Println("Error starting listener:", err)
			return
		}
		fmt.Println("Node listening on", node.Address)
		kademlia.Listener = listener
		err = kademlia.Run()
		if err != nil {
			fmt.Println("Error:", err)
		}
	}()

	go kademlia.PeriodicRefresh() // Start periodic refresh

	time.Sleep(time.Second) // Allow some time for the node to start

	// Join the network by connecting to an existing node
	otherNode := NewNode()
	otherNode.Address = "127.0.0.1:26047"
	kademlia.JoinNetwork(otherNode.Address)

	// Example: Store a value
	key := "exampleKey"
	value := "exampleValue"
	kademlia.StoreValue(key, value)

	// Example: Find a value
	result, err := kademlia.FindValue(key)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Found value:", result)
	}

	// Wait for user input to exit
	fmt.Println("Press Enter to exit...")
	fmt.Scanln()
	kademlia.Listener.Close()
}

// NewNode creates a new Kademlia node with a random ID and address.
func NewNode() *Node {
	id := generateRandomID()
	address := fmt.Sprintf("127.0.0.1:%d", rand.Intn(50000)+10000)
	return &Node{
		ID:        id,
		Address:   address,
		Neighbors: []string{},
		Values:    make(map[string]valueWithExpiration),
		Lock:      sync.Mutex{},
	}
}

// NewKademlia creates a new Kademlia instance with the given node.
func NewKademlia(node *Node) *Kademlia {
	return &Kademlia{Node: *node}
}

// Run starts the RPC server for the node.
func (kademlia *Kademlia) Run() error {
	return http.Serve(kademlia.Listener, nil)
}

// JoinNetwork adds a new node to the Kademlia network.
func (kademlia *Kademlia) JoinNetwork(address string) {
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		fmt.Println("Error joining network:", err)
		return
	}
	defer client.Close()

	var resp FindNodeResponse
	err = client.Call("Kademlia.FindNode", FindNodeRequest{Key: kademlia.ID}, &resp)
	if err != nil {
		fmt.Println("Error joining network:", err)
		return
	}

	// Add the new node to the local node's neighbors
	kademlia.Lock.Lock()
	kademlia.Neighbors = append(kademlia.Neighbors, resp.Nodes...)
	kademlia.Lock.Unlock()

	// Update buckets
	kademlia.UpdateBuckets()
}

// FindClosestNodes finds the closest nodes to a given key.
func (kademlia *Kademlia) FindClosestNodes(key string, count int) []string {
	kademlia.Lock.Lock()
	defer kademlia.Lock.Unlock()

	numNeighbors := len(kademlia.Neighbors)
	if numNeighbors == 0 {
		return nil
	}

	// Ensure count does not exceed the number of neighbors
	if count > numNeighbors {
		count = numNeighbors
	}

	distances := make([]string, numNeighbors)
	copy(distances, kademlia.Neighbors)

	sort.Slice(distances, func(i, j int) bool {
		return getNodeDistance(key, distances[i]) < getNodeDistance(key, distances[j])
	})

	return distances[:count]
}

// FindNode finds the K closest nodes to a given key.
func (kademlia *Kademlia) FindNode(req FindNodeRequest, resp *FindNodeResponse) error {
	kademlia.Lock.Lock()
	defer kademlia.Lock.Unlock()

	// Find the closest nodes to the requested key
	closestNodes := kademlia.FindClosestNodes(req.Key, bucketSize)

	// Exclude the requesting node from the response
	var filteredNodes []string
	for _, node := range closestNodes {
		if node != kademlia.ID {
			filteredNodes = append(filteredNodes, node)
		}
	}

	resp.Nodes = filteredNodes

	return nil
}

// StoreValue stores a value in the Kademlia DHT with expiration and replication.
func (kademlia *Kademlia) StoreValue(key, value string) {
	// Store the value locally with an expiration timestamp
	kademlia.Lock.Lock()
	defer kademlia.Lock.Unlock()
	kademlia.Values[key] = valueWithExpiration{
		Value:      value,
		Expiration: time.Now().Add(time.Duration(expiration) * time.Second),
	}

	// Find the closest nodes to the key
	closestNodes := kademlia.FindClosestNodes(key, alpha)

	// Store the value in the closest nodes
	for _, nodeID := range closestNodes {
		client, err := rpc.DialHTTP("tcp", kademlia.Buckets[getNodeDistance(kademlia.ID, nodeID)][0])
		if err != nil {
			fmt.Println("Error storing value:", err)
			continue
		}
		defer client.Close()

		req := StoreRequest{
			Key:   key,
			Value: value,
		}
		var resp struct{}
		err = client.Call("Kademlia.Store", req, &resp)
		if err != nil {
			fmt.Println("Error storing value:", err)
		}
	}

	// Update buckets after storing
	kademlia.UpdateBuckets()
}

// FindValue retrieves a value from the Kademlia DHT with expiration.
func (kademlia *Kademlia) FindValue(key string) (string, error) {
	// Try to find the value locally
	kademlia.Lock.Lock()
	localValue, exists := kademlia.Values[key]
	kademlia.Lock.Unlock()
	if exists && time.Now().Before(localValue.Expiration) {
		return localValue.Value, nil
	}

	// Find the closest nodes to the key
	closestNodes := kademlia.FindClosestNodes(key, alpha)

	// Query the closest nodes for the value
	for _, nodeID := range closestNodes {
		client, err := rpc.DialHTTP("tcp", kademlia.Buckets[getNodeDistance(kademlia.ID, nodeID)][0])
		if err != nil {
			fmt.Println("Error finding value:", err)
			continue
		}
		defer client.Close()

		req := FindValueRequest{
			Key: key,
		}
		var resp FindValueResponse
		err = client.Call("Kademlia.FindValue", req, &resp)
		if err != nil {
			fmt.Println("Error finding value:", err)
			continue
		}

		// Cache the result locally
		if resp.Value != "" {
			kademlia.Lock.Lock()
			kademlia.Values[key] = valueWithExpiration{
				Value:      resp.Value,
				Expiration: time.Now().Add(time.Duration(expiration) * time.Second),
			}
			kademlia.Lock.Unlock()
			return resp.Value, nil
		}
	}

	return "", fmt.Errorf("value not found")
}

// Periodically refresh buckets to maintain an up-to-date view of the network.
func (kademlia *Kademlia) PeriodicRefresh() {
	ticker := time.NewTicker(time.Hour) // Refresh every hour
	defer ticker.Stop()

	for range ticker.C {
		kademlia.Lock.Lock()
		kademlia.UpdateBuckets()
		kademlia.Lock.Unlock()
	}
}

// UpdateBuckets updates the Kademlia node's buckets after joining or storing values.
func (kademlia *Kademlia) UpdateBuckets() {
	kademlia.Lock.Lock()
	defer kademlia.Lock.Unlock()

	// Sort neighbors by distance and last contact time
	sort.Slice(kademlia.Neighbors, func(i, j int) bool {
		distI := getNodeDistance(kademlia.ID, kademlia.Neighbors[i])
		distJ := getNodeDistance(kademlia.ID, kademlia.Neighbors[j])
		return distI < distJ || (distI == distJ && kademlia.Neighbors[i] < kademlia.Neighbors[j])
	})

	// Distribute neighbors into buckets
	for _, neighbor := range kademlia.Neighbors {
		distance := getNodeDistance(kademlia.ID, neighbor)
		if len(kademlia.Buckets[distance]) < bucketSize {
			kademlia.Buckets[distance] = append(kademlia.Buckets[distance], neighbor)
		}
	}

	// Split buckets if needed
	for i := range kademlia.Buckets {
		if len(kademlia.Buckets[i]) > bucketSize {
			kademlia.splitBucket(i)
		}
	}
}

// splitBucket splits a bucket into two equal-sized buckets.
func (kademlia *Kademlia) splitBucket(index int) {
	oldBucket := kademlia.Buckets[index]
	medianIndex := len(oldBucket) / 2
	medianID := oldBucket[medianIndex]

	// Create two new buckets
	newBucket1 := oldBucket[:medianIndex]
	newBucket2 := oldBucket[medianIndex:]

	// Update the old bucket
	kademlia.Buckets[index] = newBucket1

	// Determine the index for the new bucket
	newBucketIndex := getNodeDistance(kademlia.ID, medianID)

	// Calculate the new size of the array
	newSize := max(len(kademlia.Buckets), newBucketIndex+1)
	newBuckets := make([][]string, newSize)

	// Copy existing buckets to the new array
	for i := range kademlia.Buckets {
		newBuckets[i] = make([]string, len(kademlia.Buckets[i]))
		copy(newBuckets[i], kademlia.Buckets[i])
	}

	// Add the new bucket
	newBuckets[newBucketIndex] = newBucket2

	// Update the Kademlia node's buckets
	for i := range kademlia.Buckets {
		copy(kademlia.Buckets[i], newBuckets[i])
	}
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// getNodeDistance calculates the XOR distance between two node IDs.
func getNodeDistance(id1, id2 string) int {
	hash1, _ := hex.DecodeString(id1)
	hash2, _ := hex.DecodeString(id2)

	distance := 0
	for i := range hash1 {
		distance += int(hash1[i] ^ hash2[i])
	}

	return distance
}

// generateRandomID generates a random 160-bit ID.
func generateRandomID() string {
	rand.Seed(time.Now().UnixNano())
	bytes := make([]byte, 20)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
