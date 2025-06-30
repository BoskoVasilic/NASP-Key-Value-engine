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
