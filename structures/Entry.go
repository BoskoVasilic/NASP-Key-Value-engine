package structures

import (
	"KVengine/hash"
	"encoding/binary"
	"fmt"
	"time"
)

type Entry struct {
	key     string
	value   []byte
	time    int64
	deleted bool
}

func NewEntry(key string, value []byte, deleted bool) *Entry {
	return &Entry{key, value, time.Now().UnixNano(), deleted}
}

func NewEntryWithTime(key string, value []byte, time int64, deleted bool) *Entry {
	return &Entry{key, value, time, deleted}
}

func (e Entry) GetKey() string {
	return e.key
}

func (e Entry) GetValue() []byte {
	return e.value
}

func (e Entry) IsDeleted() bool {
	return e.deleted
}

func (ent *Entry) toBytes() []byte {
	fmt.Println("Entry toBytes called" + ent.key)
	buf := make([]byte, 0)
	buf = binary.AppendUvarint(buf, uint64(len(ent.key)))
	fmt.Println("Key length: ", len(ent.key))
	fmt.Println("Key in bytes: ", []byte(ent.key))
	if ent.deleted {
		buf = binary.AppendUvarint(buf, 0) // Value length is 0 for deleted entries
		buf = append(buf, []byte(ent.key)...)
	} else {
		buf = binary.AppendUvarint(buf, uint64(len(ent.value)+1))
		buf = append(buf, []byte(ent.key)...)
		buf = append(buf, ent.value...)
	}
	fmt.Println("Key: ", buf)
	buf = binary.AppendUvarint(buf, uint64(ent.time))
	hash := hash.CreateHashFunctions(42)[0]
	fmt.Println("Buffer: ", buf)
	CRU := hash.Hash(buf)
	CRU = CRU % 1000000007 // Modulo operation to fit the hash into a smaller range
	bufCru := make([]byte, 0)
	bufCru = binary.AppendUvarint(bufCru, CRU) // Append the length of the CRU hash
	// Append the CRU hash as a varint
	fmt.Println("CRU hash:", bufCru)
	buf = append(bufCru, buf...)
	fmt.Println("Entry toBytes:", buf)
	return buf
}
func (ent *Entry) GetSize() uint64 {
	if ent.deleted {
		// Alociraj bajtove za varint (max 10 bajtova)
		buf := make([]byte, 0)
		key_len := len(ent.key)
		value_len := 0
		buf = binary.AppendUvarint(buf, uint64(key_len))
		buf = binary.AppendUvarint(buf, uint64(value_len))
		return uint64(len(buf) + key_len)

	}
	key_len := len(ent.key)
	value_len := len(ent.value)
	buf := make([]byte, 0)
	buf = binary.AppendUvarint(buf, uint64(key_len))
	buf = binary.AppendUvarint(buf, uint64(value_len+1))
	return uint64(len(buf) + key_len + value_len) //TODO dodati len od CRU i TimeStamp
}
