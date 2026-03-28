package cache

import "time"

type entry struct {
	value     any
	expiresAt time.Time
	fetchedAt time.Time
}

type TTLCache struct {
	m map[string]entry
}

func New() *TTLCache {
	return &TTLCache{
		m: make(map[string]entry),
	}
}

func (c *TTLCache) Get(key string) (value any, fetchedAt time.Time, ok bool) {

	v, ok := c.m[key]
	if !ok {
		return nil, time.Time{}, false
	}

	if v.expiresAt.After(time.Now()) {
		return v.value, v.fetchedAt, true
	}

	delete(c.m, key)

	return nil, time.Time{}, false
}

func (c *TTLCache) Set(key string, value any, ttl time.Duration) {
	c.m[key] = entry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
		fetchedAt: time.Now(),
	}
}
