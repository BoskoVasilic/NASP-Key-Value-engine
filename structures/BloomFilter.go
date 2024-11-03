package structures

import "hahs_with_seed/E:\NASP-Key-Value-engine\hash.go"

type DBF struct {
	NumberOfFunctions uint32
	DelLen            int32
	MainLen           uint32
	ArrayDel          []byte
	ArrayMain         []byte
	HashFunc          []HashWithSeed
}
