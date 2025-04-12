package tokenbucket

import (
	"net/http"
	"sync"
	"time"
)

type TokenBucket struct {
	mu           sync.Mutex
	tokens       int
	maxTokens    int
	refillRate   time.Duration
	lastRefilled time.Time
}

func NewTokenBucket(maxTokens int, refillRate time.Duration) *TokenBucket {
	return &TokenBucket{
		tokens:       maxTokens,
		maxTokens:    maxTokens,
		refillRate:   refillRate,
		lastRefilled: time.Now(),
	}
}

func (tb *TokenBucket) refill() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefilled)

	refillCount := int(elapsed / tb.refillRate)
	if refillCount > 0 {
		tb.tokens = min(tb.tokens+refillCount, tb.maxTokens)
		tb.lastRefilled = now
	}
}

func (tb *TokenBucket) allow() bool {
	tb.refill()

	tb.mu.Lock()
	defer tb.mu.Unlock()
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}

func Limiter(next http.Handler) http.Handler {
	tb := NewTokenBucket(4, time.Minute)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !tb.allow() {
			http.Error(w, "Too Many Requests!", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
