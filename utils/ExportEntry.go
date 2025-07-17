package utils

type ExportEntry struct {
	Key       string
	ValueSize uint64
	Value     []byte
}

func NewExportEntry(key string, valueSize uint64, value []byte) *ExportEntry {
	return &ExportEntry{
		Key:       key,
		ValueSize: valueSize,
		Value:     value,
	}
}
func (e *ExportEntry) GetBytes() []byte {

	keyLength := uint64(len(e.Key))
	bytes := Uint64ToBytes(e.ValueSize)
	bytes = append(bytes, Uint64ToBytes(keyLength)...)
	bytes = append(bytes, []byte(e.Key)...)
	bytes = append(bytes, e.Value...)
	return bytes
}
