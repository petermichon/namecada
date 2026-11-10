//! In-memory verdict cache. Only definitive statuses are stored; unknown
//! responses always go back to the source so transient failures self-heal.

use std::collections::HashMap;
use std::sync::Mutex;
use std::time::{Duration, Instant};

use crate::provider::Status;

/// Fresh is how long a cached verdict stays current.
const FRESH: Duration = Duration::from_secs(60 * 60);

/// Entry is a cached verdict and when it was recorded.
#[derive(Debug, Clone)]
pub struct Entry {
    pub status: Status,
    pub detail: String,
    pub url: String,
    pub at: Instant,
}

/// Cache stores verdicts by provider:label.
pub trait Cache: Send + Sync {
    fn get(&self, provider: &str, label: &str) -> Option<Entry>;
    fn set(&self, provider: &str, label: &str, status: Status, detail: &str, url: &str);
}

/// MemoryCache is a concurrency-safe in-memory cache.
pub struct MemoryCache {
    ttl: Duration,
    entries: Mutex<HashMap<String, Entry>>,
}

impl MemoryCache {
    pub fn new() -> Self {
        Self::with_ttl(FRESH)
    }

    pub fn with_ttl(ttl: Duration) -> Self {
        Self {
            ttl,
            entries: Mutex::new(HashMap::new()),
        }
    }

    pub fn len(&self) -> usize {
        self.entries.lock().expect("cache mutex poisoned").len()
    }

    pub fn is_empty(&self) -> bool {
        self.len() == 0
    }
}

impl Default for MemoryCache {
    fn default() -> Self {
        Self::new()
    }
}

impl Cache for MemoryCache {
    fn get(&self, provider: &str, label: &str) -> Option<Entry> {
        let key = format!("{provider}:{label}");
        let mut entries = self.entries.lock().expect("cache mutex poisoned");
        let entry = entries.get(&key)?.clone();
        if entry.at.elapsed() >= self.ttl {
            entries.remove(&key);
            return None;
        }
        Some(entry)
    }

    fn set(&self, provider: &str, label: &str, status: Status, detail: &str, url: &str) {
        if status == Status::Unknown {
            return;
        }
        let key = format!("{provider}:{label}");
        self.entries.lock().expect("cache mutex poisoned").insert(
            key,
            Entry {
                status,
                detail: detail.to_string(),
                url: url.to_string(),
                at: Instant::now(),
            },
        );
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn stores_and_returns_definitive_verdicts() {
        let cache = MemoryCache::new();
        cache.set(
            "npm",
            "acema",
            Status::Available,
            "",
            "https://example.test",
        );
        let entry = cache.get("npm", "acema").expect("cached");
        assert_eq!(entry.status, Status::Available);
        assert_eq!(entry.url, "https://example.test");
        assert!(cache.get("npm", "other").is_none());
    }

    #[test]
    fn never_stores_unknown() {
        let cache = MemoryCache::new();
        cache.set("npm", "acema", Status::Unknown, "", "");
        assert!(cache.get("npm", "acema").is_none());
        assert!(cache.is_empty());
    }

    #[test]
    fn expires_stale_entries() {
        let cache = MemoryCache::with_ttl(Duration::ZERO);
        cache.set("npm", "acema", Status::Taken, "", "");
        assert!(cache.get("npm", "acema").is_none());
        assert!(cache.is_empty());
    }
}
