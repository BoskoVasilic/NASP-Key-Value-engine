package structures

import (
	utils "KVengine/utils"
	"encoding/binary"
	"sort"
)

type ExportEntry struct {
	key       uint64
	value     []byte
	tombstone bool // Indicates if the entry is a tombstone
}

func NewExportEntryFromEntry(entry *Entry, key uint64) *ExportEntry {
	return &ExportEntry{
		key:       key,
		value:     entry.value,
		tombstone: entry.deleted,
	}
}
func NewExportEntry(key uint64, value []byte, toombstone bool) *ExportEntry {
	return &ExportEntry{
		key:       key,
		tombstone: toombstone,
		value:     value,
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
	binary.AppendUvarint(buf, ent.key) // Append the key as a varint)
	if ent.tombstone {
		binary.AppendUvarint(buf, 0) // Value length is 0 for deleted entries
	} else {
		binary.AppendUvarint(buf, uint64(len(ent.value)+1))
		buf = append(buf, ent.value...)
	}
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

type ListWithIterator[T any] struct {
	entries []*T // Slice of ExportEntry pointers
}

func NewListWithIterator[T any]() *ListWithIterator[T] {
	return &ListWithIterator[T]{
		entries: make([]*T, 0),
	}

}
func (list *ListWithIterator[T]) Add(entry *T) {
	list.entries = append(list.entries, entry)
}

func (list *ListWithIterator[T]) Sort(compare func(a, b *T) bool) {
	sort.Slice(list.entries, func(i, j int) bool {
		return compare(list.entries[i], list.entries[j])
	})
}

func (list *ListWithIterator[T]) GetIter() *utils.Iterator[[]*T, *T] {
	return utils.NewIterator(&list.entries, 0, func(entries *[]*T, pos int) (*T, bool) {
		if pos >= len(*entries) {
			return nil, false
		}
		pos = pos + 1
		return (*entries)[pos-1], true
	})
}
