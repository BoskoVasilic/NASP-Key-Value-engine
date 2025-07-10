package structures

import (
	utils "KVengine/utils"
	"sort"
)

type HashMap struct {
	buckets [][]Entry
	size    int
}

func NewHashMap(size int) *HashMap {
	return &HashMap{
		buckets: make([][]Entry, size),
		size:    size,
	}
}

func (h *HashMap) hash(key string) int {
	hash := 0
	for i := range len(key) {
		hash = (hash*31 + int(key[i])) % h.size
	}
	return hash
}

func (h *HashMap) insert(key string, value []byte, deleted bool) {
	index := h.hash(key)
	for i, e := range h.buckets[index] {
		if e.key == key {
			h.buckets[index][i].value = value
			h.buckets[index][i].deleted = deleted
			return
		}
	}
	h.buckets[index] = append(h.buckets[index], *NewEntry(key, value, deleted))
}

func (h *HashMap) Put(key string, value []byte) {
	h.insert(key, value, false)
}

func (h *HashMap) Get(key string) ([]byte, bool) {
	index := h.hash(key)
	for _, e := range h.buckets[index] {
		if e.key == key {
			return e.value, true
		}
	}
	return nil, false
}

func (h *HashMap) Delete(key string) {
	h.insert(key, nil, true)

}

func (h *HashMap) SerializeHM() []byte {
	buf := make([]byte, 0)

	sizeBytes := utils.Uint32ToBytes(uint32(h.size))
	buf = append(buf, sizeBytes...)

	for _, bucket := range h.buckets {
		bucketLen := utils.Uint32ToBytes(uint32(len(bucket)))
		buf = append(buf, bucketLen...)
		for _, e := range bucket {

			buf = append(buf, e.toBytes()...)
			/* keyLen := utils.Uint32ToBytes(uint32(len(e.key)))
			buf = append(buf, keyLen...)
			buf = append(buf, []byte(e.key)...)

			valLen := utils.Uint32ToBytes(uint32(len(e.value)))
			buf = append(buf, valLen...)
			buf = append(buf, e.value...) */
		}
	}

	return buf
}

func (h *HashMap) ToBytes() []byte {
	buf := make([]byte, 0)

	entries := make([]*Entry, 0)

	sizeBytes := utils.Uint32ToBytes(uint32(h.size))
	buf = append(buf, sizeBytes...)

	for _, bucket := range h.buckets {
		for _, e := range bucket {
			entries = append(entries, &e)
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].key < entries[j].key
	})
	for i := range entries {
		buf = append(buf, entries[i].toBytes()...)
	}
	return buf
}

func (h *HashMap) GetIter() *utils.Iterator[[]*Entry, []byte] {

	entries := make([]*Entry, 0)

	for _, bucket := range h.buckets {
		for _, e := range bucket {
			entries = append(entries, &e)
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].key < entries[j].key
	})

	return utils.NewIterator(&entries, 0, func(obj *[]*Entry, i int) ([]byte, bool) {
		if i >= len(*obj) {
			return nil, false
		}
		return (*obj)[i].toBytes(), true

	})

}

func DeserializeHM(data []byte) *HashMap {
	offset := 0
	size := utils.BytesToUint32(data[offset : offset+4])
	offset += 4

	h := NewHashMap(int(size))

	for range int(size) {
		bucketLen := utils.BytesToUint32(data[offset : offset+4])
		offset += 4

		for range int(bucketLen) {
			keyLen := utils.BytesToUint32(data[offset : offset+4])
			offset += 4
			key := string(data[offset : offset+int(keyLen)])
			offset += int(keyLen)

			valLen := utils.BytesToUint32(data[offset : offset+4])
			offset += 4
			val := data[offset : offset+int(valLen)]
			offset += int(valLen)

			h.Put(key, val)
		}
	}

	return h
}
