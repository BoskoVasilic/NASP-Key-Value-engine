// Global variables and constants for the Key-Value engine

package structures

import (
	"KVengine/utils"
	"sync"
)

const (
	GlobalDictionaryFile = "./data/dictionary"
	GlobalLSMTreeFile    = "./data/tree_data"
)

var GlobalBM *BlockManager
var GlobalD *GlobalDictionary
var GlobalConfig *utils.Config
var GlobalWal *WriteAheadLog
var GlobalCache *Cache[string, Entry]
var GlobalLsm *LSMTree
var GlobalMemtables *Memtables
var GlobalTokenBucket *TokenBucket
var DicLock sync.RWMutex
var WalLock sync.Mutex
var BlockManagerLock sync.Mutex
var BlockCacheLock sync.Mutex
