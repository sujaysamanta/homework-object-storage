package main

import (
	"crypto/sha256"
	"encoding/binary"
	"sort"
	"strconv"
	"sync"
)

type Node string

type HashRing struct {
	replicas int             // Number of virtual nodes per physical node
	keys     []uint64        // Sorted list of hashes (the ring)
	hashMap  map[uint64]Node // Maps hash to physical node
	mutex    sync.RWMutex
}

func NewHashRing(replicas int) *HashRing {
	return &HashRing{
		replicas: replicas,
		hashMap:  make(map[uint64]Node),
	}
}

func (h *HashRing) Add(node Node) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for i := 0; i < h.replicas; i++ {
		vnodeKey := node + Node("-"+strconv.Itoa(i))
		hash := hashKey(string(vnodeKey))
		h.keys = append(h.keys, hash)
		h.hashMap[hash] = node
	}
	sort.Slice(h.keys, func(i, j int) bool {
		return h.keys[i] < h.keys[j]
	})
}

func (h *HashRing) Remove(node Node) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for i := 0; i < h.replicas; i++ {
		vnodeKey := node + Node("-"+strconv.Itoa(i))
		hash := hashKey(string(vnodeKey))
		delete(h.hashMap, hash)
	}
	// Rebuild the keys slice
	h.keys = h.keys[:0]
	for hash := range h.hashMap {
		h.keys = append(h.keys, hash)
	}
	sort.Slice(h.keys, func(i, j int) bool {
		return h.keys[i] < h.keys[j]
	})
}

func (h *HashRing) Get(key string) Node {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if len(h.keys) == 0 {
		return ""
	}

	hash := hashKey(key)

	// Binary search for the closest hash >= our hash
	idx := sort.Search(len(h.keys), func(i int) bool {
		return h.keys[i] >= hash
	})

	if idx == len(h.keys) {
		idx = 0 // wrap around
	}

	return h.hashMap[h.keys[idx]]
}

func (h *HashRing) GetN(key string, count int) []Node {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	var result []Node
	seen := make(map[Node]bool)

	if len(h.keys) == 0 || count <= 0 {
		return result
	}

	hash := hashKey(key)
	idx := sort.Search(len(h.keys), func(i int) bool {
		return h.keys[i] >= hash
	})

	for len(result) < count {
		node := h.hashMap[h.keys[idx%len(h.keys)]]
		if !seen[node] {
			result = append(result, node)
			seen[node] = true
		}
		idx++
	}
	return result
}

func hashKey(key string) uint64 {
	h := sha256.Sum256([]byte(key))
	// Use the first 8 bytes of SHA256
	return binary.BigEndian.Uint64(h[:8])
}
