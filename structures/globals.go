// Global variables and constants for the Key-Value engine

package structures

import (
	"KVengine/utils"
	"sync"
)

var globalBM *BlockManager
var globalD GlobalDictionary
var globalConfig *utils.Config
var globalWal WriteAheadLog
var globalCache Cache[string, Entry]
var globalLsm *LSMTree
var dicLock sync.RWMutex
var walLock sync.Mutex
