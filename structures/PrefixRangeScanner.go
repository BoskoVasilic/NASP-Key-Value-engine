package structures

import (
	"fmt"
	"strings"
)

func RangeScan(first, last string, PageNumber, NumberOffEntytyinPage int, dictionary *GlobalDictionary, bm *BlockManager) {
	/*
		Rnge scan ide kroz dictionarijum i ispisuje sve kljuceve i
		vrednosti koji su u opsegu od first do last.
		Kada dobijem niz kljuceva poziva se funkcija Get iz GlobalMemtables
		koja vraca vrednost za dati kljuc.
	*/

	available := GlobalTokenBucket.CheckAvailability()
	if !available {
		fmt.Println("Access limit exceeded!")
		return
	}

	itter := dictionary.SkipList.RangeScan(first, last)
	for i := 0; i < (PageNumber-1)*NumberOffEntytyinPage; i++ {
		Value, ok := itter.Next()
		if !ok {
			return
		}
		_, ok, _ = GlobalMemtables.Get(Value.GetKey(), true)

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
		data, ok, _ := GlobalMemtables.Get(key, true)

		if !ok {
			i--
		}
		if data == nil {
			continue
		}
		fmt.Println("Key:", key[1:], "Value:", string(data))
	}

}
func PrefixScan(prefix string, PageNumber, NumberOffEntytyinPage int, dictionary *GlobalDictionary, bm *BlockManager) {
	/*
		Svodi se na range scan samo sto koristi prefix koji je first element a
		last element se dobija tako sto se doda 1 na zadnji deo u stringu.
	*/
	first := prefix
	last := incrementLastRune(prefix)
	RangeScan(first, last, PageNumber, NumberOffEntytyinPage, dictionary, bm)

}
func RangeItterator(first, last string, dictionary *GlobalDictionary, bm *BlockManager) {
	/*
		Range iterator ide kroz dictionarijum i dobavja kljuc po kljuc
		U trenutku kada dobavljanje novog kljuca nije moguce zavrsa ve se iteracija
		Vrednosti se dobavljaju iz GlobalMemtables i ispisuju se na konzoli.
	*/

	itter := dictionary.SkipList.RangeScan(first, last)
	entry, ok := itter.Next()
	if !ok {
		fmt.Println("No entries")
		return
	}
	data, ok, _ := GlobalMemtables.Get(entry.GetKey())

	var addeed string
	for !ok {
		entry, ok = itter.Next()
		if !ok {
			fmt.Println("No more entries")
			return
		}
		data, ok, _ = GlobalMemtables.Get(entry.GetKey(), true)

	}
	for {
		fmt.Println("Key:", entry.GetKey()[1:], "Value:", string(data))
		entry, ok = itter.Next()
		if !ok {
			fmt.Println("No more entries")
			break
		}
		data, ok, _ = GlobalMemtables.Get(entry.GetKey())

		for !ok {
			entry, ok = itter.Next()
			if !ok {
				fmt.Println("No more entries")
				return
			}
			data, ok, _ = GlobalMemtables.Get(entry.GetKey(), true)
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
	/*
		Svodi se na range itterator samo sto koristi prefix koji je first element a
		last element se dobija tako sto se doda 1 na zadnji deo u stringu.
	*/

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
