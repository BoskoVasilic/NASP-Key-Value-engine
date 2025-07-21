package structures

type Memtables struct {
	n            int
	maxEntry     int
	tables       []Storage
	activeIndex  int
	doneChannels []chan struct{}
}

func NewMemtables(n int, maxEntry int) *Memtables {
	return &Memtables{
		n,
		maxEntry,
		make([]Storage, n),
		0,
		make([]chan struct{}, n),
	}
}

func (mt *Memtables) getActive() Storage {
	return mt.tables[mt.activeIndex]
}

func (mt *Memtables) Put(key string, value []byte) {
	mt.getActive().Put(key, value)
	mt.checkFullness()
}

func (mt *Memtables) Get(key string) ([]byte, bool) {
	for i := mt.activeIndex; i >= 0; i-- {
		value, found := mt.getActive().Get(key)
		if found {
			return value, found
		}
	}

	return nil, false
}

func (mt *Memtables) Delete(key string) {
	mt.getActive().Delete(key)
	mt.checkFullness()
}

func (mt *Memtables) checkFullness() {
	if mt.getActive().GetSize() == mt.maxEntry {
		newIndex := (mt.activeIndex + 1) % mt.n
		<-mt.doneChannels[newIndex]
		mt.doneChannels[mt.activeIndex] = make(chan struct{})
		go mt.flush(mt.activeIndex)
		mt.activeIndex = newIndex
	}
}

func (mt *Memtables) flush(index int) {
	// ssTabel := NewSSTable(mt.tables[index].GetIter(), mt.maxEntry)
	// TODO call SSTabel
	close(mt.doneChannels[index])
}
