// Package store caches provider availability results, mirroring the frontend
// providerCache: an entry is keyed "provider:label", carries a status, an
// optional detail and the time it was recorded, and goes stale after one hour.
package store

import (
	"sync"
	"time"

	"github.com/petermichon/domain-browser/server/internal/provider"
)

// Fresh is how long a cached result stays current, matching the frontend
// providerCache FRESH constant (60 * 60 * 1000 ms).
const Fresh = time.Hour

// Entry is a cached availability result and when it was recorded.
type Entry struct {
	Status provider.Status
	Detail string
	// URL is the page the result was emitted with, so a cache hit keeps the
	// right link when a provider reports a non-default one.
	URL string
	At  time.Time
}

// Key returns the cache key for a provider and label, matching the frontend
// cacheKey ("provider:label").
func Key(providerID, label string) string {
	return providerID + ":" + label
}

// IsFresh reports whether an entry recorded at At is still current at now for
// the given ttl, matching the frontend isFresh (now - at < ttl).
func IsFresh(e Entry, now time.Time, ttl time.Duration) bool {
	return now.Sub(e.At) < ttl
}

// Store caches the availability result for provider:label pairs.
type Store interface {
	// Get returns the fresh entry for the pair. A missing or stale entry
	// reports ok=false and is dropped from the store.
	Get(providerID, label string) (Entry, bool)
	// Set records the status, detail and URL for the pair, stamped with the
	// store's current time.
	Set(providerID, label string, status provider.Status, detail, url string)
	// Delete removes the entry for the pair.
	Delete(providerID, label string)
	// Clear removes every entry.
	Clear()
}

// Memory is a concurrency-safe in-memory Store.
type Memory struct {
	mu      sync.RWMutex
	now     func() time.Time
	ttl     time.Duration
	entries map[string]Entry
}

// NewMemory returns an empty in-memory store with the default one-hour TTL.
func NewMemory() *Memory {
	return newMemory(Fresh, time.Now)
}

// NewMemoryWithTTL returns an empty in-memory store with a custom ttl and clock.
// It lets tests exercise expiry deterministically.
func NewMemoryWithTTL(ttl time.Duration, now func() time.Time) *Memory {
	return newMemory(ttl, now)
}

func newMemory(ttl time.Duration, now func() time.Time) *Memory {
	return &Memory{now: now, ttl: ttl, entries: make(map[string]Entry)}
}

// Get returns the fresh entry for the pair, dropping it when stale.
func (m *Memory) Get(providerID, label string) (Entry, bool) {
	key := Key(providerID, label)

	m.mu.RLock()
	entry, ok := m.entries[key]
	m.mu.RUnlock()
	if !ok {
		return Entry{}, false
	}

	if !IsFresh(entry, m.now(), m.ttl) {
		m.mu.Lock()
		delete(m.entries, key)
		m.mu.Unlock()
		return Entry{}, false
	}
	return entry, true
}

// Set records the status, detail and URL for the pair, stamped with the
// store's current time.
func (m *Memory) Set(providerID, label string, status provider.Status, detail, url string) {
	m.mu.Lock()
	m.entries[Key(providerID, label)] = Entry{Status: status, Detail: detail, URL: url, At: m.now()}
	m.mu.Unlock()
}

// Delete removes the entry for the pair.
func (m *Memory) Delete(providerID, label string) {
	m.mu.Lock()
	delete(m.entries, Key(providerID, label))
	m.mu.Unlock()
}

// Clear removes every entry.
func (m *Memory) Clear() {
	m.mu.Lock()
	m.entries = make(map[string]Entry)
	m.mu.Unlock()
}

var _ Store = (*Memory)(nil)
