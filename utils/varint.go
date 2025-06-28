package utils

func EncodeVarint(x uint64) []byte {
	var buf []byte
	for x >= 0x80 {
		// Store 7 bits of x, set high bit to 1 (more bytes follow)
		buf = append(buf, byte(x)|0x80)
		x >>= 7 // shift out the 7 bits
	}
	// Final byte: high bit 0 (marks end)
	buf = append(buf, byte(x))
	return buf
}

func DecodeVarint(buf []byte) (uint64, int) {
	var x uint64
	var shift uint
	for i, b := range buf {
		x |= uint64(b&0x7F) << shift
		if b&0x80 == 0 {
			return x, i + 1 // value and bytes consumed
		}
		shift += 7
	}
	return 0, 0 // incomplete encoding
}
