package main

import (
	"KVengine/cli"
	"KVengine/structures"
	"KVengine/utils"
)

func initGlobals() {
	structures.GlobalConfig = utils.ReadConfig("config.json")
	structures.GlobalBM = structures.NewBlockManager(structures.GlobalConfig.BlockManager.CacheMaxBlocks)
	structures.GlobalD = structures.GetDicteneryFormFile(structures.GlobalDictionaryFile, structures.GlobalBM)
	structures.GlobalWal = structures.NewWriteAheadLog(structures.GlobalBM)
	structures.GlobalWal.InitializeWriteAheadLog()
	structures.GlobalCache = structures.NewCache[string, structures.Entry](structures.GlobalConfig.Cache.MaxEntries)
	structures.GlobalLsm = structures.DeserializeLSM(structures.GlobalLSMTreeFile, structures.GlobalBM)
	structures.GlobalMemtables = structures.NewMemtables(structures.GlobalConfig.Memtable.InstanceCount, structures.GlobalConfig.Memtable.MaxEntry)
}

// func loadFromWal() {
// 	status := true
// 	for status {
// 		fmt.Println("Loading entry from WAL:")
// 		entry, status := structures.GlobalWal.GetWalElementByElement()
// 		fmt.Println(entry, status)
// 		if status {
// 			structures.GlobalMemtables.PutWithoutWal(entry.GetKey(), entry.GetValue())
// 		} else {
// 			break
// 		}
// 	}
// }

func main() {
	initGlobals()
	// loadFromWal()
	structures.GlobalMemtables.Init()
	cli.Run()
}
