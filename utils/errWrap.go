package utils

import "fmt"

// ErrWrap function wraps the error handling of the function passed
// It accepts the function and args to pass to it
// It returns the result of the function if no error occurred
func ErrWrap(fn func(args ...interface{}) (interface{}, error), args ...interface{}) interface{} {
	result, err := fn(args...)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	return result
}
