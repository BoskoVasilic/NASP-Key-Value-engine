package structures

import (
	"KVengine/utils"
)

type Memtables struct {
	n              int
	maxEntry       int
	tables         []Storage
	states         []WriteStatus
	activeIndex    int
	flushChannel   chan bool
	killSubroutine chan bool
	confirmKill    chan bool
}

func NewMemtables(n int, maxEntry int) *Memtables {
	if n < 2 {
		panic("N must be a natural number greater than 1")
	}

	tables := make([]Storage, n)

	for i := range n {
		tupe := GlobalConfig.Memtable.StructNum
		switch tupe {
		case 1:
			tables[i] = NewBTree(4)
		case 2:
			tables[i] = NewSkipList()
		case 3:
			tables[i] = NewHashMap(maxEntry + 31)
		}
	}

	mt := Memtables{
		n,
		maxEntry,
		tables,
		make([]WriteStatus, n),
		0,
		make(chan bool, n-1),
		make(chan bool),
		make(chan bool),
	}

	return &mt
}

func (mt *Memtables) Init() {
	go mt.flushThread()
}

func (mt *Memtables) flushThread() {
	i := 0
	for {
		select {
		case <-mt.killSubroutine:
			mt.confirmKill <- true
			for range mt.flushChannel {
				mt.flush(i)
				i = (i + 1) % mt.n
			}
			mt.confirmKill <- true
			return
		case <-mt.flushChannel:
			mt.flush(i)
			i = (i + 1) % mt.n
		}
	}

}

func (mt *Memtables) Destroy() {
	mt.killSubroutine <- true
	<-mt.confirmKill
	close(mt.flushChannel)
	<-mt.confirmKill
}

func (mt *Memtables) getActive() Storage {
	return mt.tables[mt.activeIndex]
}

func (mt *Memtables) Put(key string, value []byte, privilege ...bool) bool{
	if !(len(privilege) > 0 && privilege[0]) {
		available := GlobalTokenBucket.CheckAvailability()
		if !available {
			// fmt.Println("Access limit exceeded!")
			return false
		}
	}
	mt.getActive().Put(key, value)
	WalLock.Lock()
	GlobalWal.WriteToFile([]byte(key), value, false)
	WalLock.Unlock()
	AddItmsString(GlobalD, key)
	GlobalCache.Update(key, Entry{value: value, key: key})
	mt.shiftIfFull()	
	return true

}

func (mt *Memtables) PutWithoutWal(key string, value []byte) {
	mt.getActive().Put(key, value)
	AddItmsString(GlobalD, key)
	GlobalCache.Update(key, Entry{value: value, key: key})
	mt.shiftIfFull()
}

func (mt *Memtables) Get(key string, privilege ...bool) ([]byte, bool, bool) {
	if !(len(privilege) > 0 && privilege[0]) {
		available := GlobalTokenBucket.CheckAvailability()
		if !available {
			// fmt.Println("Access limit exceeded!")
			return nil, false, true
		}
	}
	for i := 0; i < mt.n; i++ {
		ind := (mt.activeIndex - i + mt.n) % mt.n
		value, found := mt.tables[ind].Get(key)
		if found {
			if !value.deleted {
				return value.value, true, false
			} else {
				return nil, false, false
			}
		}
	}
	value, found := GlobalCache.Get(key)
	if found {
		if !value.deleted {
			return value.value, true, false
		} else {
			return nil, false, false
		}
	}

	valPtr, found := GlobalLsm.ReadSStabels(key, GlobalBM, GlobalD)
	if found {
		GlobalCache.Put(key, *valPtr)
		if !valPtr.deleted {
			return valPtr.value, true, false
		} else {
			return nil, false, false
		}
	}

	return nil, false, false
}

func (mt *Memtables) Delete(key string, privilege ...bool) (bool){
	if !(len(privilege) > 0 && privilege[0]) {
		available := GlobalTokenBucket.CheckAvailability()
		if !available {
			// fmt.Println("Access limit exceeded!")
			return false
		}
	}
	mt.getActive().Delete(key)
	WalLock.Lock()
	GlobalWal.WriteToFile([]byte(key), make([]byte, 1), true)
	WalLock.Unlock()
	AddItmsString(GlobalD, key)
	GlobalCache.Delete(key)
	mt.shiftIfFull()
	return true
}

func (mt *Memtables) shiftIfFull() {
	if mt.getActive().GetSize() == mt.maxEntry {
		mt.states[mt.activeIndex] = GlobalWal.GetCurrentStatus()
		mt.activeIndex = (mt.activeIndex + 1) % mt.n
		mt.flushChannel <- true
		mt.tables[mt.activeIndex] = mt.getActive().GetNew()
	}
}

func (mt *Memtables) flush(index int) {
	typeInt := mt.tables[index].GetTypeInt()
	iterAny := mt.tables[index].GetIter()
	isCompressed := GlobalConfig.SSTable.CompressionEnabled
	isOneFile := !GlobalConfig.SSTable.SeparateFiles
	summary := GlobalConfig.SSTable.SummarySampling
	mergeAlg := GlobalConfig.LSMTree.AlgorithmType
	mergeSize := GlobalConfig.LSMTree.ParamNumber

	var path string
	if typeInt == 1 {
		iter := iterAny.(*utils.Iterator[BTree, btreeIterState, *Entry])
		SSTable := NewSSTable(iter, mt.tables[index].GetSize(), uint64(summary), isOneFile, GlobalBM, GlobalD, isCompressed)
		path = SSTable.GenerateSSTabel()
	} else if typeInt == 2 {
		iter := iterAny.(*utils.Iterator[Node, *Node, *Entry])
		SSTable := NewSSTable(iter, mt.tables[index].GetSize(), uint64(summary), isOneFile, GlobalBM, GlobalD, isCompressed)
		path = SSTable.GenerateSSTabel()
	} else if typeInt == 3 {
		iter := iterAny.(*utils.Iterator[[]*Entry, int, *Entry])
		SSTable := NewSSTable(iter, mt.tables[index].GetSize(), uint64(summary), isOneFile, GlobalBM, GlobalD, isCompressed)
		path = SSTable.GenerateSSTabel()
	} else {
		panic("Unknown type of Storage")
	}

	GlobalLsm.AddLevel(path)
	WalLock.Lock()
	GlobalWal.SetStatus(mt.states[index])
	WalLock.Unlock()
	GlobalLsm.StartMerging(mergeAlg, GlobalD, isOneFile, uint64(summary), mergeSize)
}
