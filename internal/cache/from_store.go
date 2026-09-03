package cache

import "github.com/Danish-13-Khan/backend-cabinet/internal/store"

// FromSet builds a Cache from the datastore set. Noop when Redis is disabled.
func FromSet(s *store.Set) Cache {
	if s == nil {
		return Noop{}
	}
	return FromClient(s.RedisClient())
}
