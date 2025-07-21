package structures

import "time"

const (
	MAX_TOKENS     = 5  // get from config file
	RESET_INTERVAL = 60 // get from config file
)

type TokenBucket struct {
	tokensAvailable int
	lastResetTime   int64
}

func NewTokenBucket() *TokenBucket {
	return &TokenBucket{
		tokensAvailable: 0,
		lastResetTime:   0,
	}
}

func (tb *TokenBucket) Initialize() {
	tb.tokensAvailable, tb.lastResetTime = tb.GetBucketState()
}

func (tb *TokenBucket) Reset() {
	currentTime := time.Now().Unix()
	if currentTime-tb.lastResetTime >= RESET_INTERVAL {
		tb.tokensAvailable = MAX_TOKENS
		tb.lastResetTime = currentTime
	}
}

func (tb *TokenBucket) CheckAvailability() bool {
	tb.Reset()
	if tb.tokensAvailable >= 0 {
		tb.tokensAvailable -= 1
		tb.SaveBucketState()
		return true
	}
	return false
}

func (tb *TokenBucket) GetBucketState() (int, int64) {
	return tb.tokensAvailable, tb.lastResetTime
}

func (tb *TokenBucket) SaveBucketState() error {
	return nil
}
