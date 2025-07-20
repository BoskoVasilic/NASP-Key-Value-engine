package utils

import (
    "encoding/json"
    "fmt"
    "os"
)

type Config struct {
	WALSegmentSize    int    `json:"WALSegmentSize"`
	BlockSizeKB       int    `json:"blockSizeKB"`
	BlockCacheSize    int    `json:"blockCacheSize"`
	Compression       string `json:"compression"`
	SSTableFiles      string `json:"SSTableFiles"`
	MemtableStrut     string `json:"MemtableStrut"`
	MaxMemtableSize   int    `json:"MaxMemtableSize"`
	MemtableInstanceNum int  `json:"MemtableInstanceNum"`
}

var defaultConfig = Config{
    WALSegmentSize:    10,
    BlockSizeKB:       4,
    BlockCacheSize:    2,
    Compression:       "Yes",
    SSTableFiles:      "",
    MemtableStrut:     "SkipList",
    MaxMemtableSize:   30,
    MemtableInstanceNum: 5,
}


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

	var errMsgs []string

	if userConfig.WALSegmentSize <= 0 {
		errMsgs = append(errMsgs, fmt.Sprintf("WALSegmentSize invalid (%d), set to default (%d)", userConfig.WALSegmentSize, defaultConfig.WALSegmentSize))
		userConfig.WALSegmentSize = defaultConfig.WALSegmentSize
	}
	if userConfig.BlockSizeKB <= 0 {
		errMsgs = append(errMsgs, fmt.Sprintf("BlockSizeKB invalid (%d), set to default (%d)", userConfig.BlockSizeKB, defaultConfig.BlockSizeKB))
		userConfig.BlockSizeKB = defaultConfig.BlockSizeKB
	}
	if userConfig.BlockCacheSize <= 0 {
		errMsgs = append(errMsgs, fmt.Sprintf("BlockCacheSize invalid (%d), set to default (%d)", userConfig.BlockCacheSize, defaultConfig.BlockCacheSize))
		userConfig.BlockCacheSize = defaultConfig.BlockCacheSize
	}
	if userConfig.MaxMemtableSize <= 0 {
		errMsgs = append(errMsgs, fmt.Sprintf("MaxMemtableSize invalid (%d), set to default (%d)", userConfig.MaxMemtableSize, defaultConfig.MaxMemtableSize))
		userConfig.MaxMemtableSize = defaultConfig.MaxMemtableSize
	}
	if userConfig.MemtableInstanceNum <= 0 {
		errMsgs = append(errMsgs, fmt.Sprintf("MemtableInstanceNum invalid (%d), set to default (%d)", userConfig.MemtableInstanceNum, defaultConfig.MemtableInstanceNum))
		userConfig.MemtableInstanceNum = defaultConfig.MemtableInstanceNum
	}
	if userConfig.Compression != "Yes" && userConfig.Compression != "No" {
		errMsgs = append(errMsgs, fmt.Sprintf("Compression invalid (%s), set to default (%s)", userConfig.Compression, defaultConfig.Compression))
		userConfig.Compression = defaultConfig.Compression
	}
	if userConfig.MemtableStrut != "SkipList" && userConfig.MemtableStrut != "Other" {
		errMsgs = append(errMsgs, fmt.Sprintf("MemtableStrut invalid (%s), set to default (%s)", userConfig.MemtableStrut, defaultConfig.MemtableStrut))
		userConfig.MemtableStrut = defaultConfig.MemtableStrut
	}
	if userConfig.SSTableFiles == "" {
		errMsgs = append(errMsgs, fmt.Sprintf("SSTableFiles empty, set to default (%s)", defaultConfig.SSTableFiles))
		userConfig.SSTableFiles = defaultConfig.SSTableFiles
	}

	var finalErr error
	if len(errMsgs) > 0 {
		finalErr = fmt.Errorf("config validation issues:\n%s", 
			fmt.Sprintf("%s", stringJoin(errMsgs, "\n")))
	}

	return &userConfig, finalErr
}

// Helper function for joining strings (since strings.Join needs []string)
func stringJoin(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}