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
