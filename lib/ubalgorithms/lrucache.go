package ubalgorithms

// lruCacheNode represents each entry in the doubly linked list.
type lruCacheNode[K comparable, V any] struct {
	key   K
	value V
	prev  *lruCacheNode[K, V]
	next  *lruCacheNode[K, V]
}

// LRUCache represents the cache structure.
type LRUCache[K comparable, V any] struct {
	capacity int
	cache    map[K]*lruCacheNode[K, V]
	head     *lruCacheNode[K, V]
	tail     *lruCacheNode[K, V]
}

// NewLRUCache initializes a new LRUCache with a given capacity.
func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	lru := &LRUCache[K, V]{
		capacity: capacity,
		cache:    make(map[K]*lruCacheNode[K, V]),
		head:     &lruCacheNode[K, V]{},
		tail:     &lruCacheNode[K, V]{},
	}
	lru.head.next = lru.tail
	lru.tail.prev = lru.head
	return lru
}

// Get retrieves the value of the key if it exists in the cache, otherwise returns zero value and false.
func (this *LRUCache[K, V]) Get(key K) (V, bool) {
	if node, exists := this.cache[key]; exists {
		this.moveToHead(node)
		return node.value, true
	}
	var zero V
	return zero, false
}

// Put inserts or updates the value of the key.
// If the cache exceeds capacity, it removes the least recently used item.
func (this *LRUCache[K, V]) Put(key K, value V) {
	if node, exists := this.cache[key]; exists {
		node.value = value
		this.moveToHead(node)
	} else {
		node := &lruCacheNode[K, V]{
			key:   key,
			value: value,
		}
		this.cache[key] = node
		this.addNode(node)
		if len(this.cache) > this.capacity {
			tail := this.popTail()
			delete(this.cache, tail.key)
		}
	}
}

// Remove removes a node from the cache by key if it exists.
func (this *LRUCache[K, V]) Remove(key K) bool {
	if node, exists := this.cache[key]; exists {
		this.removeNode(node)
		delete(this.cache, key)
		return true
	}
	return false
}

// addNode adds a new node right after the head.
func (this *LRUCache[K, V]) addNode(node *lruCacheNode[K, V]) {
	node.prev = this.head
	node.next = this.head.next
	this.head.next.prev = node
	this.head.next = node
}

// removeNode removes an existing node from the linked list.
func (this *LRUCache[K, V]) removeNode(node *lruCacheNode[K, V]) {
	prev := node.prev
	next := node.next
	prev.next = next
	next.prev = prev
}

// moveToHead moves a node to the head of the linked list (most recently used).
func (this *LRUCache[K, V]) moveToHead(node *lruCacheNode[K, V]) {
	this.removeNode(node)
	this.addNode(node)
}

// popTail removes and returns the least recently used node.
func (this *LRUCache[K, V]) popTail() *lruCacheNode[K, V] {
	res := this.tail.prev
	this.removeNode(res)
	return res
}
