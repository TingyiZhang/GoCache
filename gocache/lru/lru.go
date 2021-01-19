package lru

import "container/list"

// Cache is a LRU cache. It is not safe for concurrent access.
type Cache struct {
	// maximum memory allowed
	maxBytes int64

	// memory in use
	nbytes int64

	ll    *list.List
	cache map[string]*list.Element

	// the callback function might be called when an element is deleted
	OnEvicted func(key string, value Value)
}

// entry is the type of elements in our doubly linked list
type entry struct {
	key   string
	value Value
}

// Value is an interface for values stored in linkedlist
// Elements can be any type of variables as long as they implemented Value interface,
// becasue we need Len() to count the memory it takes
type Value interface {
	Len() int
}

// New is the Constructor of Cache
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Add adds a value to the cache.
func (c *Cache) Add(key string, value Value) {

	// type of element: list.Element
	if element, ok := c.cache[key]; ok {
		c.ll.MoveToFront(element)

		// list.Element has an empty interface Value, which is for the value stored with this element
		// In this case, element.Value.(*entry) returns an instance of type entry
		kv := element.Value.(*entry)

		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		element := c.ll.PushFront(&entry{key, value})
		c.cache[key] = element
		c.nbytes += int64(len(key)) + int64(value.Len())
	}

	// Remove oldest element until nbytes < maxBytes
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// Get finds an element with the key, and moves it to the head of linked list
func (c *Cache) Get(key string) (value Value, ok bool) {
	if element, ok := c.cache[key]; ok {
		c.ll.MoveToFront(element)
		kv := element.Value.(*entry)
		return kv.value, ok
	}
	return
}

// RemoveOldest removes the oldest item (the element at the back of the linked list)
func (c *Cache) RemoveOldest() {
	element := c.ll.Back()
	if element != nil {
		c.ll.Remove(element)
		kv := element.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Len the number of cache entries (helper function for testing)
func (c *Cache) Len() int {
	return c.ll.Len()
}
