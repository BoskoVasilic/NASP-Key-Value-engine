package structures

import "fmt"

func RangeScan(first, last string, PageNumber, NumberOffEntytyinPage int, dictionary *GlobalDictionary, bm *BlockManager) {
	DicLock.RLock()
	itter := dictionary.SkipList.RangeScan(first, last)
	for i := 0; i < (PageNumber-1)*NumberOffEntytyinPage; i++ {
		Value, ok := itter.Next()
		if !ok {
			return
		}
		_ , ok = GlobalMemtables.Get(Value.GetKey())
		if !ok {
			i--
		}
	}
	for i := 0; i < NumberOffEntytyinPage; i++ {
		entry, ok := itter.Next()
		if !ok {
			return
		}
		key := entry.GetKey()
		data, ok := GlobalMemtables.Get(key)
		if !ok {
			i--
		}
		if data == nil {
			continue
		}
		fmt.Println("Key:", key, "Value:", data)
	}
	DicLock.RUnlock()
}
func PrefixScan(prefix string, PageNumber, NumberOffEntytyinPage int, dictionary *GlobalDictionary, bm *BlockManager) {
	first := prefix
	last := incrementLastRune(prefix)
	RangeScan(first, last, PageNumber, NumberOffEntytyinPage, dictionary, bm)
}
func RangeItterator(first, last string, dictionary *GlobalDictionary, bm *BlockManager) {
	DicLock.RLock()
	itter := dictionary.SkipList.RangeScan(first, last)
	entry, ok := itter.Next()
	fmt.Println("If you want new entry tipe Next if you are finished type Exit")
	if !ok {
		DicLock.RUnlock()
		return
	}
	data, ok := GlobalMemtables.Get(entry.GetKey())
	var addeed string
	for !ok {
		entry, ok = itter.Next()
		if !ok {
			DicLock.RUnlock()
			fmt.Println("No more entries")
			return
		}
		data, ok = GlobalMemtables.Get(entry.GetKey())
	}
	for {
		fmt.Println("Key:", entry.GetKey(), "Value:", data)
		entry, ok = itter.Next()
		if !ok {
			break
		}
		data, ok = GlobalMemtables.Get(entry.GetKey())
		for !ok {
			entry, ok = itter.Next()
			if !ok {
				DicLock.RUnlock()
				fmt.Println("No more entries")
				return
			}
			data, ok = GlobalMemtables.Get(entry.GetKey())
		}
		fmt.Scanln(&addeed)
		if addeed == "Next" {
			continue
		} else {
			break
		}
	}
	DicLock.RUnlock()
}
func PrefixItterator(prefix string, dictionary *GlobalDictionary, bm *BlockManager) {
	first := prefix
	last := incrementLastRune(prefix)
	RangeItterator(first, last, dictionary, bm)
}

func incrementLastRune(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	runes[len(runes)-1]++
	return string(runes)
}
