package main

import (
	"KVengine/structures"
	"fmt"
)

func main() {
	fmt.Println("Hello, World!")
	filter := structures.NewBloomFilterWithParams(5, 5)
	fmt.Println(filter.NumberOfFunctions)
	filter.Put([]byte("Hello"))
	filter.Put([]byte("World"))
	fmt.Println(filter.Check([]byte("Hello")))
	fmt.Println(filter.Check([]byte("World")))
	fmt.Println(filter.Check([]byte("Hello, World!")))
	fmt.Println(filter.ArrayMain)
	fmt.Println(filter.SerializeBloomF())
	data := filter.SerializeBloomF()
	filter2 := structures.DeserializeBloomF(data)
	fmt.Println(filter2.ArrayMain)
	data2 := filter2.SerializeBloomF()
	fmt.Println(filter2.Check([]byte("Hello")))
	fmt.Println(filter2.Check([]byte("World")))
	fmt.Println(filter2.Check([]byte("Hello, World!")))
	fmt.Println(data2)

	structures.TestSimHashAndHemmingwayDistance()
}
