package middleware

import (
	"math"
	"net"
	"net/http"
	"sync"
	"time"
)

type peerBucket struct {
	tokens  float64
	updated time.Time
}

type peerRateLimiter struct {
	mu            sync.Mutex
	buckets       map[string]*peerBucket
	tokensPerNano float64
	burst         float64
	idleTTL       time.Duration
	now           func() time.Time
	requests      uint64
}

// RateLimit applies an in-process token bucket per direct TCP peer. Set either
// value to zero to disable it in explicitly selected non-production modes.
func RateLimit(next http.Handler, requestsPerMinute int64, burst int64) http.Handler {
	if requestsPerMinute <= 0 || burst <= 0 {
		return next
	}
	limiter := &peerRateLimiter{
		buckets:       make(map[string]*peerBucket),
		tokensPerNano: float64(requestsPerMinute) / float64(time.Minute),
		burst:         float64(burst),
		idleTTL:       max(2*time.Minute, time.Duration(math.Ceil(float64(burst)/float64(requestsPerMinute)*float64(time.Minute)))),
		now:           time.Now,
	}
	return limiter.wrap(next)
}

func (limiter *peerRateLimiter) wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.allow(peerAddress(r.RemoteAddr)) {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (limiter *peerRateLimiter) allow(peer string) bool {
	now := limiter.now()
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	limiter.requests++
	if limiter.requests%256 == 0 {
		for key, bucket := range limiter.buckets {
			if now.Sub(bucket.updated) > limiter.idleTTL {
				delete(limiter.buckets, key)
			}
		}
	}

	bucket, ok := limiter.buckets[peer]
	if !ok {
		limiter.buckets[peer] = &peerBucket{tokens: limiter.burst - 1, updated: now}
		return true
	}

	elapsed := now.Sub(bucket.updated)
	bucket.tokens = math.Min(limiter.burst, bucket.tokens+float64(elapsed)*limiter.tokensPerNano)
	bucket.updated = now
	if bucket.tokens < 1 {
		return false
	}
	bucket.tokens--
	return true
}

func peerAddress(remoteAddress string) string {
	host, _, err := net.SplitHostPort(remoteAddress)
	if err == nil && host != "" {
		return host
	}
	if remoteAddress == "" {
		return "unknown"
	}
	return remoteAddress
}
