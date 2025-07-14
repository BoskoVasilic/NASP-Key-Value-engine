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
	path1 := tabel1.GenerateSSTabel("SSTable1", false)
	path2 := tabel2.GenerateSSTabel("SSTable2", false)
	//structures.ReadSStabel(path1, "key2", bm, dictenery)
	tree.AddLevel(path1)
	tree.AddLevel(path2)
	//entery, _ := tree.ReadSStabels("key3", bm, dictenery)
	fmt.Println("_________________________________________________________________________________")
	tree.StartMerging(1, dictenery, true, 4, 2)
	fmt.Println("Merging completed.")
	entery, _ := tree.ReadSStabels("key3", bm, dictenery)
	fmt.Println("Entry for key3 after merging:", entery)
	fmt.Println("_________________________________________________________________________________")
	tree2 := structures.Deserialize("tree_data", bm)
	fmt.Println("Deserialized tree:", tree2)
	fmt.Println("_________________________________________________________________________________")
	entry, _ := tree2.ReadSStabels("key3", bm, dictenery)
	fmt.Println("Entry for key3 in deserialized tree:", entry)
	dicnery2 := structures.GetDicteneryFormFile("test", bm)
	dicnery2.SkipList.Print()
}
