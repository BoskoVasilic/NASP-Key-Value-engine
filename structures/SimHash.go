package structures

import (
	"crypto/md5"
	"fmt"
	"strings"
)

// Helper function: Generates a binary hash string from data.
func getHashAsString(data []byte) string {
	hash := md5.Sum(data)
	res := ""
	for _, b := range hash {
		res += fmt.Sprintf("%08b", b) // Ensure each byte is represented as 8 bits
	}
	return res
}

// TokenizeString prepares the input text for the SimHash function.
// It removes special characters, stop words, and counts word occurrences.
func tokenizeString(text string) map[string]int {
	splitWords := strings.FieldsFunc(text, func(r rune) bool {
		return r == ' ' || r == ',' || r == '.' || r == '\n' || r == ';'
	})

	stopWords := map[string]struct{}{
		"i": {}, "me": {}, "my": {}, "myself": {}, "we": {}, "our": {}, "ours": {}, "ourselves": {}, "you": {}, "your": {},
		"yours": {}, "yourself": {}, "yourselves": {}, "he": {}, "him": {}, "his": {}, "himself": {}, "she": {}, "her": {},
		"hers": {}, "herself": {}, "it": {}, "its": {}, "itself": {}, "they": {}, "them": {}, "their": {}, "theirs": {},
		"themselves": {}, "what": {}, "which": {}, "who": {}, "whom": {}, "this": {}, "that": {}, "these": {}, "those": {},
		"am": {}, "is": {}, "are": {}, "was": {}, "were": {}, "be": {}, "been": {}, "being": {}, "have": {}, "has": {},
		"had": {}, "having": {}, "do": {}, "does": {}, "did": {}, "doing": {}, "a": {}, "an": {}, "the": {}, "and": {},
		"but": {}, "if": {}, "or": {}, "because": {}, "as": {}, "until": {}, "while": {}, "of": {}, "at": {}, "by": {},
		"for": {}, "with": {}, "about": {}, "against": {}, "between": {}, "into": {}, "through": {}, "during": {},
		"before": {}, "after": {}, "above": {}, "below": {}, "to": {}, "from": {}, "up": {}, "down": {}, "in": {},
		"out": {}, "on": {}, "off": {}, "over": {}, "under": {}, "again": {}, "further": {}, "then": {}, "once": {},
		"here": {}, "there": {}, "when": {}, "where": {}, "why": {}, "how": {}, "all": {}, "any": {}, "both": {},
		"each": {}, "few": {}, "more": {}, "most": {}, "other": {}, "some": {}, "such": {}, "no": {}, "nor": {}, "not": {},
		"only": {}, "own": {}, "same": {}, "so": {}, "than": {}, "too": {}, "very": {}, "s": {}, "t": {}, "can": {},
		"will": {}, "just": {}, "don": {}, "should": {}, "now": {},
	}

	wordOccurrences := make(map[string]int)

	for _, word := range splitWords {
		word = strings.ToLower(word) // Normalize case
		if _, found := stopWords[word]; found {
			continue
		}
		wordOccurrences[word]++
	}

	return wordOccurrences
}

// SimHash generates a SimHash fingerprint for the given text.
func SimHash(text string) []byte {
	tokens := tokenizeString(text)
	hashVector := make([]int, 128) // 128-bit vector initialized to zero

	// Iterate through tokens and update the hash vector
	for word, count := range tokens {
		binaryHash := getHashAsString([]byte(word)) // Binary hash string for the word
		for i, bit := range binaryHash {
			if bit == '1' {
				hashVector[i] += count
			} else {
				hashVector[i] -= count
			}
		}
	}

	// Convert the hash vector into a final binary fingerprint
	fingerprint := make([]byte, 16) // 16 bytes for a 128-bit fingerprint
	for i := 0; i < 128; i++ {
		byteIndex := i / 8
		bitIndex := i % 8
		if hashVector[i] > 0 {
			fingerprint[byteIndex] |= (1 << (7 - bitIndex))
		}
	}

	return fingerprint
}

// CalculateHemmingwayDistance compares two fingerprints and calculates the Hamming distance.
func CalculateHemmingwayDistance(fingerprint1, fingerprint2 []byte) int {
	if len(fingerprint1) != len(fingerprint2) {
		panic("Inputs must have the same length")
	}

	distance := 0
	for i := 0; i < len(fingerprint1); i++ {
		xorResult := fingerprint1[i] ^ fingerprint2[i]
		distance += bitsSetCount(xorResult)
	}
	return distance
}

// Helper function to count the number of 1s in a byte
func bitsSetCount(b byte) int {
	count := 0
	for b > 0 {
		count += int(b & 1)
		b >>= 1
	}
	return count
}

// TestSimHashAndHemmingwayDistance tests SimHash and Hamming distance calculation
func TestSimHashAndHemmingwayDistance() {
	texts := []string{
		"Climate change is reshaping the Arctic regime. Fires burned 50 percent more acreage each year in the Arctic.",
		"Fires in the Arctic have evolved to burn infrequently, but climate change is increasing their frequency.",
		"The Arctic region is undergoing rapid transformation due to climate change and increasing fire events.",
		"This is an unrelated text with different content to test SimHash distance.",
		"Ancient civilizations built remarkable structures and achieved great technological advancements.",
		"SimHash is a technique used for detecting near-duplicate documents efficiently.",
	}

	fmt.Println("\nGenerated Fingerprints:")
	fingerprints := make([][]byte, len(texts))
	for i, text := range texts {
		fingerprints[i] = SimHash(text)
		fmt.Printf("Text %d: %08b\n", i+1, fingerprints[i])
	}

	fmt.Println("\nHamming Distance Matrix:")
	for i := 0; i < len(texts); i++ {
		for j := 0; j < len(texts); j++ {
			distance := CalculateHemmingwayDistance(fingerprints[i], fingerprints[j])
			fmt.Printf("%3d ", distance)
		}
		fmt.Println()
	}

	fmt.Println("\nAnalysis:")
	if CalculateHemmingwayDistance(fingerprints[0], fingerprints[1]) <
		CalculateHemmingwayDistance(fingerprints[0], fingerprints[3]) {
		fmt.Println("GOOD: Text 1 and Text 2 are more similar than Text 1 and Text 4, as expected.")
	} else {
		fmt.Println("BAD: Unexpected result: Text 1 and Text 4 appear more similar than Text 1 and Text 2.")
	}

	if CalculateHemmingwayDistance(fingerprints[0], fingerprints[2]) <
		CalculateHemmingwayDistance(fingerprints[0], fingerprints[4]) {
		fmt.Println("GOOD: Text 1 and Text 3 are more similar than Text 1 and Text 5, as expected.")
	} else {
		fmt.Println("BAD: Unexpected result: Text 1 and Text 5 appear more similar than Text 1 and Text 3.")
	}

	if CalculateHemmingwayDistance(fingerprints[1], fingerprints[2]) <
		CalculateHemmingwayDistance(fingerprints[1], fingerprints[5]) {
		fmt.Println("GOOD: Text 2 and Text 3 are more similar than Text 2 and Text 6, as expected.")
	} else {
		fmt.Println("BAD: Unexpected result: Text 2 and Text 6 appear more similar than Text 2 and Text 3.")
	}

}
