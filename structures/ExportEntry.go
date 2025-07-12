package structures

import (
	"KVengine/hash"
	"encoding/binary"
	"time"
)

type ExportEntry struct {
	key       uint64
	value     []byte
	tombstone bool  // Indicates if the entry is a tombstone
	time      int64 // Time of creation or update
}

func NewExportEntryFromEntry(entry *Entry, key uint64) *ExportEntry {
	return &ExportEntry{
		key:       key,
		value:     entry.value,
		tombstone: entry.deleted,
		time:      entry.time,
	}
}
func NewExportEntry(key uint64, value []byte, toombstone bool) *ExportEntry {
	return &ExportEntry{
		key:       key,
		tombstone: toombstone,
		value:     value,
		time:      time.Now().UnixMicro(), // Time is not used in this context, but can be set if needed
	}
}
func (e ExportEntry) GetKey() uint64 {
	return e.key
}

func (e ExportEntry) GetValue() []byte {
	return e.value
}

func (e ExportEntry) IsDeleted() bool {
	return e.tombstone
}

func (ent *ExportEntry) toBytes() []byte {
	buf := make([]byte, 0)
	buf = binary.AppendUvarint(buf, ent.key) // Append the key as a varint)
	if ent.tombstone {
		buf = binary.AppendUvarint(buf, 0) // Value length is 0 for deleted entries
	} else {
		buf = binary.AppendUvarint(buf, uint64(len(ent.value)+1))
		buf = append(buf, ent.value...)
	}
	buf = binary.AppendUvarint(buf, uint64(ent.time)) // Append the time as a varint
	hash := hash.CreateHashFunctions(42)[0]
	CRU := hash.Hash(buf)
	CRU = CRU % 1000000007 // Modulo operation to fit the hash into a smaller range
	bufCru := make([]byte, 0)
	bufCru = binary.AppendUvarint(bufCru, CRU) // Append the length of the CRU hash
	// Append the CRU hash as a varint
	//fmt.Println("CRU hash:", bufCru)
	buf = append(bufCru, buf...)
	//fmt.Println("ExportEntry toBytes:", buf)
	return buf
}
func (ent *ExportEntry) GetSize() uint64 {
	if ent.tombstone {
		// Alociraj bajtove za varint (max 10 bajtova)
		buf := make([]byte, 0)
		value_len := 0
		buf = binary.AppendUvarint(buf, ent.key)
		buf = binary.AppendUvarint(buf, uint64(value_len))
		return uint64(len(buf))

	}
	value_len := len(ent.value)
	buf := make([]byte, 0)
	buf = binary.AppendUvarint(buf, ent.key) // Append the key as a varint
	buf = binary.AppendUvarint(buf, uint64(value_len+1))
	return uint64(len(buf) + value_len) //TODO dodati len od CRU i TimeStamp
}
func comper(a, b *ExportEntry) bool {
	return a.key < b.key
}
