package structures

import (
	"KVengine/structures/auxiliary"
)

type MergeData[T any] struct {
	fileReader *BlockManagerWrap
	list       *auxiliary.ListWithIterator[T]
	function   func(*MergeData[T])
	record     *T // Record in both case have key value and Timestamp i just need to create it
	Read       int
	maxSize    int

	// Function to compare two elements of type T
}

func NewMergeData[T any](fileReader *BlockManagerWrap, list *auxiliary.ListWithIterator[T], function func(*MergeData[T]), maxSize int) *MergeData[T] {
	return &MergeData[T]{
		fileReader: fileReader,
		list:       list,
		function:   function,
		maxSize:    maxSize,
		record:     nil,
		Read:       0,
	}
}
func (mergeData *MergeData[T]) GetRecord() *T {
	if mergeData.record == nil {
		return nil
	}
	return mergeData.record
}

func (mergeData *MergeData[T]) SetRecord(record *T) {
	mergeData.record = record
}
