package structures

import utils "KVengine/utils"

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
