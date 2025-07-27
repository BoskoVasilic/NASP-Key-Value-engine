package structures

import (
	"fmt"
	"strings"
)

func RangeScan(first, last string, PageNumber, NumberOffEntytyinPage int, dictionary *GlobalDictionary, bm *BlockManager) {
	available := GlobalTokenBucket.CheckAvailability()
	if !available {
		fmt.Println("Access limit exceeded!")
		return
	}
	DicLock.RLock()
	itter := dictionary.SkipList.RangeScan(first, last)
	for i := 0; i < (PageNumber-1)*NumberOffEntytyinPage; i++ {
		Value, ok := itter.Next()
		if !ok {
			return
		}
		_, ok = GlobalMemtables.Get(Value.GetKey(), true)
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
		data, ok := GlobalMemtables.Get(key, true)
		if !ok {
			i--
		}
		if data == nil {
			continue
		}
		fmt.Println("Key:", key[1:], "Value:", string(data))
	}
	DicLock.RUnlock()
}
func PrefixScan(prefix string, PageNumber, NumberOffEntytyinPage int, dictionary *GlobalDictionary, bm *BlockManager) {
	first := prefix
	last := incrementLastRune(prefix)
	RangeScan(first, last, PageNumber, NumberOffEntytyinPage, dictionary, bm)
}
func RangeItterator(first, last string, dictionary *GlobalDictionary, bm *BlockManager) {
	itter := dictionary.SkipList.RangeScan(first, last)
	entry, ok := itter.Next()
	if !ok {
		fmt.Println("No entries")
		return
	}
	data, ok := GlobalMemtables.Get(entry.GetKey())
	var addeed string
	for !ok {
		entry, ok = itter.Next()
		if !ok {
			fmt.Println("No more entries")
			return
		}
		data, ok = GlobalMemtables.Get(entry.GetKey(), true)
	}
	for {
		fmt.Println("Key:", entry.GetKey()[1:], "Value:", string(data))
		entry, ok = itter.Next()
		if !ok {
			fmt.Println("No more entries")
			break
		}
		data, ok = GlobalMemtables.Get(entry.GetKey())
		for !ok {
			entry, ok = itter.Next()
			if !ok {
				fmt.Println("No more entries")
				return
			}
			data, ok = GlobalMemtables.Get(entry.GetKey(), true)
		}
		for {
			fmt.Print("[ITERATOR]> ")
			fmt.Scanln(&addeed)
			if strings.ToUpper(addeed) == "NEXT" {
				break
			} else if strings.ToUpper(addeed) == "STOP" {
				return
			} else {
				fmt.Println("Invalid command. Type 'NEXT' to continue or 'STOP' to exit.")
				continue
			}
		}
	}
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
