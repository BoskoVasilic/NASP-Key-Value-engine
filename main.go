package main



func main() {
	fmt.Println("Hello, World!")
	//filter := structures.NewBloomFilterWithParams(5, 5)
	//fmt.Println(filter.NumberOfFunctions)
	//filter.Put([]byte("Hello"))
	//filter.Put([]byte("World"))
	//fmt.Println(filter.Check([]byte("Hello")))
	//fmt.Println(filter.Check([]byte("World")))
	//fmt.Println(filter.Check([]byte("Hello, World!")))
	//fmt.Println(filter.ArrayMain)
	//fmt.Println(filter.Serialize())
	//data := filter.Serialize()
	//filter2 := structures.DeSerialize(data)
	//fmt.Println(filter2.ArrayMain)
	//data2 := filter2.Serialize()
	//fmt.Println(filter2.Check([]byte("Hello")))
	//fmt.Println(filter2.Check([]byte("World")))
	//fmt.Println(filter2.Check([]byte("Hello, World!")))
	//fmt.Println(data2)

	element := structures.NewWriteAheadLogElement(5, 5, []byte("Hello"), []byte("World"))
	fmt.Println(element.Serialize())
	element2 := structures.NewWriteAheadLogElement(7, 8, []byte("Hello00"), []byte("World112"))
	fmt.Println(element2.Serialize())
	elementSer := element.Serialize()
	fmt.Println(structures.Deserialize(elementSer))
	log := structures.NewWriteAheadLog()
	fmt.Println(log)
	log.AddElement(element)
	log.AddElement(element2)
	fmt.Println(log)
	log.ReadFromFile()
	fmt.Println(log)
}
