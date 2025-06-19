package structures

import (
	utils "KVengine/utils"
)

type HashMap struct {
	buckets [][]entry
	size    int
}

type entry struct {
	key     string
	value   []byte
	deleted bool
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

func (h *HashMap) Put(key string, value []byte) {
	index := h.hash(key)
	for i, e := range h.buckets[index] {
		if e.key == key {
			h.buckets[index][i].value = value
			return
		}
	}
	h.buckets[index] = append(h.buckets[index], entry{key, value, false})
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
	index := h.hash(key)
	for i, e := range h.buckets[index] {
		if e.key == key {
			h.buckets[index][i].value = []byte{}
			h.buckets[index][i].deleted = true
			return
		}
	}

}

func (h *HashMap) SerializeHM() []byte {
	buf := make([]byte, 0)

	sizeBytes := utils.Uint32ToBytes(uint32(h.size))
	buf = append(buf, sizeBytes...)

	for _, bucket := range h.buckets {
		bucketLen := utils.Uint32ToBytes(uint32(len(bucket)))
		buf = append(buf, bucketLen...)
		for _, e := range bucket {
			keyLen := utils.Uint32ToBytes(uint32(len(e.key)))
			buf = append(buf, keyLen...)
			buf = append(buf, []byte(e.key)...)

			valLen := utils.Uint32ToBytes(uint32(len(e.value)))
			buf = append(buf, valLen...)
			buf = append(buf, e.value...)
		}
	}

	return buf
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
