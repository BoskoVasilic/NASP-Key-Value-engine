package main

import (
	"KVengine/cli"
	"KVengine/structures"
	"KVengine/utils"
)

func initGlobals() {
	structures.GlobalBM = structures.NewBlockManager(structures.GlobalConfig.BlockManager.CacheMaxBlocks)
	structures.GlobalD = structures.GetDicteneryFormFile(structures.GlobalDictionaryFile, structures.GlobalBM)
	structures.GlobalConfig = utils.ReadConfig("config.json")
	structures.GlobalWal = structures.NewWriteAheadLog(structures.GlobalBM)
	structures.GlobalWal.InitializeWriteAheadLog()
	structures.GlobalCache = structures.NewCache[string, structures.Entry](structures.GlobalConfig.Cache.MaxEntries)
	structures.GlobalLsm = structures.DeserializeLSM(structures.GlobalLSMTreeFile, structures.GlobalBM)
	structures.GlobalMemtables = structures.NewMemtables(structures.GlobalConfig.Memtable.StructNum, structures.GlobalConfig.Memtable.InstanceCount)
}

func loadFromWal() {
	status := true
	for status {
		entry, status := structures.GlobalWal.GetWalElementByElement()
		if status {
			structures.GlobalMemtables.Put(entry.GetKey(), entry.GetValue())
		}
	}
}

func main() {
	initGlobals()
	loadFromWal()
	cli.Run()
}
