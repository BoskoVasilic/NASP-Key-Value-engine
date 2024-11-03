package hash_with_seed

var functions = []func(string) int{
	hash1,
	hash2,
	hash3,
}

func HashString(s string) []int {
	results := []int{}

	for _, fn := range functions {
		result := fn(s)
		results = append(results, result)
	}

	return results
}

// TODO
// Dummy hash functions, to replace and add new below

func hash1(s string) int {
	return 1
}

func hash2(s string) int {
	return 2
}

func hash3(s string) int {
	return 3
}
