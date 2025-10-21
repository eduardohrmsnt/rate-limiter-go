package storage

import (
	"context"
	"sync"
	"time"
)

type entry struct {
	value     int64
	expiresAt time.Time
}

type MemoryStorage struct {
	data map[string]*entry
	mu   sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	storage := &MemoryStorage{
		data: make(map[string]*entry),
	}

	go storage.cleanupExpired()

	return storage
}

func (m *MemoryStorage) cleanupExpired() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for key, entry := range m.data {
			if !entry.expiresAt.IsZero() && now.After(entry.expiresAt) {
				delete(m.data, key)
			}
		}
		m.mu.Unlock()
	}
}

func (m *MemoryStorage) Increment(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	if entry, exists := m.data[key]; exists {
		if !entry.expiresAt.IsZero() && now.After(entry.expiresAt) {
			delete(m.data, key)
		}
	}

	if entry, exists := m.data[key]; exists {
		entry.value++
		return entry.value, nil
	}

	m.data[key] = &entry{
		value:     1,
		expiresAt: now.Add(expiration),
	}

	return 1, nil
}

func (m *MemoryStorage) Get(ctx context.Context, key string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.data[key]
	if !exists {
		return 0, nil
	}

	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		return 0, nil
	}

	return entry.value, nil
}

func (m *MemoryStorage) SetBlock(ctx context.Context, key string, duration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = &entry{
		value:     1,
		expiresAt: time.Now().Add(duration),
	}

	return nil
}

func (m *MemoryStorage) IsBlocked(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.data[key]
	if !exists {
		return false, nil
	}

	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		return false, nil
	}

	return entry.value == 1, nil
}

func (m *MemoryStorage) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.data[key]
	if !exists {
		return 0, nil
	}

	if entry.expiresAt.IsZero() {
		return 0, nil
	}

	ttl := time.Until(entry.expiresAt)
	if ttl < 0 {
		return 0, nil
	}

	return ttl, nil
}

func (m *MemoryStorage) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data = make(map[string]*entry)
	return nil
}
