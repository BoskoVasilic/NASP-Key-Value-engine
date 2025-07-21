package structures

type Memtables struct {
	n           int
	maxEntry    int
	tables      []Storage
	activeIndex int
}

func NewMemtables(n int, maxEntry int) *Memtables {
	return &Memtables{
		n,
		maxEntry,
		make([]Storage, n),
		0,
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
		mt.activeIndex++
		if mt.activeIndex == mt.n {
			// TODO flush
		}
	}
}
