package structures

import (
	"KVengine/structures/auxiliary"
	"KVengine/utils"
)

type GlobalDictionary struct {
	SkipList *SkipList         // SkipList is a data structure that allows for fast search, insertion, and deletion
	Size     uint64            // Size is the number of entries in the dictionary
	file     *BlockManagerWrap // BlockManager is responsible for managing blocks of data in the file
}

func NewGlobalDictionary(filePath string, blocMenager *BlockManager) *GlobalDictionary {
	return &GlobalDictionary{
		SkipList: NewSkipList(),
		Size:     1, // Start with 1 to avoid zero key
		file:     NewBlockManagerWrap(filePath, blocMenager),
	}
}
func (g GlobalDictionary) Flush() {
	// Flush the file buffer to ensure all data is written to disk
	g.file.FlushBuffer(false)
}
func AddItems[T any, H any](globalDictionary *GlobalDictionary, itterator *utils.Iterator[T, H, *Entry]) *utils.Iterator[[]*ExportEntry, int, *ExportEntry] {
	// Treba da postoji flush opcija
	defer globalDictionary.file.FlushBuffer(false)
	list := auxiliary.NewListWithIterator[ExportEntry]()
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
		//fmt.Println("Adding item with key:", entry.key, "and new key:", newKey, "to the global dictionary.", globalDictionary.Size)
	}
	list.Sort(func(a, b *ExportEntry) bool {
		return a.key < b.key
	})
	//Hear i need to flush document
	return list.GetIter()
}
func AddItemsWithNoCompresion[T any, H any](globalDictionary *GlobalDictionary, itterator *utils.Iterator[T, H, *Entry]) *utils.Iterator[[]*Entry, int, *Entry] {
	// Treba da postoji flush opcija
	defer globalDictionary.file.FlushBuffer(false)
	list := auxiliary.NewListWithIterator[Entry]()
	for true {
		entry, ok := itterator.Next()
		if !ok {
			break
		}

		returnData, _ := globalDictionary.SkipList.Get(entry.key)
		if returnData.key != entry.key {
			globalDictionary.AddItem(entry)
		}

		list.Add(entry)
		//fmt.Println("Adding item with key:", entry.key, "and new key:", newKey, "to the global dictionary.", globalDictionary.Size)
	}
	//Hear i need to flush document
	return list.GetIter()
}

func AddItmsString(globalDictionary *GlobalDictionary, key string) {
	entry := &Entry{
		key:   key,
		value: utils.Uint64ToBytes(globalDictionary.Size),
	}
	returnData, _ := globalDictionary.SkipList.Get(entry.key)
	if returnData.key != entry.key {
		globalDictionary.AddItem(entry)
	}

}
func (GlobalDictionary *GlobalDictionary) AddItem(entry *Entry) uint64 {
	buffer := make([]byte, 0)
	GlobalDictionary.SkipList.Put(entry.key, utils.Uint64ToBytes(GlobalDictionary.Size))
	buffer = append(buffer, utils.Uint64ToBytes(uint64(len(entry.key)))...)
	buffer = append(buffer, []byte(entry.key)...)
	buffer = append(buffer, utils.Uint64ToBytes(GlobalDictionary.Size)...)
	GlobalDictionary.file.WriteToFile(buffer)
	befor := GlobalDictionary.Size
	GlobalDictionary.Size = GlobalDictionary.Size + 1
	//fmt.Println("Added item with key:", befor, "to the global dictionary. New size is:", GlobalDictionary.Size)
	return befor
}
func (GlobalDictionary *GlobalDictionary) GetItem(key string) (Entry, bool) {
	value, ok := GlobalDictionary.SkipList.Get(key)
	if ok != nil {
		return Entry{}, false
	}
	if value.key != key {
		return Entry{}, false
	}
	return value, true
}

// ConvertListToExportEntry converts a ListWithIterator of Entry to a ListWithIterator of ExportEntry.
func (GlobalDictionary *GlobalDictionary) ConvertListToExportEntry(list *auxiliary.ListWithIterator[Entry]) *auxiliary.ListWithIterator[ExportEntry] {
	// Implement the conversion logic here
	iterator := list.GetIter()
	list2 := auxiliary.NewListWithIterator[ExportEntry]()

	for true {
		entry, ok := iterator.Next()
		if !ok {
			break
		}
		newKeyBytes, found := GlobalDictionary.SkipList.Get(entry.key)
		if found != nil {
			// If the key is not found in the SkipList, we can skip this entry
			continue
		}
		if newKeyBytes.key == entry.key {
			newKey := utils.BytesToUint64(newKeyBytes.value)
			exportEntry := NewExportEntryFromEntry(entry, newKey)
			list2.Add(exportEntry)
		}

	}
	list2.Sort(func(a, b *ExportEntry) bool {
		return a.key < b.key
	})
	return list2
}
