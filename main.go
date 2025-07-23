package main

import (
	"KVengine/cli"
	"KVengine/utils"
	"KVengine/structures"
)

func initGlobals() {
	structures.GlobalConfig = utils.ReadConfig("config.json")
	structures.GlobalBM = structures.NewBlockManager(structures.GlobalConfig.BlockManager.CacheMaxBlocks)
	structures.GlobalD = structures.NewGlobalDictionary("globalDictionary.bin", structures.GlobalBM)
	structures.GlobalCache = structures.NewCache[string, []byte](structures.GlobalConfig.Cache.MaxEntries)
}

func main() {
	cli.Run()
}

