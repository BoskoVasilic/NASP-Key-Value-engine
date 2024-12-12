package main

import (
	"KVengine/structures"
	"fmt"
)

func main() {
	fmt.Println("Hello, World!")
	filter := structures.NewBloomFilterWithParams(1000, 5)
	fmt.Println(filter.NumberOfFunctions)
	filter.Put([]byte("Hello"))
	filter.Put([]byte("World"))
	fmt.Println(filter.Check([]byte("Hello")))
	fmt.Println(filter.Check([]byte("World")))
	fmt.Println(filter.Check([]byte("Hello, World!")))
	fmt.Println(filter.ArrayMain)
}
