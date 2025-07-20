package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// This is just a placeholder interface for the integration with the actual engine.
// Hopefully these are all the function CLI should call.
type Engine interface {
	// Basic operations
	Put(key string, value []byte) error
	Get(key string) ([]byte, error)
	Delete(key string) error
	PrefixScan(prefix string, pageNumber, pageSize int) ([][2]string, error)
	RangeScan(minKey, maxKey string, pageNumber, pageSize int) ([][2]string, error)
	PrefixIterate(prefix string) (Iterator, error)
	RangeIterate(minKey, maxKey string) (Iterator, error)
	// Probability structures
	BloomCreate(name string) error
	BloomDelete(name string) error
	BloomAdd(name, elem string) error
	BloomCheck(name, elem string) (bool, error)
	CountMinCreate(name string) error
	CountMinDelete(name string) error
	CountMinAdd(name, elem string) error
	CountMinCheck(name, elem string) (int, error)
	HLLCreate(name string) error
	HLLDelete(name string) error
	HLLAdd(name, elem string) error
	HLLCount(name string) (int, error)
	SimHashSave(name, text string) error
	SimHashDistance(name1, name2 string) (int, error)
}

// Iterator for PREFIX_ITERATE and RANGE_ITERATE
type Iterator interface {
	Next() (string, []byte, bool) // key, value, ok
	Stop() error
}

// Ovaj deo ti kasnije zamenjuješ stvarnim engine-om
var engine Engine = nil

func Run() {
	fmt.Println("Welcome to our Key-Value Engine CLI!")
	fmt.Println("Type HELP for a list of commands.")
	reader := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !reader.Scan() {
			break
		}
		line := strings.TrimSpace(reader.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := strings.ToUpper(parts[0])

		switch cmd {
		case "HELP":
			printCommands()

		// Basic operations: PUT, GET, DELETE
		// PUT command expects at least 2 parts: key and value and adds them to the engine
		case "PUT":
			if len(parts) < 3 {
				fmt.Println("Usage: PUT key value")
				continue
			}
			key, value := parts[1], strings.Join(parts[2:], " ")
			fmt.Println("[PUT]", key, value)
			// engine.Put(key, []byte(value))

		// GET command expects at least 1 part: key and then retrieves the value from the engine
		case "GET":
			if len(parts) < 2 {
				fmt.Println("Usage: GET key")
				continue
			}
			key := parts[1]
			fmt.Println("[GET]", key)
			// val, _ := eng.Get(key); fmt.Println(string(val))

		// DELETE command expects at least 1 part: key and deletes it from the engine
		case "DELETE":
			if len(parts) < 2 {
				fmt.Println("Usage: DELETE key")
				continue
			}
			key := parts[1]
			fmt.Println("[DELETE]", key)
			// engine.Delete(key)

		// Scanners from 2.4
		case "PREFIX_SCAN":
			if len(parts) < 4 {
				fmt.Println("Usage: PREFIX_SCAN prefix pageNumber pageSize")
				continue
			}
			fmt.Println("[PREFIX_SCAN]", parts[1], parts[2], parts[3])
			// engine.PrefixScan(parts[1], parts[2], parts[3])

		case "RANGE_SCAN":
			if len(parts) < 5 {
				fmt.Println("Usage: RANGE_SCAN minKey maxKey pageNumber pageSize")
				continue
			}
			fmt.Println("[RANGE_SCAN]", parts[1], parts[2], parts[3], parts[4])
			// engine.RangeScan(parts[1], parts[2], parts[3], parts[4])

		// Iterators from 2.5
		case "PREFIX_ITERATE":
			if len(parts) < 2 {
				fmt.Println("Usage: PREFIX_ITERATE prefix")
				continue
			}
			fmt.Println("[PREFIX_ITERATE]", parts[1], " -> use NEXT and STOP")

		case "RANGE_ITERATE":
			if len(parts) < 3 {
				fmt.Println("Usage: RANGE_ITERATE minKey maxKey")
				continue
			}
			fmt.Println("[RANGE_ITERATE]", parts[1], parts[2], " -> use NEXT and STOP")
			// iterator := engine.RangeIterate(parts[1], parts[2])
			// fmt.Println("Iterator created. Use NEXT and STOP commands.")

		case "NEXT":
			fmt.Println("[NEXT] fetch next from iterator")
			// if iterator == nil {
			// 	fmt.Println("No active iterator. Use PREFIX_ITERATE or RANGE_ITERATE first.")
			// 	continue	
			// }
			// key, value, ok := iterator.Next()
		case "STOP":
			fmt.Println("[STOP] stop current iterator")
			// if iterator == nil {
			// 	fmt.Println("No active iterator to stop.")
			// 	continue
			// }
			// err := iterator.Stop()

			// if err != nil {
			// 	fmt.Println("Error stopping iterator:", err)
			// }

		// Probabilistic structures

		// possibly change the command names to be more consistent, change ADD to SAVE or INSERT

		// BloomFilter
		case "BF_CREATE":
			if len(parts) < 2 {
				fmt.Println("Usage: BF_CREATE name")
				continue
			}
			fmt.Println("[BloomFilter Create]", parts[1])
			// engine.NewBloomFilter(parts[1])

		case "BF_DELETE":
			if len(parts) < 2 {
				fmt.Println("Usage: BF_DELETE name")
				continue
			}
			fmt.Println("[BloomFilter Delete]", parts[1])
			// engine.DeleteBloomFilter(parts[1])

		case "BF_PUT":
			if len(parts) < 3 {
				fmt.Println("Usage: BF_PUT name element")
				continue
			}
			fmt.Println("[BloomFilter Put]", parts[1], parts[2])
			// engine.BloomPUT(parts[1], parts[2])

		case "BF_CHECK":
			if len(parts) < 3 {
				fmt.Println("Usage: BF_CHECK name element")
				continue
			}
			fmt.Println("[BloomFilter Check]", parts[1], parts[2])

		// CountMinSketch
		case "CMS_CREATE":
			if len(parts) < 2 {
				fmt.Println("Usage: CMS_CREATE name")
				continue
			}
			fmt.Println("[CountMinSketch Create]", parts[1])
			// engine.NewCountMinSketch(parts[1])

		case "CMS_DELETE":
			if len(parts) < 2 {
				fmt.Println("Usage: CMS_DELETE name")
				continue
			}
			fmt.Println("[CountMinSketch Delete]", parts[1])
			// engine.DeleteCountMinSketch(parts[1])

		case "CMS_ADD":
			if len(parts) < 3 {
				fmt.Println("Usage: CMS_ADD name element")
				continue
			}
			fmt.Println("[CountMinSketch Add]", parts[1], parts[2])
			// engine.CountMinAdd(parts[1], parts[2])

		case "CMS_CHECK":
			if len(parts) < 3 {
				fmt.Println("Usage: CMS_CHECK name element")
				continue
			}
			fmt.Println("[CountMinSketch Check]", parts[1], parts[2])
			// count, _ := engine.CountMinCheck(parts[1], parts[2])

		// HyperLogLog
		case "HLL_CREATE":
			if len(parts) < 2 {
				fmt.Println("Usage: HLL_CREATE name")
				continue
			}
			fmt.Println("[HyperLogLog Create]", parts[1])
			// engine.NewHyperLogLog(parts[1])

		case "HLL_DELETE":
			if len(parts) < 2 {
				fmt.Println("Usage: HLL_DELETE name")
				continue
			}
			fmt.Println("[HyperLogLog Delete]", parts[1])
			// engine.DeleteHyperLogLog(parts[1])

		case "HLL_ADD":
			if len(parts) < 3 {
				fmt.Println("Usage: HLL_ADD name element")
				continue
			}
			fmt.Println("[HyperLogLog Add]", parts[1], parts[2])
			// engine.HyperLogLogAdd(parts[1], parts[2])

		case "HLL_COUNT":
			if len(parts) < 2 {
				fmt.Println("Usage: HLL_COUNT name")
				continue
			}
			fmt.Println("[HyperLogLog Count]", parts[1])
			// count, _ := engine.HyperLogLogCount(parts[1])

		// SimHash
		case "SIMHASH_SAVE":
			if len(parts) < 3 {
				fmt.Println("Usage: SIMHASH_SAVE name text")
				continue
			}
			name := parts[1]
			text := strings.Join(parts[2:], " ")
			fmt.Println("[SimHash Save]", name, text)
			// engine.SimHashSave(name, text)

		case "SIMHASH_DIST":
			if len(parts) < 3 {
				fmt.Println("Usage: SIMHASH_DIST name1 name2")
				continue
			}
			fmt.Println("[SimHash Distance]", parts[1], parts[2])
			// Hamming distance between two SimHashes
			// dist, _ := engine.SimHashDistance(parts[1], parts[2])

		case "EXIT", "QUIT":
			fmt.Println("Exiting CLI. Goodbye!")
			return

		default:
			fmt.Println("Unknown command. Type HELP for a list.")
		}
	}
}

func printCommands() {
	fmt.Println("Available commands:")
	fmt.Println("PUT key value")
	fmt.Println("GET key")
	fmt.Println("DELETE key")
	fmt.Println("------------------------------")
	fmt.Println("PREFIX_SCAN prefix pageNumber pageSize")
	fmt.Println("RANGE_SCAN minKey maxKey pageNumber pageSize")
	fmt.Println("------------------------------")
	fmt.Println("PREFIX_ITERATE prefix  	(then use NEXT/STOP)")
	fmt.Println("RANGE_ITERATE minKey maxKey  	(then use NEXT/STOP)")
	fmt.Println("NEXT / STOP")
	fmt.Println("------------------------------")
	fmt.Println("Work with Probabilistic Structures:")
	fmt.Println("BF=BloomFilter, CMS=CountMinSketch, HLL=HyperLogLog, SIMHASH=SimHash")
	fmt.Println("BF_CREATE name / BF_DELETE name / BF_PUT name elem / BF_CHECK name elem")
	fmt.Println("CMS_CREATE name / CMS_DELETE name / CMS_ADD name elem / CMS_CHECK name elem")
	fmt.Println("HLL_CREATE name / HLL_DELETE name / HLL_ADD name elem / HLL_COUNT name")
	fmt.Println("SIMHASH_SAVE name text / SIMHASH_DIST name1 name2")
	fmt.Println("------------------------------")
	fmt.Println("HELP")
	fmt.Println("-------------------------------")
	fmt.Println("EXIT or QUIT")
}