//! Bounded-concurrency scan engine and its in-memory verdict cache.

use std::collections::HashMap;
use std::sync::atomic::{AtomicUsize, Ordering};
use std::sync::{mpsc, Arc, Mutex};
use std::thread;
use std::time::{Duration, Instant};

use serde::Serialize;

use crate::provider::{Outcome, Provider, Status};

/// Fresh is how long a cached verdict stays current.
const FRESH: Duration = Duration::from_secs(60 * 60);

/// DEFAULT_CONCURRENCY caps the checks in flight. Every transport (the dev
/// server, the Tauri shell and Android) builds its engine here, so one scan
/// opens this many sockets instead of one per provider; `with_concurrency`
/// overrides it for experiments.
pub const DEFAULT_CONCURRENCY: usize = 16;

/// DEFAULT_DEADLINE bounds a whole scan. Whatever has not resolved by then is
/// reported unknown — which is never cached, so a later scan retries it — and
/// one stalled host cannot stretch the report to a minute.
pub const DEFAULT_DEADLINE: Duration = Duration::from_secs(30);

/// DeadlineDetail explains an unknown produced by the scan deadline.
pub const DEADLINE_DETAIL: &str = "no verdict before the deadline";

/// Entry is a cached verdict and when it was recorded.
#[derive(Debug, Clone)]
pub struct Entry {
    pub status: Status,
    pub detail: String,
    pub url: String,
    pub at: Instant,
}

/// MemoryCache is a concurrency-safe in-memory verdict cache.
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

impl MemoryCache {
    pub fn get(&self, provider: &str, label: &str) -> Option<Entry> {
        let key = format!("{provider}:{label}");
        let mut entries = self.entries.lock().expect("cache mutex poisoned");
        let entry = entries.get(&key)?.clone();
        if entry.at.elapsed() >= self.ttl {
            entries.remove(&key);
            return None;
        }
        Some(entry)
    }

    pub fn set(&self, provider: &str, label: &str, status: Status, detail: &str, url: &str) {
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

/// ProviderResult is the public shape streamed to the UI and CLI.
#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
#[serde(rename_all = "camelCase")]
pub struct ProviderResult {
    pub id: String,
    pub platform: String,
    pub platform_name: String,
    pub name: String,
    pub url: String,
    pub status: Status,
    #[serde(skip_serializing_if = "String::is_empty")]
    pub detail: String,
    pub paid: bool,
}

/// Engine scans a label across a fixed, ordered provider set.
pub struct Engine {
    providers: Vec<Arc<dyn Provider>>,
    concurrency: usize,
    deadline: Duration,
    cache: Option<Arc<MemoryCache>>,
}

impl Engine {
    pub fn new(providers: Vec<Arc<dyn Provider>>) -> Self {
        let concurrency = DEFAULT_CONCURRENCY.min(providers.len()).max(1);
        Self {
            providers,
            concurrency,
            deadline: DEFAULT_DEADLINE,
            cache: None,
        }
    }

    pub fn with_concurrency(mut self, concurrency: usize) -> Self {
        self.concurrency = concurrency.max(1);
        self
    }

    pub fn with_deadline(mut self, deadline: Duration) -> Self {
        self.deadline = deadline;
        self
    }

    pub fn with_cache(mut self, cache: Arc<MemoryCache>) -> Self {
        self.cache = Some(cache);
        self
    }

    /// scan returns a receiver that emits each result as it resolves and closes
    /// at the deadline, with any unresolved source reported unknown.
    pub fn scan(&self, label: &str) -> mpsc::Receiver<ProviderResult> {
        let (tx, rx) = mpsc::channel();
        if self.providers.is_empty() {
            return rx;
        }

        let expires = Instant::now() + self.deadline;
        let providers = self.providers.clone();
        let label = label.to_string();
        let cache = self.cache.clone();
        let next = Arc::new(AtomicUsize::new(0));
        let total = providers.len();
        let workers = self.concurrency.min(total).max(1);

        // Workers report to an internal channel so a supervisor can close the
        // caller's stream at the deadline and pad the sources still pending.
        let (inner_tx, inner_rx) = mpsc::channel();
        for _ in 0..workers {
            let providers = providers.clone();
            let inner_tx = inner_tx.clone();
            let label = label.clone();
            let cache = cache.clone();
            let next = Arc::clone(&next);
            thread::spawn(move || loop {
                if Instant::now() >= expires {
                    break;
                }
                let index = next.fetch_add(1, Ordering::Relaxed);
                if index >= providers.len() {
                    break;
                }
                let result = check_one(providers[index].as_ref(), &label, cache.as_deref());
                if inner_tx.send((index, result)).is_err() {
                    break;
                }
            });
        }
        drop(inner_tx);

        thread::spawn(move || {
            let mut seen = vec![false; total];
            let mut left = total;
            while left > 0 {
                let wait = expires.saturating_duration_since(Instant::now());
                match inner_rx.recv_timeout(wait) {
                    Ok((index, result)) => {
                        if !seen[index] {
                            seen[index] = true;
                            left -= 1;
                        }
                        if tx.send(result).is_err() {
                            return;
                        }
                    }
                    Err(_) => break,
                }
            }
            for (index, provider) in providers.iter().enumerate() {
                if seen[index] {
                    continue;
                }
                if tx.send(pending_result(provider.as_ref(), &label)).is_err() {
                    return;
                }
            }
        });
        rx
    }

    /// scan_all collects scan and returns results in provider order.
    pub fn scan_all(&self, label: &str) -> Vec<ProviderResult> {
        let mut by_id = HashMap::new();
        for result in self.scan(label) {
            by_id.insert(result.id.clone(), result);
        }
        self.providers
            .iter()
            .filter_map(|provider| by_id.remove(provider.id()))
            .collect()
    }
}

/// pending_result describes a source that has not answered by the deadline.
fn pending_result(provider: &dyn Provider, raw: &str) -> ProviderResult {
    base_result(provider, raw, DEADLINE_DETAIL)
}

fn base_result(provider: &dyn Provider, raw: &str, detail: &str) -> ProviderResult {
    let label = provider.normalize_label(raw);
    ProviderResult {
        id: provider.id().to_string(),
        platform: provider.platform().to_string(),
        platform_name: provider.platform_name(),
        name: provider.name(&label),
        url: provider.url(&label),
        status: Status::Unknown,
        detail: detail.to_string(),
        paid: provider.paid(),
    }
}

fn check_one(provider: &dyn Provider, raw: &str, cache: Option<&MemoryCache>) -> ProviderResult {
    let label = provider.normalize_label(raw);
    let mut result = base_result(provider, raw, "");

    let length = label.chars().count();
    if provider.min_length() > 0 && length < provider.min_length() {
        result.status = Status::Invalid;
        result.detail = "outside the length limit".to_string();
        return result;
    }
    if let Some(max) = provider.max_length() {
        if length > max {
            result.status = Status::Invalid;
            result.detail = "outside the length limit".to_string();
            return result;
        }
    }
    if !provider.chars().allows(&label) {
        result.status = Status::Invalid;
        result.detail = "invalid characters".to_string();
        return result;
    }

    if let Some(cache) = cache {
        if let Some(entry) = cache.get(provider.id(), &label) {
            result.status = entry.status;
            result.detail = entry.detail;
            if !entry.url.is_empty() {
                result.url = entry.url;
            }
            return result;
        }
    }

    match provider.check(&label) {
        Ok(outcome) => apply(&mut result, outcome),
        Err(_) => {
            result.status = Status::Unknown;
            result.detail = "unreachable".to_string();
        }
    }

    if let Some(cache) = cache {
        if result.status != Status::Unknown {
            cache.set(
                provider.id(),
                &label,
                result.status,
                &result.detail,
                &result.url,
            );
        }
    }
    result
}

fn apply(result: &mut ProviderResult, outcome: Outcome) {
    result.status = outcome.status;
    result.detail = outcome.detail;
    if let Some(url) = outcome.url {
        if !url.is_empty() {
            result.url = url;
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::provider::{Error, Outcome};

    struct CountingProvider {
        calls: Arc<AtomicUsize>,
    }

    impl Provider for CountingProvider {
        fn id(&self) -> &str {
            "stub"
        }
        fn platform(&self) -> &str {
            "stub"
        }
        fn name(&self, label: &str) -> String {
            label.to_string()
        }
        fn url(&self, label: &str) -> String {
            format!("https://example.test/{label}")
        }
        fn paid(&self) -> bool {
            false
        }
        fn check(&self, _: &str) -> Result<Outcome, Error> {
            self.calls.fetch_add(1, Ordering::SeqCst);
            Ok(Outcome::new(Status::Available))
        }
    }

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

    #[test]
    fn cache_avoids_a_second_check() {
        let calls = Arc::new(AtomicUsize::new(0));
        let provider = Arc::new(CountingProvider {
            calls: Arc::clone(&calls),
        });
        let cache = Arc::new(MemoryCache::new());
        let engine = Engine::new(vec![provider]).with_cache(cache);

        let first = engine.scan_all("acema");
        let second = engine.scan_all("acema");

        assert_eq!(calls.load(Ordering::SeqCst), 1);
        assert_eq!(first[0].status, Status::Available);
        assert_eq!(second[0].status, Status::Available);
    }

    /// A label outside the source's own grammar cannot be claimed, so the
    /// engine reports it taken without spending a request.
    struct BoundedProvider {
        calls: Arc<AtomicUsize>,
    }

    impl Provider for BoundedProvider {
        fn id(&self) -> &str {
            "bounded"
        }
        fn platform(&self) -> &str {
            "bounded"
        }
        fn name(&self, label: &str) -> String {
            label.to_string()
        }
        fn url(&self, label: &str) -> String {
            format!("https://example.test/{label}")
        }
        fn paid(&self) -> bool {
            false
        }
        fn min_length(&self) -> usize {
            3
        }
        fn max_length(&self) -> Option<usize> {
            Some(15)
        }
        fn check(&self, _: &str) -> Result<Outcome, Error> {
            self.calls.fetch_add(1, Ordering::SeqCst);
            Ok(Outcome::new(Status::Available))
        }
    }

    #[test]
    fn length_limits_skip_the_request() {
        let calls = Arc::new(AtomicUsize::new(0));
        let engine = Engine::new(vec![Arc::new(BoundedProvider {
            calls: Arc::clone(&calls),
        })]);

        for label in ["ab", "abcdefghijklmnop"] {
            let results = engine.scan_all(label);
            assert_eq!(results[0].status, Status::Invalid, "{label}");
            assert_eq!(calls.load(Ordering::SeqCst), 0, "{label}");
        }

        let results = engine.scan_all("abcd");
        assert_eq!(results[0].status, Status::Available);
        assert_eq!(calls.load(Ordering::SeqCst), 1);
    }

    #[test]
    fn out_of_alphabet_labels_skip_the_request() {
        let calls = Arc::new(AtomicUsize::new(0));
        let engine = Engine::new(vec![Arc::new(BoundedProvider {
            calls: Arc::clone(&calls),
        })]);

        // The stub is LDH-only, so a dotted name cannot exist there.
        let results = engine.scan_all("example.com");
        assert_eq!(results[0].status, Status::Invalid);
        assert_eq!(results[0].detail, "invalid characters");
        assert_eq!(calls.load(Ordering::SeqCst), 0);
    }

    #[test]
    fn default_concurrency_is_capped() {
        let calls = Arc::new(AtomicUsize::new(0));
        let providers: Vec<Arc<dyn Provider>> = (0..DEFAULT_CONCURRENCY * 4)
            .map(|_| {
                Arc::new(BoundedProvider {
                    calls: Arc::clone(&calls),
                }) as Arc<dyn Provider>
            })
            .collect();
        assert_eq!(Engine::new(providers).concurrency, DEFAULT_CONCURRENCY);
        assert_eq!(Engine::new(Vec::new()).concurrency, 1);

        let one = Engine::new(vec![Arc::new(BoundedProvider { calls })]);
        assert_eq!(one.concurrency, 1);
    }

    struct SlowProvider {
        delay: Duration,
    }

    impl Provider for SlowProvider {
        fn id(&self) -> &str {
            "slow"
        }
        fn platform(&self) -> &str {
            "slow"
        }
        fn name(&self, label: &str) -> String {
            label.to_string()
        }
        fn url(&self, label: &str) -> String {
            format!("https://slow.test/{label}")
        }
        fn paid(&self) -> bool {
            false
        }
        fn check(&self, _: &str) -> Result<Outcome, Error> {
            thread::sleep(self.delay);
            Ok(Outcome::new(Status::Taken))
        }
    }

    #[test]
    fn deadline_reports_pending_sources_as_unknown() {
        let engine = Engine::new(vec![
            Arc::new(BoundedProvider {
                calls: Arc::new(AtomicUsize::new(0)),
            }),
            Arc::new(SlowProvider {
                delay: Duration::from_secs(2),
            }),
        ])
        .with_deadline(Duration::from_millis(100));

        let started = Instant::now();
        let results = engine.scan_all("abcd");
        assert!(started.elapsed() < Duration::from_secs(1));

        assert_eq!(results.len(), 2);
        assert_eq!(results[0].id, "bounded");
        assert_eq!(results[0].status, Status::Available);
        assert_eq!(results[1].id, "slow");
        assert_eq!(results[1].status, Status::Unknown);
        assert_eq!(results[1].detail, DEADLINE_DETAIL);
    }
}
