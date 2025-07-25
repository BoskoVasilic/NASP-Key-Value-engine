package structures

// import (
// 	"encoding/binary"
// 	"time"
// )

// const (
// 	MAX_TOKENS      = 5  // get from config file
// 	RESET_INTERVAL  = 60 // get from config file
// 	TOKENS_SIZE     = 4
// 	RESET_TIME_SIZE = 8
// )

// type TokenBucket struct {
// 	tokensAvailable int
// 	lastResetTime   int64
// }

// func NewTokenBucket() *TokenBucket {
// 	return &TokenBucket{
// 		tokensAvailable: 0,
// 		lastResetTime:   0,
// 	}
// }

// func (tb *TokenBucket) Initialize() {
// 	tb.tokensAvailable, tb.lastResetTime = tb.GetBucketState()
// }

// func (tb *TokenBucket) Reset() {
// 	currentTime := time.Now().Unix()
// 	if currentTime-tb.lastResetTime >= RESET_INTERVAL {
// 		tb.tokensAvailable = MAX_TOKENS
// 		tb.lastResetTime = currentTime
// 	}
// }

// func (tb *TokenBucket) CheckAvailability() bool {
// 	tb.Reset()
// 	if tb.tokensAvailable >= 0 {
// 		tb.tokensAvailable -= 1
// 		tb.SaveBucketState()
// 		return true
// 	}
// 	return false
// }

// func (tb *TokenBucket) GetBucketState() (int, int64) {
// 	data, _ := Get("token_bucket")
// 	tb.tokensAvailable = int(binary.LittleEndian.Uint32(data[:TOKENS_SIZE]))
// 	tb.lastResetTime = int64(binary.LittleEndian.Uint64(data[TOKENS_SIZE : RESET_TIME_SIZE+TOKENS_SIZE]))
// 	return tb.tokensAvailable, tb.lastResetTime
// }

// func (tb *TokenBucket) SaveBucketState() {
// 	data := make([]byte, TOKENS_SIZE+RESET_TIME_SIZE)
// 	binary.LittleEndian.PutUint32(data[:TOKENS_SIZE], uint32(tb.tokensAvailable))
// 	binary.LittleEndian.PutUint64(data[TOKENS_SIZE:], uint64(tb.lastResetTime))
// 	Put("token_bucket", data)
// }
