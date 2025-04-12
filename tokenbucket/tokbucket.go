// This package implements a simple yet effective token bucket rate limiting algorithm.
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

// refills the bucket with discrete amount of tokens.
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

// Limiter is a middleware which initializes a Token Bucket with provided tokens limit and rate of token consumption.
//
// If all the tokens are consumed prior the rate limit, it drops subsequent requests untill next refill.
func Limiter(next http.Handler, tokens int, rate time.Duration) http.Handler {
	tb := NewTokenBucket(tokens, rate)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !tb.allow() {
			http.Error(w, "Too Many Requests!", http.StatusTooManyRequests)
			http.Header.Add(r.Header, "X-RequestThrottled", "true")
			return
		}

		next.ServeHTTP(w, r)
	})
}
