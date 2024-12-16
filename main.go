package main

import (
	"KVengine/structures"
	"fmt"
)

func main() {
	s := structures.NewSkipList()
	for i := 0; s.GetHeight() < 3; i++ {
		str := fmt.Sprintf("%5d", i)
		s.Insert(str, []byte(str))
	}
	s.Print()
	return
	s.Insert("a", []byte("1"))
	s.Insert("b", []byte("1"))
	s.Insert("c", []byte("1"))
	s.Insert("d", []byte("1"))
	s.Insert("e", []byte("1"))
}
