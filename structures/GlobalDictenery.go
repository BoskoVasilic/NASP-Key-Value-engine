package structures

import (
	"KVengine/utils"
	"fmt"
)

type GlobalDictionary struct {
	SkipList *SkipList         // SkipList is a data structure that allows for fast search, insertion, and deletion
	Size     uint64            // Size is the number of entries in the dictionaryMerge branch 'SSTable' into Develop
	file     *BlockManagerWrap // BlockManager is responsible for managing blocks of data in the file
}

func NewGlobalDictionary(filePath string, blocMenager *BlockManager) *GlobalDictionary {
	return &GlobalDictionary{
		SkipList: NewSkipList(),
		Size:     0,
		file:     NewBlockManagerWrap(filePath, blocMenager),
	}
}
func AddItems[T any](globalDictionary *GlobalDictionary, itterator *utils.Iterator[T, *Entry]) *utils.Iterator[[]*ExportEntry, *ExportEntry] {
	// Treba da postoji flush opcija
	defer globalDictionary.file.FlushBuffer()
	list := NewListWithIterator[ExportEntry]()
	for true {
		entry, ok := itterator.Next()
		if !ok {
			break
		}

		var newKey uint64
		returnData, _ := globalDictionary.SkipList.Get(entry.key)
		if returnData.key != entry.key {
			newKey = globalDictionary.AddItem(entry)
		} else {
			newKey = utils.BytesToUint64(returnData.value)
		}

		list.Add(NewExportEntryFromEntry(entry, newKey))
		fmt.Println("Adding item with key:", entry.key, "and new key:", newKey, "to the global dictionary.", globalDictionary.Size)
	}
	list.Sort(func(a, b *ExportEntry) bool {
		return a.key < b.key
	})
	//Hear i need to flush document
	return list.GetIter()
}
func (GlobalDictionary *GlobalDictionary) AddItem(entry *Entry) uint64 {
	buffer := make([]byte, 0)
	GlobalDictionary.SkipList.Put(entry.key, utils.Uint64ToBytes(GlobalDictionary.Size), false)
	buffer = append(buffer, utils.Uint64ToBytes(uint64(len(entry.key)))...)
	buffer = append(buffer, []byte(entry.key)...)
	buffer = append(buffer, utils.Uint64ToBytes(GlobalDictionary.Size)...)
	GlobalDictionary.file.WriteToFile(buffer)
	befor := GlobalDictionary.Size
	GlobalDictionary.Size = GlobalDictionary.Size + 1
	fmt.Println("Added item with key:", befor, "to the global dictionary. New size is:", GlobalDictionary.Size)
	return befor
}
func (GlobalDictionary GlobalDictionary) GetItem(key string) (Entry, bool) {
	value, ok := GlobalDictionary.SkipList.Get(key)
	if ok != nil {
		return Entry{}, false
	}
	return value, true
}
