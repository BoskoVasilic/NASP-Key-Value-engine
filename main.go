package main

import (
	"KVengine/cli"
	"KVengine/utils"
	"KVengine/structures"
)

func initGlobals() {
	globalConfig := utils.ReadConfig("config.json")
	globalBM := structures.NewBlockManager(globalConfig.BlockManager.CacheMaxBlocks)
	globalD := structures.NewGlobalDictionary("globalDictionary.bin", globalBM)
	globalCache := structures.NewCache[string, []byte](globalConfig.Cache.MaxEntries)
}

func main() {
	cli.Run()
}

