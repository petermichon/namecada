package store

import (
	"sync"
	"testing"
	"time"

	"github.com/petermichon/domain-browser/backend/internal/provider"
)

func TestKey(t *testing.T) {
	if got := Key("npm", "acema"); got != "npm:acema" {
		t.Errorf("Key = %q, want %q", got, "npm:acema")
	}
}

func TestIsFresh(t *testing.T) {
	at := time.Unix(1_000, 0)
	tests := []struct {
		name string
		now  time.Time
		ttl  time.Duration
		want bool
	}{
		{"within ttl", time.Unix(2_000, 0), time.Hour, true},
		{"past ttl", time.Unix(5_000, 0), time.Hour, false},
		{"exactly at ttl", time.Unix(4_600, 0), time.Hour, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsFresh(Entry{Status: provider.StatusTaken, At: at}, tt.now, tt.ttl)
			if got != tt.want {
				t.Errorf("IsFresh = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemorySetGet(t *testing.T) {
	s := NewMemory()
	s.Set("npm", "acema", provider.StatusAvailable, "")

	got, ok := s.Get("npm", "acema")
	if !ok {
		t.Fatal("Get(npm, acema) reported a miss")
	}
	if got.Status != provider.StatusAvailable {
		t.Errorf("Status = %q, want %q", got.Status, provider.StatusAvailable)
	}
	if _, ok := s.Get("github", "acema"); ok {
		t.Error("Get(github, acema) unexpectedly hit")
	}
}

func TestMemoryKeepsDetail(t *testing.T) {
	s := NewMemory()
	s.Set("domain.com", "acema", provider.StatusTaken, "registered 1997-09-15")

	got, ok := s.Get("domain.com", "acema")
	if !ok {
		t.Fatal("Get reported a miss")
	}
	if got.Status != provider.StatusTaken || got.Detail != "registered 1997-09-15" {
		t.Errorf("entry = %+v, want taken/registered 1997-09-15", got)
	}
}

func TestMemoryExpires(t *testing.T) {
	now := time.Unix(1_000, 0)
	s := NewMemoryWithTTL(time.Hour, func() time.Time { return now })
	s.Set("npm", "acema", provider.StatusAvailable, "")

	now = now.Add(59 * time.Minute)
	if _, ok := s.Get("npm", "acema"); !ok {
		t.Fatal("entry expired before the ttl")
	}

	now = now.Add(2 * time.Minute)
	if _, ok := s.Get("npm", "acema"); ok {
		t.Fatal("entry survived past the ttl")
	}
	// A dropped stale entry must not resurface.
	if _, ok := s.Get("npm", "acema"); ok {
		t.Fatal("stale entry resurfaced after being dropped")
	}
}

func TestMemoryDelete(t *testing.T) {
	s := NewMemory()
	s.Set("npm", "acema", provider.StatusAvailable, "")
	s.Delete("npm", "acema")
	if _, ok := s.Get("npm", "acema"); ok {
		t.Fatal("entry survived Delete")
	}
}

func TestMemoryClear(t *testing.T) {
	s := NewMemory()
	s.Set("npm", "acema", provider.StatusAvailable, "")
	s.Set("github", "acema", provider.StatusTaken, "")
	s.Clear()
	if _, ok := s.Get("npm", "acema"); ok {
		t.Error("npm entry survived Clear")
	}
	if _, ok := s.Get("github", "acema"); ok {
		t.Error("github entry survived Clear")
	}
}

func TestMemoryConcurrent(t *testing.T) {
	s := NewMemory()
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				s.Set("npm", "acema", provider.StatusAvailable, "")
				s.Get("npm", "acema")
				s.Delete("npm", "acema")
			}
		}()
	}
	wg.Wait()
}
