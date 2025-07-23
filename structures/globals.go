// Global variables and constants for the Key-Value engine

package structures

import (
	"KVengine/utils"
	"sync"
)

var GlobalBM *BlockManager
var GlobalD GlobalDictionary
var GlobalConfig *utils.Config
var GlobalWal WriteAheadLog
var GlobalCache Cache[string, Entry]
var GlobalLsm *LSMTree
var dicLock sync.RWMutex
var walLock sync.Mutex
