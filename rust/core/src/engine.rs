//! Bounded-concurrency scan engine. Results stream as each provider resolves.

use std::collections::HashMap;
use std::sync::atomic::{AtomicUsize, Ordering};
use std::sync::{mpsc, Arc};
use std::thread;

use serde::Serialize;

use crate::cache::MemoryCache;
use crate::provider::{Outcome, Provider, Status};

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
    cache: Option<Arc<MemoryCache>>,
}

impl Engine {
    pub fn new(providers: Vec<Arc<dyn Provider>>) -> Self {
        let concurrency = providers.len().max(1);
        Self {
            providers,
            concurrency,
            cache: None,
        }
    }

    pub fn with_concurrency(mut self, concurrency: usize) -> Self {
        self.concurrency = concurrency.max(1);
        self
    }

    pub fn with_cache(mut self, cache: Arc<MemoryCache>) -> Self {
        self.cache = Some(cache);
        self
    }

    /// scan returns a receiver that emits each result as it resolves and closes
    /// when every worker has finished.
    pub fn scan(&self, label: &str) -> mpsc::Receiver<ProviderResult> {
        let (tx, rx) = mpsc::channel();
        if self.providers.is_empty() {
            return rx;
        }

        let providers = self.providers.clone();
        let label = label.to_string();
        let cache = self.cache.clone();
        let next = Arc::new(AtomicUsize::new(0));
        let workers = self.concurrency.min(providers.len()).max(1);

        for _ in 0..workers {
            let providers = providers.clone();
            let tx = tx.clone();
            let label = label.clone();
            let cache = cache.clone();
            let next = Arc::clone(&next);
            thread::spawn(move || loop {
                let index = next.fetch_add(1, Ordering::Relaxed);
                if index >= providers.len() {
                    break;
                }
                let result = check_one(providers[index].as_ref(), &label, cache.as_deref());
                if tx.send(result).is_err() {
                    break;
                }
            });
        }
        drop(tx);
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

fn check_one(provider: &dyn Provider, raw: &str, cache: Option<&MemoryCache>) -> ProviderResult {
    let label = provider.normalize_label(raw);
    let mut result = ProviderResult {
        id: provider.id().to_string(),
        platform: provider.platform().to_string(),
        platform_name: provider.platform_name(),
        name: provider.name(&label),
        url: provider.url(&label),
        status: Status::Unknown,
        detail: String::new(),
        paid: provider.paid(),
    };

    let length = label.chars().count();
    if provider.min_length() > 0 && length < provider.min_length() {
        result.status = Status::Taken;
        return result;
    }
    if let Some(max) = provider.max_length() {
        if length > max {
            result.status = Status::Taken;
            return result;
        }
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
    use crate::cache::MemoryCache;
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
            assert_eq!(results[0].status, Status::Taken, "{label}");
            assert_eq!(calls.load(Ordering::SeqCst), 0, "{label}");
        }

        let results = engine.scan_all("abcd");
        assert_eq!(results[0].status, Status::Available);
        assert_eq!(calls.load(Ordering::SeqCst), 1);
    }
}
