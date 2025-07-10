package structures

import (
	utils "KVengine/utils"
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
	buf := make([]byte, 0)

	keyLen := utils.Uint32ToBytes(uint32(len(ent.key)))
	buf = append(buf, keyLen...)

	if ent.deleted {
		buf = append(buf, utils.Uint32ToBytes(^uint32(0))...)
		buf = append(buf, []byte(ent.key)...)
	} else {
		valLen := utils.Uint32ToBytes(uint32(len(ent.value)))
		buf = append(buf, valLen...)
		buf = append(buf, []byte(ent.key)...)
		buf = append(buf, ent.value...)
	}
	return buf
}
