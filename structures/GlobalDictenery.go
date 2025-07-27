package structures

import (
	"KVengine/structures/auxiliary"
	"KVengine/utils"
	"os"
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
	DicLock.Lock()
	g.file.FlushBuffer(false)
	DicLock.Unlock()
}
func AddItems[T any, H any](globalDictionary *GlobalDictionary, itterator *utils.Iterator[T, H, *Entry]) (*utils.Iterator[[]*ExportEntry, int, *ExportEntry], string, string) {
	// Treba da postoji flush opcija
	defer globalDictionary.file.FlushBuffer(false)
	list := auxiliary.NewListWithIterator[ExportEntry]()
	firsKey := ""
	lastKey := ""
	for {
		entry, ok := itterator.Next()
		if !ok {
			break
		}
		if firsKey == "" {
			firsKey = entry.key
		}
		lastKey = entry.key
		var newKey uint64
		DicLock.RLock()
		returnData, _ := globalDictionary.SkipList.Get(entry.key)
		DicLock.RUnlock()
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
	return list.GetIter(), firsKey, lastKey
}
func AddItemsWithNoCompresion[T any, H any](globalDictionary *GlobalDictionary, itterator *utils.Iterator[T, H, *Entry]) (*utils.Iterator[[]*Entry, int, *Entry], string, string) {
	// Treba da postoji flush opcija
	defer globalDictionary.file.FlushBuffer(false)
	list := auxiliary.NewListWithIterator[Entry]()
	firsKey := ""
	lastKey := ""
	for {
		entry, ok := itterator.Next()
		if !ok {
			break
		}
		if firsKey == "" {
			firsKey = entry.key
		}
		lastKey = entry.key
		DicLock.RLock()
		returnData, _ := globalDictionary.SkipList.Get(entry.key)
		DicLock.RUnlock()
		if returnData.key != entry.key {
			globalDictionary.AddItem(entry)
		}

		list.Add(entry)
		//fmt.Println("Adding item with key:", entry.key, "and new key:", newKey, "to the global dictionary.", globalDictionary.Size)
	}
	//Hear i need to flush document
	return list.GetIter(), firsKey, lastKey
}

func AddItmsString(globalDictionary *GlobalDictionary, key string) {

	defer globalDictionary.Flush()
	entry := &Entry{
		key:   key,
		value: utils.Uint64ToBytes(globalDictionary.Size),
	}
	DicLock.RLock()
	returnData, _ := globalDictionary.SkipList.Get(entry.key)
	DicLock.RUnlock()
	if returnData.key != entry.key {
		globalDictionary.AddItem(entry)
	}

}
func (GlobalDictionary *GlobalDictionary) AddItem(entry *Entry) uint64 {
	DicLock.Lock()
	buffer := make([]byte, 0)
	GlobalDictionary.SkipList.Put(entry.key, utils.Uint64ToBytes(GlobalDictionary.Size))
	buffer = append(buffer, utils.Uint64ToBytes(uint64(len(entry.key)))...)
	buffer = append(buffer, []byte(entry.key)...)
	buffer = append(buffer, utils.Uint64ToBytes(GlobalDictionary.Size)...)
	GlobalDictionary.file.WriteToFile(buffer)
	befor := GlobalDictionary.Size
	GlobalDictionary.Size = GlobalDictionary.Size + 1
	//fmt.Println("Added item with key:", befor, "to the global dictionary. New size is:", GlobalDictionary.Size)
	DicLock.Unlock()
	return befor
}
func (GlobalDictionary *GlobalDictionary) GetItem(key string) (Entry, bool) {
	DicLock.RLock()
	value, found := GlobalDictionary.SkipList.Get(key)
	DicLock.RUnlock()
	if !found {
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

	for {
		entry, ok := iterator.Next()
		if !ok {
			break
		}
		DicLock.RLock()
		newKeyBytes, found := GlobalDictionary.SkipList.Get(entry.key)
		DicLock.RUnlock()
		if !found {
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
func GetDicteneryFormFile(path string, bm *BlockManager) *GlobalDictionary {
	// Create a new BlockManager with the specified file path
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return NewGlobalDictionary(path, bm)
	}
	file := NewBlockManagerWrap(path, bm)
	start := 0
	end := file.GetLastWrittenByteOffset()
	dictionary := NewGlobalDictionary(path, bm)
	for start < end {
		_, keySizeBytes := file.ReadFromCurrentPosition(8)
		keySize := utils.BytesToUint64(keySizeBytes)
		_, keyBytes := file.ReadFromCurrentPosition(int(keySize))
		key := string(keyBytes)
		_, keyComprestedBytes := file.ReadFromCurrentPosition(8)
		keyComprested := utils.BytesToUint64(keyComprestedBytes)
		entry := NewEntry(key, utils.Uint64ToBytes(keyComprested), false)
		dictionary.SkipList.Put(entry.key, entry.value)
		dictionary.Size++
		start = start + 8 + int(keySize) + 8 // Move to the next entry
	}
	return dictionary
}
