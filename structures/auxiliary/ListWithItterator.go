package auxiliary

import (
	"KVengine/utils"
	"sort"
)

type ListWithIterator[T any] struct {
	entries []*T // Slice of ExportEntry pointers
}

func NewListWithIterator[T any]() *ListWithIterator[T] {
	return &ListWithIterator[T]{
		entries: make([]*T, 0),
	}

}
func (list *ListWithIterator[T]) Add(entry *T) {
	list.entries = append(list.entries, entry)
}

func (list *ListWithIterator[T]) Sort(compare func(a, b *T) bool) {
	sort.Slice(list.entries, func(i, j int) bool {
		return compare(list.entries[i], list.entries[j])
	})
}

func (list *ListWithIterator[T]) GetIter() *utils.Iterator[[]*T, int, *T] {
	zero := 0
	return utils.NewIterator(&list.entries, &zero, func(entries *[]*T, pos *int) (*T, bool) {
		if *pos >= len(*entries) {
			return nil, false
		}
		*pos = *pos + 1
		return (*entries)[*pos-1], true
	})
}
