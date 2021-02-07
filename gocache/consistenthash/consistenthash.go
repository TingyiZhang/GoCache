package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash maps bytes to uint32 (2^32 space)
type Hash func(data []byte) uint32

// Map constains all hashed keys
type Map struct {
	// hash function
	hash Hash

	// the number of virtual nodes
	replicas int

	// the hash ring, sorted
	keys []int

	// key: hash value of virtual nodes, value: name of real nodes
	hashMap map[int]string
}

// New creates a Map instance
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}

	// default hash function
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}

	return m
}

// Add adds keys to the hash.
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		// for every key, create m.replicas replicas
		for i := 0; i < m.replicas; i++ {
			// name: key + i
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			// map the hash of replicas to the real key
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// Get gets the closest item in the hash to the provided key.
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	// Binary search for appropriate replica.
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[idx%len(m.keys)]]
}
