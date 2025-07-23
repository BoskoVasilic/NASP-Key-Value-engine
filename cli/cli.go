package cli

import (
	"KVengine/structures"
	"bufio"
	"fmt"
	"os"
	"strings"
	"strconv"
)

// Iterator for PREFIX_ITERATE and RANGE_ITERATE
type Iterator interface {
	Next() (string, []byte, bool) // key, value, ok
	Stop() error
}

func Run() {

	// Initialize the global engine
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

			structures.GlobalMemtables.Put("d"+key, []byte(value))

		// GET command expects at least 1 part: key and then retrieves the value from the engine
		case "GET":
			if len(parts) < 2 {
				fmt.Println("Usage: GET key")
				continue
			}
			key := parts[1]
			fmt.Println("[GET]", "d"+key)
			val, _ := structures.GlobalMemtables.Get("d"+key)
			fmt.Println(string(val))

		// DELETE command expects at least 1 part: key and deletes it from the engine
		case "DELETE":
			if len(parts) < 2 {
				fmt.Println("Usage: DELETE key")
				continue
			}
			key := parts[1]
			fmt.Println("[DELETE]", "d"+key)
			structures.GlobalMemtables.Delete("d" + key)

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

		// BloomFilter
		case "BF_CREATE":
			if len(parts) < 4 {
				fmt.Println("Usage: BF_CREATE key expectedInsertions falsePositiveRate")
				continue
			}
			fmt.Println("[BloomFilter Create]", parts[1])

			if len(parts) < 4 {
				fmt.Println("Usage: BF_CREATE key expectedInsertions falsePositiveRate")
				continue
			}
			expectedInsertions, err1 := strconv.Atoi(parts[2])
			falsePositiveRate, err2 := strconv.ParseFloat(parts[3], 64)
			if err1 != nil || err2 != nil {
				fmt.Println("Invalid parameters for BF_CREATE. expectedInsertions should be int, falsePositiveRate should be float.")
				continue
			}
			bloomFilter := structures.NewBloomFilterWithParams(expectedInsertions, falsePositiveRate)
			// memtables
			structures.GlobalMemtables.Put("b"+parts[1], bloomFilter.SerializeBloomF())

		case "BF_DELETE":
			if len(parts) < 2 {
				fmt.Println("Usage: BF_DELETE key")
				continue
			}
			fmt.Println("[BloomFilter Delete]", parts[1])

			structures.GlobalMemtables.Delete("b" + parts[1])


		case "BF_PUT":
			if len(parts) < 3 {
				fmt.Println("Usage: BF_PUT key element")
				continue
			}
			fmt.Println("[BloomFilter Put]", parts[1], parts[2])

			bloomFilterData, found := structures.GlobalMemtables.Get("b" + parts[1])
			if !found {
				fmt.Println("BloomFilter not found")
				continue
			}
			bloomFilter := structures.DeserializeBloomF(bloomFilterData)
			bloomFilter.Put([]byte(parts[2]))
			structures.GlobalMemtables.Put("b"+parts[1], bloomFilter.SerializeBloomF())

		case "BF_CHECK":
			if len(parts) < 3 {
				fmt.Println("Usage: BF_CHECK key element")
				continue
			}
			fmt.Println("[BloomFilter Check]", parts[1], parts[2])

			bloomFilterData, found := structures.GlobalMemtables.Get("b" + parts[1])
			if !found {
				fmt.Println("BloomFilter not found")
				continue
			}
			bloomFilter := structures.DeserializeBloomF(bloomFilterData)
			isPresent := bloomFilter.Check([]byte(parts[2]))
			if isPresent {
				fmt.Println("Element is probably present in the BloomFilter.")
			} else {
				fmt.Println("Element is definitely not present in the BloomFilter.")
			}

		// CountMinSketch
		case "CMS_CREATE":
			if len(parts) < 4 {
				fmt.Println("Usage: CMS_CREATE key depth width")
				continue
			}
			fmt.Println("[CountMinSketch Create]", parts[1])

			depth, err1 := strconv.ParseUint(parts[2], 10, 32)
			width, err2 := strconv.ParseUint(parts[3], 10, 32)
			if err1 != nil || err2 != nil {
				fmt.Println("Invalid parameters for CMS_CREATE. depth and width should be integers.")
				continue
			}
			cms := structures.NewCountMinSketch(uint32(depth), uint32(width))
			structures.GlobalMemtables.Put("c"+parts[1], cms.Serialize())
			

		case "CMS_DELETE":
			if len(parts) < 2 {
				fmt.Println("Usage: CMS_DELETE key")
				continue
			}
			fmt.Println("[CountMinSketch Delete]", parts[1])
			
			structures.GlobalMemtables.Delete("c" + parts[1])

		case "CMS_ADD":
			if len(parts) < 3 {
				fmt.Println("Usage: CMS_ADD key element")
				continue
			}
			fmt.Println("[CountMinSketch Add]", parts[1], parts[2])
			
			cmsData, found := structures.GlobalMemtables.Get("c" + parts[1])
			if !found {
				fmt.Println("CountMinSketch not found")
				continue
			}
			cms := structures.DeserializeCMS(cmsData)
			cms.Add(parts[2])
			structures.GlobalMemtables.Put("c"+parts[1], cms.Serialize())


		case "CMS_CHECK":
			if len(parts) < 3 {
				fmt.Println("Usage: CMS_CHECK key element")
				continue
			}
			fmt.Println("[CountMinSketch Check]", parts[1], parts[2])
			
			cmsData, found := structures.GlobalMemtables.Get("c" + parts[1])
			if !found {
				fmt.Println("CountMinSketch not found")
				continue
			}
			cms := structures.DeserializeCMS(cmsData)
			count := cms.Estimate(parts[2])
			fmt.Printf("Estimated count for '%s': %d\n", parts[2], count)

		// HyperLogLog
		case "HLL_CREATE":
			if len(parts) < 2 {
				fmt.Println("Usage: HLL_CREATE key p")
				continue
			}
			fmt.Println("[HyperLogLog Create]", parts[1])

			p, err := strconv.ParseUint(parts[2], 10, 8)
			if err != nil {	
				fmt.Println("Invalid parameter for HLL_CREATE. p should be an integer.")
				continue	
			}
			p8 := uint8(p)
			hll := structures.NewHyperLogLog(p8) // Default p value
			structures.GlobalMemtables.Put("h"+parts[1], hll.SerializeHLL())


		case "HLL_DELETE":
			if len(parts) < 2 {
				fmt.Println("Usage: HLL_DELETE key")
				continue
			}
			fmt.Println("[HyperLogLog Delete]", parts[1])
			structures.GlobalMemtables.Delete("h" + parts[1])

		case "HLL_ADD":
			if len(parts) < 3 {
				fmt.Println("Usage: HLL_ADD key element")
				continue
			}
			fmt.Println("[HyperLogLog Add]", parts[1], parts[2])
			hllData, found := structures.GlobalMemtables.Get("h" + parts[1])
			if !found {
				fmt.Println("HyperLogLog not found")
				continue
			}
			hll := structures.DeserializeHLL(hllData)
			hll.Add(parts[2])
			structures.GlobalMemtables.Put("h"+parts[1], hll.SerializeHLL())

		case "HLL_COUNT":
			if len(parts) < 2 {
				fmt.Println("Usage: HLL_COUNT key")
				continue
			}
			fmt.Println("[HyperLogLog Count]", parts[1])

			hllData, found := structures.GlobalMemtables.Get("h" + parts[1])
			if !found {
				fmt.Println("HyperLogLog not found")
				continue
			}
			hll := structures.DeserializeHLL(hllData)
			count := hll.Estimate()
			fmt.Printf("Estimated count for '%s': %.0f\n", parts[1], count)


		// SimHash
		case "SIMHASH_SAVE":
			if len(parts) < 3 {
				fmt.Println("Usage: SIMHASH_SAVE key text")
				continue
			}
			name := parts[1]
			text := strings.Join(parts[2:], " ")
			fmt.Println("[SimHash Save]", name, text)

			structures.GlobalMemtables.Put("s"+name, []byte(text))	

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
	fmt.Println("BF_CREATE key expectedInsertions falsePositiveRate / BF_DELETE key / BF_PUT key elem / BF_CHECK key elem ")
	fmt.Println("CMS_CREATE key depth width / CMS_DELETE key / CMS_ADD key element / CMS_CHECK key element")
	fmt.Println("HLL_CREATE key p / HLL_DELETE key / HLL_ADD key element / HLL_COUNT key")
	fmt.Println("SIMHASH_SAVE key text / SIMHASH_DIST key1 key2")
	fmt.Println("------------------------------")
	fmt.Println("HELP")
	fmt.Println("-------------------------------")
	fmt.Println("EXIT or QUIT")
}
