package structures

import "KVengine/structures/auxiliary"

type CacheEntry[K comparable, V any] struct {
	key     K
	value   V
	deleted bool
}

type Cache[K comparable, V any] struct {
	maxSize int
	size    int
	dll     *auxiliary.DLList[CacheEntry[K, V]]
	hm      map[K]*auxiliary.Node[CacheEntry[K, V]]
}

func NewCache[K comparable, V any](maCashSize int) *Cache[K, V] {
	return &Cache[K, V]{
		size:    0,
		maxSize: maCashSize,
		dll:     auxiliary.NewDLList[CacheEntry[K, V]](),
		hm:      make(map[K]*auxiliary.Node[CacheEntry[K, V]]),
	}
}

func (c *Cache[K, V]) Update(key K, value V) {
	node, ok := c.hm[key]
	if ok {
		node.Value.value = value
		node.Value.deleted = false
		c.dll.MoveToFront(node)
	}
}

func (c *Cache[K, V]) Put(key K, value V) {
	node, ok := c.hm[key]
	if ok {
		c.dll.MoveToFront(node)
		return
	}
	entry := &CacheEntry[K, V]{key: key, value: value, deleted: false}

	node = c.dll.PushFront(entry)
	c.hm[key] = node
	if c.size >= c.maxSize {
		last, success := c.dll.PopBack()
		if success {
			delete(c.hm, last.key)
		}
	} else {
		c.size++
	}
}

func (c *Cache[K, V]) Delete(key K) {
	node, ok := c.hm[key]
	if !ok {
		return
	}
	node.Value.deleted = true
	c.dll.MoveToFront(node)
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	var zero V
	node, ok := c.hm[key]
	if !ok {
		return zero, false
	}
	c.dll.MoveToFront(node)
	return node.Value.value, true
}
