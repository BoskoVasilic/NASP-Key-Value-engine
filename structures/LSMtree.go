package structures

import (
	"KVengine/utils"
	"fmt"
	"os"
	"time"
)

// tree_data hardcoded path to LSM tabel
type LSMTree struct {
	maxLevel     int
	currentLevel int
	LevelsSize   []int
	Levels       []*[]*string
	BlockMenager *BlockManager
	// Levels is a slice of pointers to slices of strings, representing the levels of the LSM tree.
	// Each level contains a slice of strings, which can be thought of as key-value pairs or records.
}

func NewLSMTree(maxLevel int, BlockMenager *BlockManager) *LSMTree {
	return &LSMTree{
		Levels:       make([]*[]*string, 0),
		currentLevel: 0,
		maxLevel:     maxLevel,
		LevelsSize:   make([]int, maxLevel),
		BlockMenager: BlockMenager,
		// LevelsSize is a slice that keeps track of the size of each level in the LSM tree.
		// It is initialized to a slice of integers with length equal to maxLevel, all set to zero.
		// Initialize the LSM tree with an empty slice of levels.
	}
}
func (lt *LSMTree) AddLevel(path string) {
	if len(lt.Levels) == 0 {
		// If the Levels slice is empty, initialize it with a new slice.
		lt.Levels = append(lt.Levels, &[]*string{})
	}
	level1 := lt.Levels[0]
	PathStringAdded := []*string{&path}
	*level1 = append(PathStringAdded, *level1...)
	lt.LevelsSize[0]++
	lt.SerializeLSM(GlobalLSMTreeFile, lt.BlockMenager)
}
func (lt *LSMTree) Merge(first int, last int, level int, dictenery *GlobalDictionary, oneFile bool, summeryIndex uint64) string {
	DataCompresedLists := make([]bool, 0)
	MaxSizes := make([]int, 0)
	OneFileLise := make([]bool, 0)
	levleSData := *(lt.Levels[level])
	last = last + 1
	ForceCompresion := false
	firstKey := ""
	lastKey := ""
	for i := first; i < last; i++ {
		file := NewBlockManagerWrap(*levleSData[i], lt.BlockMenager)
		_, IsCompresed := file.ReadFromCurrentPosition(1)
		_, OneFile := file.ReadFromCurrentPosition(1)
		_, data := file.ReadFromCurrentPosition(8)
		maxSizeC := utils.BytesToUint64(data)
		file.SetCurrentPosition(0)
		stringKeyFirst, stringKeyLast := GetFirstAndLAstKeyCompresed(*levleSData[i], lt.BlockMenager)
		if firstKey == "" {
			firstKey = stringKeyFirst
		}
		if lastKey == "" {
			lastKey = stringKeyLast
		}
		if firstKey > stringKeyFirst {
			firstKey = stringKeyFirst
		}
		if lastKey < stringKeyLast {
			lastKey = stringKeyLast
		}
		if IsCompresed[0] == 1 {
			DataCompresedLists = append(DataCompresedLists, true)
			ForceCompresion = true
		} else {
			DataCompresedLists = append(DataCompresedLists, false)
		}
		if OneFile[0] == 1 {
			OneFileLise = append(OneFileLise, true)
		} else {
			OneFileLise = append(OneFileLise, false)
		}
		MaxSizes = append(MaxSizes, int(maxSizeC))
	}

	if ForceCompresion {
		// If any of the files are compressed, we need to compres all

		pq := NewPriorityQueue(func(a, b *MergeData[ExportEntry]) bool {
			if a.GetRecord().GetKey() == b.GetRecord().GetKey() {
				return a.GetRecord().time > b.GetRecord().time
			}
			return a.GetRecord().GetKey() < b.GetRecord().GetKey()
		})
		for i := first; i < last; i++ {
			if DataCompresedLists[i-first] {
				file := GetDataFile(*levleSData[i], lt.BlockMenager)
				mergaData := NewMergeData[ExportEntry](file, nil, func(mergeData *MergeData[ExportEntry]) {
					if mergeData.Read >= mergeData.maxSize {
						//fmt.Println("Reading more than max size")
						//fmt.Println("File name:", mergeData.Read, "Max size:", mergeData.maxSize)
						mergeData.SetRecord(&ExportEntry{
							key:       0,
							value:     nil,
							tombstone: false,
							time:      0,
						})
						mergeData.Read++
						return
					}
					entry := ReadDataCompressed(mergeData.fileReader)
					mergeData.SetRecord(entry)
					mergeData.Read++
				}, MaxSizes[i-first])
				mergaData.function(mergaData)
				pq.Push(mergaData)
				//fmt.Println("Priority Queue Data", pq.data)
			} else {
				offset := 0
				if OneFileLise[i-first] {
					offset = 58
				}
				file := GetDataFile(*levleSData[i], lt.BlockMenager)
				fileIndex := GetIndexFile(*levleSData[i], lt.BlockMenager)
				list_Enterys := GetListInex(fileIndex, uint64(MaxSizes[i-first]))
				list_ExportEntries := dictenery.ConvertListToExportEntry(list_Enterys)
				mergaData := NewMergeData[ExportEntry](file, list_ExportEntries, func(mergeData *MergeData[ExportEntry]) {
					if mergeData.Read >= mergeData.maxSize {
						//fmt.Println("Reading more than max size")
						//fmt.Println("File name:", mergeData.Read, "Max size:", mergeData.maxSize)
						mergeData.SetRecord(&ExportEntry{
							key:       0,
							value:     nil,
							tombstone: false,
							time:      0,
						})
						mergeData.Read++
						return
					}
					EntryEnters := list_ExportEntries.GetEntrie(mergeData.Read)
					offsetData := utils.BytesToUint64(EntryEnters.value) + uint64(offset)
					mergeData.fileReader.SetCurrentPosition(int(offsetData))
					entry := ReadData(mergeData.fileReader)
					ExportEntry := NewExportEntryFromEntry(entry, EntryEnters.key)
					mergeData.SetRecord(ExportEntry)
					mergeData.Read++
				}, MaxSizes[i-first])
				mergaData.function(mergaData)
				pq.Push(mergaData)
			}
		}
		noKey := uint64(0)
		iter := utils.NewIterator(pq, &noKey, func(pq *PriorityQueue[MergeData[ExportEntry]], s *uint64) (*ExportEntry, bool) {
			if pq == nil || pq.Len() == 0 {
				return nil, false
			}
			top := pq.Peek()
			pq.Pop()
			//fmt.Println("Read value", top.Read)
			//fmt.Println("Top Record:", top.GetRecord().GetKey(), "Value:", top.GetRecord().value, "Deleted:", top.GetRecord().deleted, "Time:", top.GetRecord().time)
			for top.Read > top.maxSize {
				// Handling no more data anywhere
				if pq.Len() == 0 {
					return nil, false
				}
				top = pq.Peek()
				pq.Pop()
			}
			//fmt.Println("AFter for loop", top.Read)
			if top == nil {
				return nil, false
			}
			record := top.GetRecord()
			//fmt.Println("FIRST record past")
			if record == nil {
				return nil, false
			}
			//fmt.Println("Record before entering:", record.GetKey(), "Value:", record.value, "Deleted:", record.deleted, "Time:", record.time)
			for record.GetKey() == *s {
				top.function(top)
				//fmt.Println("Record after function call:", record.GetKey(), "Value:", record.value, "Deleted:", record.deleted, "Time:", record.time)
				//fmt.Println("Top Record:", top.GetRecord().GetKey(), "Value:", top.GetRecord().value, "Deleted:", top.GetRecord().deleted, "Time:", top.GetRecord().time)
				pq.Push(top)
				top = pq.Peek()
				pq.Pop()
				for top.Read > top.maxSize {
					//Handling no more data enywhere
					if pq.Len() == 0 {
						return nil, false
					}
					top = pq.Peek()
					pq.Pop()
				}
				record = top.GetRecord()
			}
			*s = record.GetKey()
			recordCopy := &ExportEntry{
				key:       record.GetKey(),
				value:     record.value,
				tombstone: record.tombstone,
				time:      record.time,
			}
			// Create a copy of the record to return
			// This is to avoid returning a pointer to the original record
			//fmt.Println("Returning Record:", recordCopy.GetKey(), "Value:", recordCopy.value, "Deleted:", recordCopy.deleted, "Time:", recordCopy.time)
			top.function(top)
			pq.Push(top)
			return recordCopy, true
		})
		startpath := "./data/SSTabel" + fmt.Sprintf("%d", time.Now().UnixMicro())

		sum := 0
		for i := first; i < last; i++ {
			sum += MaxSizes[i-first]
		}
		Fileter := NewBloomFilterWithParams(sum, 0.01)
		tree := NewMerkleTree()
		path1 := WriteToFilesCompresed(startpath, oneFile, lt.BlockMenager, sum, Fileter, summeryIndex, iter, firstKey, lastKey, tree)
		lt.AddtoNextLevel(level+1, path1)
		return path1
	} else {
		pq := NewPriorityQueue(func(a, b *MergeData[Entry]) bool {
			if a.GetRecord().GetKey() == b.GetRecord().GetKey() {
				return a.GetRecord().time > b.GetRecord().time

			}
			return a.GetRecord().GetKey() < b.GetRecord().GetKey()
		})
		// We now that all files are not compresed :)
		for i := first; i < last; i++ { // Make a function that returns file with data from my path
			file := GetDataFile(*levleSData[i], lt.BlockMenager)
			mergaData := NewMergeData[Entry](file, nil, func(mergeData *MergeData[Entry]) {
				if mergeData.Read >= mergeData.maxSize {
					//fmt.Println("Reading more than max size")
					//fmt.Println("File name:", mergeData.Read, "Max size:", mergeData.maxSize)
					mergeData.SetRecord(&Entry{
						key:     "",
						value:   nil,
						deleted: false,
						time:    0,
					})
					mergeData.Read++
					return
				}

				entry := ReadData(mergeData.fileReader)
				mergeData.SetRecord(entry)
				mergeData.Read++
			}, MaxSizes[i-first])
			mergaData.function(mergaData)
			//fmt.Println("Record:", mergaData.GetRecord().GetKey(), "Value:", mergaData.GetRecord().value, "Deleted:", mergaData.GetRecord().deleted, "Time:", mergaData.GetRecord().time)
			pq.Push(mergaData)
			//fmt.Println("Priority Queue Data", pq.data)
		}
		emptyString := ""
		iter := utils.NewIterator(pq, &emptyString, func(pq *PriorityQueue[MergeData[Entry]], s *string) (*Entry, bool) {
			if pq == nil || pq.Len() == 0 {
				return nil, false
			}
			top := pq.Peek()
			pq.Pop()
			//fmt.Println("Read value", top.Read)
			//fmt.Println("Top Record:", top.GetRecord().GetKey(), "Value:", top.GetRecord().value, "Deleted:", top.GetRecord().deleted, "Time:", top.GetRecord().time)
			for top.Read > top.maxSize {
				// Handling no more data anywhere
				if pq.Len() == 0 {
					return nil, false
				}
				top = pq.Peek()
				pq.Pop()
			}
			//fmt.Println("AFter for loop", top.Read)
			if top == nil {
				return nil, false
			}
			record := top.GetRecord()
			//fmt.Println("FIRST record past")
			if record == nil {
				return nil, false
			}
			//fmt.Println("Record before entering:", record.GetKey(), "Value:", record.value, "Deleted:", record.deleted, "Time:", record.time)
			for record.GetKey() == *s {
				top.function(top)
				//fmt.Println("Record after function call:", record.GetKey(), "Value:", record.value, "Deleted:", record.deleted, "Time:", record.time)
				//fmt.Println("Top Record:", top.GetRecord().GetKey(), "Value:", top.GetRecord().value, "Deleted:", top.GetRecord().deleted, "Time:", top.GetRecord().time)
				pq.Push(top)
				top = pq.Peek()
				pq.Pop()
				for top.Read > top.maxSize {
					//Handling no more data enywhere
					if pq.Len() == 0 {
						return nil, false
					}
					top = pq.Peek()
					pq.Pop()
				}
				record = top.GetRecord()
			}
			*s = record.GetKey()
			recordCopy := &Entry{
				key:     record.GetKey(),
				value:   record.value,
				deleted: record.deleted,
				time:    record.time,
			}
			// Create a copy of the record to return
			// This is to avoid returning a pointer to the original record
			//fmt.Println("Returning Record:", recordCopy.GetKey(), "Value:", recordCopy.value, "Deleted:", recordCopy.deleted, "Time:", recordCopy.time)
			top.function(top)
			pq.Push(top)
			return recordCopy, true
		})
		startpath := "./data/SSTabel" + fmt.Sprintf("%d", time.Now().UnixMicro())
		sum := 0
		for i := first; i < last; i++ {
			sum += MaxSizes[i-first]
		}
		Fileter := NewBloomFilterWithParams(sum, 0.01)
		tree := NewMerkleTree()
		path1 := WriteToFile(startpath, oneFile, lt.BlockMenager, sum, Fileter, summeryIndex, iter, firstKey, lastKey, tree)
		lt.AddtoNextLevel(level+1, path1)
		return path1
	}
}
func (lt *LSMTree) AddtoNextLevel(level int, path string) {
	if level >= lt.maxLevel {
		level = lt.maxLevel - 1 // Ensure we do not exceed the maximum level
		if len(lt.Levels) <= level {
			// If the Levels slice is empty, initialize it with a new slice.
			lt.Levels = append(lt.Levels, &[]*string{})
		}
		if lt.LevelsSize[level] == 0 {
			// If the size of the current level is zero, initialize it with a new slice.
			lt.Levels[level] = &[]*string{}
		}
		level1 := lt.Levels[level]
		appended := []*string{&path}
		*level1 = append(*level1, appended...)
		lt.LevelsSize[level]++
		return
	}
	if len(lt.Levels) <= level {
		// If the Levels slice is empty, initialize it with a new slice.
		lt.Levels = append(lt.Levels, &[]*string{})
	}
	if lt.LevelsSize[level] == 0 {
		// If the size of the current level is zero, initialize it with a new slice.
		lt.Levels[level] = &[]*string{}
	}
	level1 := lt.Levels[level]
	appended := []*string{&path}
	*level1 = append(appended, *level1...)
	lt.LevelsSize[level]++
}
func (lt *LSMTree) StartMerging(configMergeFunction int, dictenery *GlobalDictionary, oneFile bool, summeryIndex uint64, MergeSize int) {
	i := 0
	for i < len(lt.Levels) {
		if lt.LevelsSize[i] > 0 {
			if configMergeFunction == 0 {
				first, last, levle, ok := lt.GetSizeTieredMerge(i, MergeSize)
				if ok {
					lt.Merge(first, last, levle, dictenery, oneFile, summeryIndex)
					// After merging, we need to remove the merged files from the current level
					lt.LevelsSize[i] -= (last - first + 1)
					level := lt.Levels[i]
					j := 0
					levelForDelete := make([]*string, 0)
					newLevel := make([]*string, 0)
					for j < len(*level) {
						if j < first || j > last {
							newLevel = append(newLevel, (*level)[j])
						} else {
							levelForDelete = append(levelForDelete, (*level)[j])
						}
						j++
					}
					lt.Levels[i] = &newLevel
					// Remove the merged files from the current level
					lt.SerializeLSM(GlobalLSMTreeFile, lt.BlockMenager)
					j = 0
					for j := 0; j < len(levelForDelete); j++ {
						RemoveSStabel(*levelForDelete[j], lt.BlockMenager)
					}

					// Update the level with the new level without the merged files

					j = 0
					levelSerch := *lt.Levels[i]
					for j < len(levelSerch) {
					}
				}
			} else {
				// Implement other merge strategies here
				first, last, levle, ok := lt.GetLevelTieredMerge(i, MergeSize)
				if ok {
					lt.Merge(first, last, levle, dictenery, oneFile, summeryIndex)
					// After merging, we need to remove the merged files from the current level
					lt.LevelsSize[i] -= (last - first + 1)
					// Remove the merged files from the current level
					level := lt.Levels[i]
					j := 0
					levelForDelete := make([]*string, 0)
					newLevel := make([]*string, 0)
					for j < len(*level) {
						if j < first || j > last {
							newLevel = append(newLevel, (*level)[j])
						} else {
							levelForDelete = append(levelForDelete, (*level)[j])
						}
						j++
					}
					// Update the level with the new level without the merged files
					lt.Levels[i] = &newLevel
					j = 0
					lt.SerializeLSM(GlobalLSMTreeFile, lt.BlockMenager)
					for j := 0; j < len(levelForDelete); j++ {
						RemoveSStabel(*levelForDelete[j], lt.BlockMenager)
					}

					levelSerch := *lt.Levels[i]
					for j < len(levelSerch) {
					}
				}
			}

		}
		i++
	}
}
func (lt *LSMTree) GetSizeTieredMerge(levle int, mergeSize int) (int, int, int, bool) {
	if levle >= lt.maxLevel {
		return 0, 0, 0, false // Ensure we do not exceed the maximum level
	}
	if len(lt.Levels) <= levle || lt.LevelsSize[levle] == 0 {
		return 0, 0, 0, false // No data to merge
	}
	first := 0
	last := lt.LevelsSize[levle]
	if last < mergeSize {
		return first, last, levle, false // Not enough data to merge
	}
	if last >= mergeSize {
		last = mergeSize - 1 // Limit to merge size
	}
	return first, last, levle, true // Ready for merging
}
func (lt *LSMTree) GetLevelTieredMerge(level int, mergeSize int) (int, int, int, bool) {
	if level >= lt.maxLevel {
		return level, 0, 0, false // Ensure we do not exceed the maximum level
	}
	if len(lt.Levels) <= level || lt.LevelsSize[level] == 0 {
		return level, 0, 0, false // No data to merge
	} //Make merging easy ready first and last key
	if lt.LevelsSize[level] < mergeSize {
		return level, 0, 0, false // Not enough data to merge
	}
	last := lt.LevelsSize[level] - 1
	path := *(*lt.Levels[level])[last]
	FirstFKey, LastFKey := GetFirstAndLAstKeyCompresed(path, lt.BlockMenager)
	last -= 1
	for last >= 0 {
		path := *(*lt.Levels[level])[last]
		FirstSKey, LastSKey := GetFirstAndLAstKeyCompresed(path, lt.BlockMenager)
		if LastSKey >= FirstFKey || FirstSKey <= LastFKey {
			return last, mergeSize - 1, level, true // Overlap detected in either order, ready for merging
		}
		last--
	}
	// Rearenge lsm tree
	last = lt.LevelsSize[level] - 1
	pathToLast := *(*lt.Levels[level])[last]
	// If we reach here, it means there is no overlap, so we cant merge just send the last file to next level
	// Remove the last file from the current level
	levelData := lt.Levels[level]
	newLevelData := make([]*string, 0)
	for i := 0; i < len(*levelData); i++ {
		if i != last {
			newLevelData = append(newLevelData, (*levelData)[i])
		}
	}
	lt.Levels[level] = &newLevelData
	lt.LevelsSize[level]--
	lt.AddtoNextLevel(level+1, pathToLast)
	return level, 0, lt.LevelsSize[level] - 1, false // Ready for merging
}
func (lt *LSMTree) ReadSStabels(key string, bm *BlockManager, dictenery *GlobalDictionary) (*Entry, bool) {
	// Implement reading from SSTable here
	// This function should read the SSTable and return the ExportEntry for the given key
	i := 0
	for i < len(lt.Levels) {
		if lt.LevelsSize[i] > 0 {
			level := lt.Levels[i]
			for j := 0; j < lt.LevelsSize[i]; j++ {
				filePath := *(*level)[j]
				// Read the data from the file
				entry, found := ReadSStabel(filePath, key, bm, dictenery)
				if found {
					return entry, true
				}
			}
		}
		i++
	}
	return nil, false // Key not found in any SSTable
}
func (lt *LSMTree) SerializeLSM(path string, bm *BlockManager) {
	// Return the levels of the LSM tree
	path1 := path + "backup"
	file := NewBlockManagerWrap(path1, bm)
	file.WriteToFile(utils.Uint64ToBytes(uint64(lt.maxLevel)))
	file.WriteToFile(utils.Uint64ToBytes(uint64(len(lt.Levels))))
	for i := 0; i < len(lt.Levels); i++ {
		level := lt.Levels[i]
		file.WriteToFile(utils.Uint64ToBytes(uint64(lt.LevelsSize[i])))
		for j := 0; j < lt.LevelsSize[i]; j++ {
			path := *(*level)[j]
			file.WriteToFile(utils.Uint64ToBytes(uint64(len(path))))
			file.WriteToFile([]byte(path))
		}
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		// If the file does  exist, remove it
		err := os.Remove(path)
		if err != nil {
			fmt.Println("Error removing existing file:", err)
			return
		}
	}
	file.FlushBuffer(false)
	os.Rename(path1, path)
}
func DeserializeLSM(path string, bm *BlockManager) *LSMTree {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		path2 := path + "backup"

		if _, err := os.Stat(path2); os.IsNotExist(err) {
			return NewLSMTree(4, bm) //TODO make confing like that
		} else {
			os.Rename(path2, path)
		}
	}
	file := NewBlockManagerWrap(path, bm)
	_, maxLevelBytes := file.ReadFromCurrentPosition(8)
	maxLevel := utils.BytesToUint64(maxLevelBytes)
	_, levelsCountBytes := file.ReadFromCurrentPosition(8)
	levelsCount := utils.BytesToUint64(levelsCountBytes)
	tree := NewLSMTree(int(maxLevel), bm)
	for i := 0; i < int(levelsCount); i++ {
		_, sizeBytes := file.ReadFromCurrentPosition(8)
		size := utils.BytesToUint64(sizeBytes)
		levelPaths := make([]*string, 0)

		tree.LevelsSize[i] = int(size)
		for j := 0; j < int(size); j++ {
			_, pathLengthBytes := file.ReadFromCurrentPosition(8)
			pathLength := utils.BytesToUint64(pathLengthBytes)
			_, pathBytes := file.ReadFromCurrentPosition(int(pathLength))
			pathString := string(pathBytes)
			// Check if file exist as SStabel
			if CheckSStable(pathString, bm) {
				levelPaths = append(levelPaths, &pathString)
			} else {
				tree.LevelsSize[i]-- // If the file does not exist, we do not count it in the size of the level
			}
		}
		tree.Levels = append(tree.Levels, &levelPaths)
	}
	return tree
}

func (lt *LSMTree) Print() {
	if len(lt.Levels) == 0 {
		fmt.Println("Tree is empty.")
		return
	}

	for i := 0; i < len(lt.Levels); i++ {
		if lt.LevelsSize[i] > 0 {
			fmt.Println("Level ", i, " with size ", lt.LevelsSize[i])
			level := lt.Levels[i]
			for j := 0; j < lt.LevelsSize[i]; j++ {
				filePath := *(*level)[j]
				fmt.Println("Path: ", filePath, " with index ", j)
			}
		}
	}
}

func (lt *LSMTree) ValidateTable(level int, tableNumber int) (bool, int) {
	if level < 0 || level >= len(lt.Levels) {
		fmt.Println("Invalid level:", level)
		return false, -1
	}
	if tableNumber < 0 || tableNumber >= lt.LevelsSize[level] {
		fmt.Println("Invalid table number:", tableNumber)
		return false, -1
	}

	return ValidateSStabel(*(*lt.Levels[level])[tableNumber], lt.BlockMenager)
}
