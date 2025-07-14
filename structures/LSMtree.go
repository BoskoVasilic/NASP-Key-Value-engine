package structures

import (
	"KVengine/utils"
	"fmt"
)

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
	*level1 = append(*level1, &path)
	lt.LevelsSize[0]++
}
func (lt *LSMTree) Merge(first int, last int, level int, dictenery *GlobalDictionary) {
	DataCompresedLists := make([]bool, 0)
	MaxSizes := make([]int, 0)
	OneFileLise := make([]bool, 0)
	levleSData := *(lt.Levels[level])
	ForceCompresion := false
	for i := first; i < last; i++ {
		file := NewBlockManagerWrap(*levleSData[i], lt.BlockMenager)
		_, IsCompresed := file.ReadFromCurrentPosition(1)
		_, OneFile := file.ReadFromCurrentPosition(1)
		fmt.Println("IsCompresed:", IsCompresed[0])
		_, data := file.ReadFromCurrentPosition(8)
		maxSizeC := utils.BytesToUint64(data)
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
				return a.GetRecord().time < b.GetRecord().time
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
				fmt.Println("_________________________________________________________________________________")
				j := 0
				for j < MaxSizes[i-first] {
					fmt.Println("List Export Entries:", list_ExportEntries.GetEntrie(j).GetKey(), "Value:", list_ExportEntries.GetEntrie(j).value, "Deleted:", list_ExportEntries.GetEntrie(j).tombstone, "Time:", list_ExportEntries.GetEntrie(j).time)
					j++
				}
				fmt.Println("List Export Entries:", list_ExportEntries)
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
		record, ok := iter.Next()
		for ok {
			fmt.Println("Record:", record.GetKey(), "Value:", record.value, "Deleted:", record.tombstone, "Time:", record.time)
			record, ok = iter.Next()
		}
	} else {
		pq := NewPriorityQueue(func(a, b *MergeData[Entry]) bool {
			if a.GetRecord().GetKey() == b.GetRecord().GetKey() {
				return a.GetRecord().time < b.GetRecord().time
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
		record, ok := iter.Next()
		for ok {
			fmt.Println("Record:", record.GetKey(), "Value:", record.value, "Deleted:", record.deleted, "Time:", record.time)
			record, ok = iter.Next()
		}
	}
}
