package structures

import (
	"encoding/binary"
)

type Entry struct {
	key     string
	value   []byte
	deleted bool
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
	buf := make([]byte, 0)
	binary.AppendUvarint(buf, uint64(len(ent.key)))
	if ent.deleted {
		binary.AppendUvarint(buf, 0) // Value length is 0 for deleted entries
		buf = append(buf, []byte(ent.key)...)
	} else {
		binary.AppendUvarint(buf, uint64(len(ent.value)+1))
		buf = append(buf, []byte(ent.key)...)
		buf = append(buf, ent.value...)
	}
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
