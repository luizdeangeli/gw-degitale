package cache

import (
	"sync"
	"time"
)

type item struct {
	provider string
	data     any
	expires  time.Time
}

// Memory guarda consultas bem-sucedidas até o TTL. Sem persistência.
type Memory struct {
	mu    sync.Mutex
	items map[string]item
	now   func() time.Time
}

func New() *Memory {
	return &Memory{
		items: map[string]item{},
		now:   time.Now,
	}
}

func (m *Memory) Get(key string) (provider string, data any, ok bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	it, found := m.items[key]
	if !found {
		return "", nil, false
	}
	if !m.now().Before(it.expires) {
		delete(m.items, key)
		return "", nil, false
	}
	return it.provider, it.data, true
}

func (m *Memory) Set(key, provider string, data any, ttl time.Duration) {
	if ttl <= 0 {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[key] = item{
		provider: provider,
		data:     data,
		expires:  m.now().Add(ttl),
	}
}
