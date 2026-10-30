package middleware

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// limiterIdleTTL drops a key's bucket after this long without a request, so the
// map cannot grow without bound. Every limit here refills fully well inside it.
const limiterIdleTTL = 10 * time.Minute

type bucket struct {
	lim  *rate.Limiter
	seen time.Time
}

// keyedLimiter keeps one token bucket per key (client IP, user id).
// Buckets live in process memory; each API instance limits on its own.
type keyedLimiter struct {
	mu        sync.Mutex
	rps       rate.Limit
	burst     int
	buckets   map[string]*bucket
	lastSweep time.Time
	now       func() time.Time
}

func newKeyedLimiter(rps rate.Limit, burst int) *keyedLimiter {
	return &keyedLimiter{rps: rps, burst: burst, buckets: map[string]*bucket{}, now: time.Now}
}

func (k *keyedLimiter) allow(key string) bool {
	k.mu.Lock()
	defer k.mu.Unlock()
	now := k.now()
	if now.Sub(k.lastSweep) > limiterIdleTTL {
		for id, b := range k.buckets {
			if now.Sub(b.seen) > limiterIdleTTL {
				delete(k.buckets, id)
			}
		}
		k.lastSweep = now
	}
	b, ok := k.buckets[key]
	if !ok {
		b = &bucket{lim: rate.NewLimiter(k.rps, k.burst)}
		k.buckets[key] = b
	}
	b.seen = now
	return b.lim.AllowN(now, 1)
}

func (k *keyedLimiter) size() int {
	k.mu.Lock()
	defer k.mu.Unlock()
	return len(k.buckets)
}

func rejectRateLimited(c *gin.Context, burst int) {
	c.Header("Retry-After", "1")
	c.Header("X-RateLimit-Limit", strconv.Itoa(burst))
	c.Header("X-RateLimit-Remaining", "0")
	c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Second).Unix(), 10))
	Abort(c, utils.E(utils.CodeRateLimited))
}
