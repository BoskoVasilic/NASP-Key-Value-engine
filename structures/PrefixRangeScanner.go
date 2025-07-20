package structures

import "fmt"

// Need to be tested
func RangeScan(first, last string, PageNumber, NumberOffEntytyinPage int, dictionary *GlobalDictionary, LSM *LSMTree, bm *BlockManager) {
	itter := dictionary.SkipList.RangeScan(first, last)
	for i := 0; i < (PageNumber-1)*NumberOffEntytyinPage; i++ {
		_, ok := itter.Next()
		if !ok {
			return
		}
	}
	for i := 0; i < NumberOffEntytyinPage; i++ {
		entry, ok := itter.Next()
		if !ok {
			return
		}
		key := entry.GetKey()
		data, ok := LSM.ReadSStabels(key, bm, dictionary)
		if !ok {
			i--
		}
		if data == nil {
			continue
		}
		fmt.Println("Key:", key, "Value:", data.GetValue(), "Deleted:", data.IsDeleted())
	}

}
func PrefixScan(prefix string, PageNumber, NumberOffEntytyinPage int, dictionary *GlobalDictionary, LSM *LSMTree, bm *BlockManager) {
	first := prefix
	last := incrementLastRune(prefix)
	RangeScan(first, last, PageNumber, NumberOffEntytyinPage, dictionary, LSM, bm)
}
func RangeItterator(first, last string, dictionary *GlobalDictionary, LSM *LSMTree, bm *BlockManager) {
	itter := dictionary.SkipList.RangeScan(first, last)
	entry, ok := itter.Next()
	fmt.Println("If you want new entry tipe Next if you are finished type Exit")
	if !ok {
		return
	}
	data, ok := LSM.ReadSStabels(entry.GetKey(), bm, dictionary)
	var addeed string
	for 1 == 1 {
		fmt.Println("Key:", entry.GetKey(), "Value:", data.GetValue(), "Deleted:", data.IsDeleted())
		entry, ok = itter.Next()
		if !ok {
			break
		}
		data, ok = LSM.ReadSStabels(entry.GetKey(), bm, dictionary)

		fmt.Scanln(&addeed)
		if addeed == "Next" {
			continue
		} else {
			break
		}
	}

}
func PrefixRangeScan(prefix string, PageNumber, NumberOffEntytyinPage int, dictionary *GlobalDictionary, LSM *LSMTree, bm *BlockManager) {
	first := prefix
	last := incrementLastRune(prefix)
	RangeScan(first, last, PageNumber, NumberOffEntytyinPage, dictionary, LSM, bm)
}

func incrementLastRune(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	runes[len(runes)-1]++
	return string(runes)
}
