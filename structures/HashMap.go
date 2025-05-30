package structures

type HashMap struct {
	buckets [][]entry
	size    int
}

type entry struct {
	key   string
	value interface{}
}

func NewHashMap(size int) *HashMap {
	return &HashMap{
		buckets: make([][]entry, size),
		size:    size,
	}
}

func (h *HashMap) hash(key string) int {
	hash := 0
	for i := 0; i < len(key); i++ {
		hash = (hash*31 + int(key[i])) % h.size
	}
	return hash
}

func (h *HashMap) Put(key string, value interface{}) {
	index := h.hash(key)
	for i, e := range h.buckets[index] {
		if e.key == key {
			h.buckets[index][i].value = value
			return
		}
	}
	h.buckets[index] = append(h.buckets[index], entry{key, value})
}

func (h *HashMap) Get(key string) (interface{}, bool) {
	index := h.hash(key)
	for _, e := range h.buckets[index] {
		if e.key == key {
			return e.value, true
		}
	}
	return nil, false
}

func (h *HashMap) Delete(key string) {
	index := h.hash(key)
	for i, e := range h.buckets[index] {
		if e.key == key {
			h.buckets[index] = append(h.buckets[index][:i], h.buckets[index][i+1:]...)
			return
		}
	}
}
