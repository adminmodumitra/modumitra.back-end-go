package middleware

import (
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// EngineQuota meters /v1/engine/* calls (roadmap 4.7: nobody should be able to
// fit the pricing formula from thousands of requests). Per user and per org
// per minute, plus a daily cap per org that resets at UTC midnight.
//
// Like the Phase 0 limiters, all counters live in process memory, so each API
// instance meters on its own. With more than one instance the limits multiply;
// move these counters to Redis (INCR + EXPIRE) before scaling out.
type EngineQuota struct {
	user     *keyedLimiter
	org      *keyedLimiter
	userWait time.Duration
	orgWait  time.Duration
	perDay   int

	mu    sync.Mutex
	day   string
	daily map[string]int
	now   func() time.Time
}

// NewEngineQuota takes per-minute budgets (also the burst) and a daily cap.
func NewEngineQuota(userPerMin, orgPerMin, orgPerDay int) *EngineQuota {
	q := &EngineQuota{
		user:     newKeyedLimiter(rate.Limit(float64(userPerMin)/60), userPerMin),
		org:      newKeyedLimiter(rate.Limit(float64(orgPerMin)/60), orgPerMin),
		userWait: refill(userPerMin),
		orgWait:  refill(orgPerMin),
		perDay:   orgPerDay,
		daily:    map[string]int{},
		now:      time.Now,
	}
	q.user.now = func() time.Time { return q.now() }
	q.org.now = func() time.Time { return q.now() }
	return q
}

// refill is how long one token takes to come back at perMin per minute.
func refill(perMin int) time.Duration {
	if perMin <= 0 {
		return time.Minute
	}
	return time.Duration(math.Ceil(float64(time.Minute) / float64(perMin)))
}

// Allow charges one call to userID and orgKey. On refusal it returns how long
// the caller should wait. A call refused by any budget is not counted daily.
func (q *EngineQuota) Allow(userID, orgKey string) (bool, time.Duration) {
	if !q.user.allow(userID) {
		return false, q.userWait
	}
	if !q.org.allow(orgKey) {
		return false, q.orgWait
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	now := q.now().UTC()
	if day := now.Format("2006-01-02"); day != q.day {
		q.day = day
		q.daily = map[string]int{}
	}
	if q.daily[orgKey] >= q.perDay {
		midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
		return false, midnight.Sub(now)
	}
	q.daily[orgKey]++
	return true, 0
}

// RejectQuota answers 429 quota_exceeded with a whole-second Retry-After.
func RejectQuota(c *gin.Context, wait time.Duration) {
	secs := int(math.Ceil(wait.Seconds()))
	if secs < 1 {
		secs = 1
	}
	c.Header(utils.HeaderRetryAfter, strconv.Itoa(secs))
	Abort(c, utils.E(utils.CodeQuotaExceeded).WithDetails(map[string]any{"retry_after_seconds": secs}))
}
