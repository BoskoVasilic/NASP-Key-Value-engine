package structures

import (
	"KVengine/utils"
)

type Memtables struct {
	n              int
	maxEntry       int
	tables         []Storage
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
	mt.shiftIfFull()
}

func (mt *Memtables) Get(key string) ([]byte, bool) {
	for i := 0; i < mt.n; i++ {
		ind := (mt.activeIndex - i + mt.n) % mt.n
		value, found := mt.tables[ind].Get(key)
		if found {
			return value.value, found
		}
	}

	return nil, false
}

func (mt *Memtables) Delete(key string) {
	mt.getActive().Delete(key)
	mt.shiftIfFull()
}

func (mt *Memtables) shiftIfFull() {
	if mt.getActive().GetSize() == mt.maxEntry {
		mt.activeIndex = (mt.activeIndex + 1) % mt.n
		mt.flushChannel <- true
		// clear Storage at activeIndex


	}
}

func (mt *Memtables) flush(index int) {
	typeInt := mt.tables[index].GetTypeInt()
	iterAny := mt.tables[index].GetIter()
	isCompressed := globalConfig.SSTable.CompressionEnabled
	isOneFile := !globalConfig.SSTable.SeparateFiles
	summary := globalConfig.SSTable.SummarySampling
	mergeAlg := globalConfig.LSMTree.AlgorithmType
	mergeSize := globalConfig.LSMTree.ParamNumber

	var path string
	if typeInt == 1 {
		iter := iterAny.(*utils.Iterator[BTree, btreeIterState, *Entry])
		SSTable := NewSSTable(iter, mt.tables[index].GetSize(), uint64(summary), isOneFile, &globalBM, &globalD, isCompressed)
		path = SSTable.GenerateSSTabel()
	} else if typeInt == 2 {
		iter := iterAny.(*utils.Iterator[*Node, **Node, *Entry])
		SSTable := NewSSTable(iter, mt.tables[index].GetSize(), uint64(summary), isOneFile, &globalBM, &globalD, isCompressed)
		path = SSTable.GenerateSSTabel()
	} else if typeInt == 3 {
		iter := iterAny.(*utils.Iterator[*[]*Entry, *int, *Entry])
		SSTable := NewSSTable(iter, mt.tables[index].GetSize(), uint64(summary), isOneFile, &globalBM, &globalD, isCompressed)
		path = SSTable.GenerateSSTabel()
	} else {
		panic("Unknown type of Storage")
	}

	lsm := NewLSMTree(mt.maxEntry, &globalBM)
	lsm.AddLevel(path)
	lsm.StartMerging(mergeAlg, &globalD, isOneFile, uint64(summary), mergeSize)

}
