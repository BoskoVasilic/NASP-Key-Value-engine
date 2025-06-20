package structures

import "KVengine/structures/auxiliary"

type Entry struct {
	key   string
	value any
}

type Cache struct {
	maxSize int
	size    int
	dll     *auxiliary.DLList[Entry]
	hm      map[string]*auxiliary.Node[Entry]
}

func NewCache(maCashSize int) *Cache {
	return &Cache{
		size:    0,
		maxSize: maCashSize,
		dll:     auxiliary.NewDLList[Entry](),
		hm:      make(map[string]*auxiliary.Node[Entry]),
	}
}

func (c *Cache) Put(key string, value any) {
	node, ok := c.hm[key]
	if ok {
		c.dll.MoveToFront(node)
		return
	}
	entry := &Entry{key: key, value: value}
	node = c.dll.PushFront(entry)
	c.hm[key] = node
	if c.size >= c.maxSize {
		last, success := c.dll.PopBack()
		if success {
			delete(c.hm, last.key)
		}
	}
}

func (c *Cache) Get(key string) (any, bool) {
	node, ok := c.hm[key]
	if !ok {
		return nil, false
	}
	c.dll.MoveToFront(node)
	return node.Value, true
}
