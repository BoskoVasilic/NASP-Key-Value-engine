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
	mt := Memtables{
		n,
		maxEntry,
		make([]Storage, n),
		make([]WriteStatus, n),
		0,
		make(chan bool, n-1),
		make(chan bool),
		make(chan bool),
	}
	go mt.flushThread()

	return &mt
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

func (mt *Memtables) Put(key string, value []byte) {
	mt.getActive().Put(key, value)
	WalLock.Lock()
	GlobalWal.WriteToFile([]byte(key), value, false)
	WalLock.Unlock()
	DicLock.Lock()
	AddItmsString(GlobalD, key)
	DicLock.Unlock()
	GlobalCache.Update(key, Entry{value: value, key: key})
	mt.shiftIfFull()
}

func (mt *Memtables) Get(key string) ([]byte, bool) {
	for i := 0; i < mt.n; i++ {
		ind := (mt.activeIndex - i + mt.n) % mt.n
		value, found := mt.tables[ind].Get(key)
		if found {
			if !value.deleted {
				return value.value, true
			} else {
				return nil, false
			}
		}
	}
	value, found := GlobalCache.Get(key)
	if found {
		if !value.deleted {
			return value.value, true
		} else {
			return nil, false
		}
	}

	valPtr, found := GlobalLsm.ReadSStabels(key, GlobalBM, GlobalD)
	if found {
		if !valPtr.deleted {
			return valPtr.value, true
		} else {
			return nil, false
		}
	}

	return nil, false
}

func (mt *Memtables) Delete(key string) {
	mt.getActive().Delete(key)
	GlobalWal.WriteToFile([]byte(key), make([]byte, 0), true)
	AddItmsString(GlobalD, key)
	GlobalCache.Delete(key)
	mt.shiftIfFull()
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
		iter := iterAny.(*utils.Iterator[*Node, **Node, *Entry])
		SSTable := NewSSTable(iter, mt.tables[index].GetSize(), uint64(summary), isOneFile, GlobalBM, GlobalD, isCompressed)
		path = SSTable.GenerateSSTabel()
	} else if typeInt == 3 {
		iter := iterAny.(*utils.Iterator[*[]*Entry, *int, *Entry])
		SSTable := NewSSTable(iter, mt.tables[index].GetSize(), uint64(summary), isOneFile, GlobalBM, GlobalD, isCompressed)
		path = SSTable.GenerateSSTabel()
	} else {
		panic("Unknown type of Storage")
	}

	lsm := NewLSMTree(mt.maxEntry, GlobalBM)
	lsm.AddLevel(path)
	WalLock.Lock()
	GlobalWal.SetStatus(mt.states[index])
	WalLock.Unlock()
	lsm.StartMerging(mergeAlg, GlobalD, isOneFile, uint64(summary), mergeSize)
}
