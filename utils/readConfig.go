package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config is split by structures for better organization
type WALConfig struct {
	SegmentMaxBlocks int `json:"segmentMaxBlocks"`
}

type BlockManagerConfig struct {
	BlockSizeKB    int `json:"blockSizeKB"`
	CacheMaxBlocks int `json:"cacheMaxBlocks"`
}

type SSTableConfig struct {
	SummarySampling    int  `json:"summarySampling"`
	SeparateFiles      bool `json:"separateFiles"`
	CompressionEnabled bool `json:"compressionEnabled"`
}

type MemtableConfig struct {
	StructureType string `json:"structureType"`

	StructNum     int    `json:"structNum"`
	InstanceCount int    `json:"instanceCount"`
	MaxEntry      int    `json:"maxEntry"`
}

type CacheConfig struct {
	MaxEntries int `json:"maxEntries"`
}

type LSMTreeConfig struct {
	AlgorithmType int `json:"algorithmType"`
	ParamNumber   int `json:"paramNumber"`
	MaxLevel      int `json:"maxLevel"` // New field for max level
}

type TokenBucketConfig struct {
	Tokens           int `json:"tokens"`
	ResetIntervalSec int `json:"resetIntervalSec"`
}

// High level config that combines all components
// This is the main config structure that will be used throughout the application
type Config struct {
	WAL                WALConfig          `json:"wal"`
	BlockManager       BlockManagerConfig `json:"blockManager"`
	SSTable            SSTableConfig      `json:"sstable"`
	Memtable           MemtableConfig     `json:"memtable"`
	Cache              CacheConfig        `json:"cache"`
	LSMTree            LSMTreeConfig      `json:"lsm-tree"`
	TokenBucketRateLim TokenBucketConfig  `json:"tokenBucketRateLimit"`
}

// defaultConfig provides a fallback configuration
// This is used when the user config is invalid or missing
var defaultConfig = Config{
	WAL: WALConfig{
		SegmentMaxBlocks: 5,

	},
	BlockManager: BlockManagerConfig{
		BlockSizeKB:    8,
		CacheMaxBlocks: 1024,
	},
	SSTable: SSTableConfig{
		SummarySampling:    5,
		SeparateFiles:      true,
		CompressionEnabled: true,
	},
	Memtable: MemtableConfig{
		StructureType: "skiplist",
		StructNum:     2,
		InstanceCount: 5,
	},
	Cache: CacheConfig{
		MaxEntries: 10000,
	},
	LSMTree: LSMTreeConfig{
		AlgorithmType: 1,
		ParamNumber:   3,
	},
	TokenBucketRateLim: TokenBucketConfig{
		Tokens:           20,

		ResetIntervalSec: 30,
	},
}

// Read user config from config file
// If any validation fails, it will use the default config and return an error
// Therefore it will always return a valid config just may not be the user-provided one
func ReadConfig(filePath string) *Config {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening config file, using default config:", err)
		return &defaultConfig
	}
	defer file.Close()

	var userConfig Config

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&userConfig); err != nil {
		fmt.Println("Error decoding config file, using default config:", err)
		return &defaultConfig
	}

	if userConfig.Memtable.StructureType == "skiplist" {
		userConfig.Memtable.StructNum = 2
	} else if userConfig.Memtable.StructureType == "hashmap" {
		userConfig.Memtable.StructNum = 3
	} else if userConfig.Memtable.StructureType == "btree" {
		userConfig.Memtable.StructNum = 1
	} else {
		userConfig.Memtable.StructNum = defaultConfig.Memtable.StructNum // Use default if invalid
	}

	var errs []string

	// Validate WAL
	if userConfig.WAL.SegmentMaxBlocks <= 0 {
		errs = append(errs, fmt.Sprintf("wal.segmentMaxBlocks invalid (%d), defaulting to %d",
			userConfig.WAL.SegmentMaxBlocks, defaultConfig.WAL.SegmentMaxBlocks))
		userConfig.WAL.SegmentMaxBlocks = defaultConfig.WAL.SegmentMaxBlocks
	}

	// Validate BlockManager

	if userConfig.BlockManager.BlockSizeKB != 4 && userConfig.BlockManager.BlockSizeKB != 8 && userConfig.BlockManager.BlockSizeKB != 16 {
		errs = append(errs, fmt.Sprintf("blockManager.blockSizeKB invalid (%d), defaulting to %d",
			userConfig.BlockManager.BlockSizeKB, defaultConfig.BlockManager.BlockSizeKB))
		userConfig.BlockManager.BlockSizeKB = defaultConfig.BlockManager.BlockSizeKB
	}

	if userConfig.BlockManager.CacheMaxBlocks <= 0 {
		errs = append(errs, fmt.Sprintf("blockManager.cacheMaxBlocks invalid (%d), defaulting to %d",
			userConfig.BlockManager.CacheMaxBlocks, defaultConfig.BlockManager.CacheMaxBlocks))
		userConfig.BlockManager.CacheMaxBlocks = defaultConfig.BlockManager.CacheMaxBlocks
	}

	// Validate SSTable
	if userConfig.SSTable.SummarySampling <= 0 {
		errs = append(errs, fmt.Sprintf("sstable.summarySampling invalid (%d), defaulting to %d",
			userConfig.SSTable.SummarySampling, defaultConfig.SSTable.SummarySampling))
		userConfig.SSTable.SummarySampling = defaultConfig.SSTable.SummarySampling
	}
	// no validation needed for bools: SeparateFiles and CompressionEnabled

	// Validate Memtable
	if userConfig.Memtable.StructureType != "skiplist" && userConfig.Memtable.StructureType != "hashmap" && userConfig.Memtable.StructureType != "btree" {
		errs = append(errs, fmt.Sprintf("memtable.structureType invalid (%s), defaulting to %s",
			userConfig.Memtable.StructureType, defaultConfig.Memtable.StructureType))
		userConfig.Memtable.StructureType = defaultConfig.Memtable.StructureType
	}
	if userConfig.Memtable.InstanceCount <= 0 {
		errs = append(errs, fmt.Sprintf("memtable.instanceCount invalid (%d), defaulting to %d",
			userConfig.Memtable.InstanceCount, defaultConfig.Memtable.InstanceCount))
		userConfig.Memtable.InstanceCount = defaultConfig.Memtable.InstanceCount
	}
	userConfig.Memtable.MaxEntry++

	// Validate Cache
	if userConfig.Cache.MaxEntries <= 0 {
		errs = append(errs, fmt.Sprintf("cache.maxEntries invalid (%d), defaulting to %d",
			userConfig.Cache.MaxEntries, defaultConfig.Cache.MaxEntries))
		userConfig.Cache.MaxEntries = defaultConfig.Cache.MaxEntries
	}

	// Validate LSMTree
	if userConfig.LSMTree.AlgorithmType != 1 && userConfig.LSMTree.AlgorithmType != 2 {
		errs = append(errs, fmt.Sprintf("lsm-tree.algorithmType invalid (%d), defaulting to %d",
			userConfig.LSMTree.AlgorithmType, defaultConfig.LSMTree.AlgorithmType))
		userConfig.LSMTree.AlgorithmType = defaultConfig.LSMTree.AlgorithmType
	}
	if userConfig.LSMTree.ParamNumber <= 0 {
		errs = append(errs, fmt.Sprintf("lsm-tree.paramNumber invalid (%d), defaulting to %d",
			userConfig.LSMTree.ParamNumber, defaultConfig.LSMTree.ParamNumber))
		userConfig.LSMTree.ParamNumber = defaultConfig.LSMTree.ParamNumber
	}

	// Validate Token Bucket
	if userConfig.TokenBucketRateLim.Tokens <= 0 {
		errs = append(errs, fmt.Sprintf("tokenBucketRateLimit.tokens invalid (%d), defaulting to %d",
			userConfig.TokenBucketRateLim.Tokens, defaultConfig.TokenBucketRateLim.Tokens))
		userConfig.TokenBucketRateLim.Tokens = defaultConfig.TokenBucketRateLim.Tokens
	}
	if userConfig.TokenBucketRateLim.ResetIntervalSec <= 0 {
		errs = append(errs, fmt.Sprintf("tokenBucketRateLimit.resetIntervalSec invalid (%d), defaulting to %d",
			userConfig.TokenBucketRateLim.ResetIntervalSec, defaultConfig.TokenBucketRateLim.ResetIntervalSec))
		userConfig.TokenBucketRateLim.ResetIntervalSec = defaultConfig.TokenBucketRateLim.ResetIntervalSec
	}

	if len(errs) > 0 {
		fmt.Printf("Error(s) in config file:\n%s\n", stringJoin(errs, "\n"))
	}

	return &userConfig
}

// Helper join to print error messages for user
func stringJoin(strs []string, sep string) string {
	out := ""
	for i, s := range strs {
		if i > 0 {
			out += sep
		}
		out += s
	}
	return out
}
