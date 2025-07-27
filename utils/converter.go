package utils

func Uint32ToBytes(number uint32) []byte {
	b := make([]byte, 4)
	b[3] = byte(number)
	b[2] = byte(number >> 8)
	b[1] = byte(number >> 16)
	b[0] = byte(number >> 24)
	return b
}
func BytesToUint32(b []byte) uint32 {
	return uint32(b[3]) + uint32(b[2])<<8 + uint32(b[1])<<16 + uint32(b[0])<<24
}
func Uint64ToBytes(number uint64) []byte {
	b := make([]byte, 8)
	b[7] = byte(number)
	b[6] = byte(number >> 8)
	b[5] = byte(number >> 16)
	b[4] = byte(number >> 24)
	b[3] = byte(number >> 32)
	b[2] = byte(number >> 40)
	b[1] = byte(number >> 48)
	b[0] = byte(number >> 56)
	return b
}
func BytesToUint64(b []byte) uint64 {
	return uint64(b[7]) + uint64(b[6])<<8 + uint64(b[5])<<16 + uint64(b[4])<<24 + uint64(b[3])<<32 + uint64(b[2])<<40 + uint64(b[1])<<48 + uint64(b[0])<<56
}
