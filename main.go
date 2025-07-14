package main

import (
	"KVengine/structures"
	"fmt"
)

func main() {

	list1 := structures.NewSkipList()
	list2 := structures.NewSkipList()
	list1.Put("key2", []byte("value1"))
	list1.Put("key3", []byte("value2"))
	list2.Put("key1", []byte("value2"))
	list2.Put("key3", []byte("value3"))
	list2.Put("key4", []byte("value4"))
	bm := structures.NewBlockManager(100)
	dictenery := structures.NewGlobalDictionary("test", bm)
	tree := structures.NewLSMTree(3, bm)
	tabel1 := structures.NewSSTable(list1.GetIter(), 2, 4, true, bm, dictenery)
	tabel2 := structures.NewSSTable(list2.GetIter(), 3, 4, false, bm, dictenery)
	path1 := tabel1.GenerateSSTabel("SSTable1", true)
	path2 := tabel2.GenerateSSTabel("SSTable2", false)
	structures.ReadSStabel(path1, "key2", bm, dictenery)
	tree.AddLevel(path1)
	tree.AddLevel(path2)
	file := structures.NewBlockManagerWrap(path1, bm)
	file.SetCurrentPosition(58)
	entry := structures.ReadData(file)
	fmt.Println("Read entry:", entry)
	fmt.Println("Current position:", file.GetCurrentPosition())
	entry = structures.ReadData(file)
	fmt.Println("Read entry:", entry)
	fmt.Println("_______________________________________________________-----")
	tree.Merge(0, 2, 0, dictenery)

	//tree.Merge(0, 2, 0)

}
