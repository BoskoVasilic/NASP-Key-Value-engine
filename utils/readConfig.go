package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config is split by structures for better organization
type WALConfig struct {
	SegmentMaxRecords int `json:"segmentMaxRecords"`
	BlockSizeKB       int `json:"blockSizeKB"`
}

type BlockManagerConfig struct {
	BlockSizeKB    int `json:"blockSizeKB"`
	CacheMaxBlocks int `json:"cacheMaxBlocks"`
}

type SSTableConfig struct {
	KeyCompressionEnabled bool   `json:"keyCompressionEnabled"`
	Layout                string `json:"layout"`
	SummarySampling       int    `json:"summarySampling"`
}

type MemtableConfig struct {
	StructureType string `json:"structureType"`
	MaxElements   int    `json:"maxElements"`
	InstanceCount int    `json:"instanceCount"`
}

type CacheConfig struct {
	MaxEntries int `json:"maxEntries"`
}

type RateLimitConfig struct {
	Tokens          int `json:"tokens"`
	ResetIntervalMs int `json:"resetIntervalMs"`
}

// High level config that combines all components
// This is the main config structure that will be used throughout the application
type Config struct {
	WAL          WALConfig          `json:"wal"`
	BlockManager BlockManagerConfig `json:"blockManager"`
	SSTable      SSTableConfig      `json:"sstable"`
	Memtable     MemtableConfig     `json:"memtable"`
	Cache        CacheConfig        `json:"cache"`
	RateLimit    RateLimitConfig    `json:"rateLimit"`
}

// defaultConfig provides a fallback configuration 
// This is used when the user config is invalid or missing
var defaultConfig = Config{
	WAL: WALConfig{
		SegmentMaxRecords: 10,
		BlockSizeKB:       4,
	},
	BlockManager: BlockManagerConfig{
		BlockSizeKB:    8,
		CacheMaxBlocks: 1024,
	},
	SSTable: SSTableConfig{
		KeyCompressionEnabled: true,
		Layout:                "separate",
		SummarySampling:       5,
	},
	Memtable: MemtableConfig{
		StructureType: "skiplist",
		MaxElements:   30,
		InstanceCount: 5,
	},
	Cache: CacheConfig{
		MaxEntries: 10000,
	},
	RateLimit: RateLimitConfig{
		Tokens:          10,
		ResetIntervalMs: 16,
	},
}

// Read user config from config file
// If any validation fails, it will use the default config and return an error
// Therefore it will always return a valid config just may not be the user-provided one

func ReadConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return &defaultConfig, fmt.Errorf("error opening config file, using default: %v", err)
	}
	defer file.Close()

	var userConfig Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&userConfig); err != nil {
		return &defaultConfig, fmt.Errorf("error decoding config file, using default: %v", err)
	}

	var errs []string

	// Validate WAL
	if userConfig.WAL.SegmentMaxRecords <= 0 {
		errs = append(errs, fmt.Sprintf("WAL.segmentMaxRecords invalid (%d), defaulting to %d",
			userConfig.WAL.SegmentMaxRecords, defaultConfig.WAL.SegmentMaxRecords))
		userConfig.WAL.SegmentMaxRecords = defaultConfig.WAL.SegmentMaxRecords
	}
	if userConfig.WAL.BlockSizeKB != 4 && userConfig.WAL.BlockSizeKB != 8 && userConfig.WAL.BlockSizeKB != 16 {
		errs = append(errs, fmt.Sprintf("WAL.blockSizeKB invalid (%d), defaulting to %d",
			userConfig.WAL.BlockSizeKB, defaultConfig.WAL.BlockSizeKB))
		userConfig.WAL.BlockSizeKB = defaultConfig.WAL.BlockSizeKB
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
	if userConfig.SSTable.Layout != "separate" && userConfig.SSTable.Layout != "combined" {
		errs = append(errs, fmt.Sprintf("sstable.layout invalid (%s), defaulting to %s",
			userConfig.SSTable.Layout, defaultConfig.SSTable.Layout))
		userConfig.SSTable.Layout = defaultConfig.SSTable.Layout
	}
	if userConfig.SSTable.SummarySampling <= 0 {
		errs = append(errs, fmt.Sprintf("sstable.summarySampling invalid (%d), defaulting to %d",
			userConfig.SSTable.SummarySampling, defaultConfig.SSTable.SummarySampling))
		userConfig.SSTable.SummarySampling = defaultConfig.SSTable.SummarySampling
	}

	// Validate Memtable
	if userConfig.Memtable.StructureType != "skiplist" && userConfig.Memtable.StructureType != "hashmap" && userConfig.Memtable.StructureType != "bst" {
		errs = append(errs, fmt.Sprintf("memtable.structureType invalid (%s), defaulting to %s",
			userConfig.Memtable.StructureType, defaultConfig.Memtable.StructureType))
		userConfig.Memtable.StructureType = defaultConfig.Memtable.StructureType
	}
	if userConfig.Memtable.MaxElements <= 0 {
		errs = append(errs, fmt.Sprintf("memtable.maxElements invalid (%d), defaulting to %d",
			userConfig.Memtable.MaxElements, defaultConfig.Memtable.MaxElements))
		userConfig.Memtable.MaxElements = defaultConfig.Memtable.MaxElements
	}
	if userConfig.Memtable.InstanceCount <= 0 {
		errs = append(errs, fmt.Sprintf("memtable.instanceCount invalid (%d), defaulting to %d",
			userConfig.Memtable.InstanceCount, defaultConfig.Memtable.InstanceCount))
		userConfig.Memtable.InstanceCount = defaultConfig.Memtable.InstanceCount
	}

	// Validate Cache
	if userConfig.Cache.MaxEntries <= 0 {
		errs = append(errs, fmt.Sprintf("cache.maxEntries invalid (%d), defaulting to %d",
			userConfig.Cache.MaxEntries, defaultConfig.Cache.MaxEntries))
		userConfig.Cache.MaxEntries = defaultConfig.Cache.MaxEntries
	}

	// Validate RateLimit
	if userConfig.RateLimit.Tokens <= 0 {
		errs = append(errs, fmt.Sprintf("rateLimit.tokens invalid (%d), defaulting to %d",
			userConfig.RateLimit.Tokens, defaultConfig.RateLimit.Tokens))
		userConfig.RateLimit.Tokens = defaultConfig.RateLimit.Tokens
	}
	if userConfig.RateLimit.ResetIntervalMs <= 0 {
		errs = append(errs, fmt.Sprintf("rateLimit.resetIntervalMs invalid (%d), defaulting to %d",
			userConfig.RateLimit.ResetIntervalMs, defaultConfig.RateLimit.ResetIntervalMs))
		userConfig.RateLimit.ResetIntervalMs = defaultConfig.RateLimit.ResetIntervalMs
	}

	var finalErr error
	if len(errs) > 0 {
		finalErr = fmt.Errorf("config validation issues:\n%s", stringJoin(errs, "\n"))
	}

	return &userConfig, finalErr
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
