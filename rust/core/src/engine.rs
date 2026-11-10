//! Bounded-concurrency scan engine. Results stream as each provider resolves.

use std::collections::HashMap;
use std::sync::atomic::{AtomicUsize, Ordering};
use std::sync::{mpsc, Arc};
use std::thread;

use serde::Serialize;

use crate::provider::{Outcome, Provider, Status};

/// DefaultConcurrency bounds how many provider checks run at once.
pub const DEFAULT_CONCURRENCY: usize = 8;

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
}

impl Engine {
    pub fn new(providers: Vec<Arc<dyn Provider>>) -> Self {
        Self {
            providers,
            concurrency: DEFAULT_CONCURRENCY,
        }
    }

    pub fn with_concurrency(mut self, concurrency: usize) -> Self {
        self.concurrency = concurrency.max(1);
        self
    }

    pub fn providers(&self) -> &[Arc<dyn Provider>] {
        &self.providers
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
        let next = Arc::new(AtomicUsize::new(0));
        let workers = self.concurrency.min(providers.len()).max(1);

        for _ in 0..workers {
            let providers = providers.clone();
            let tx = tx.clone();
            let label = label.clone();
            let next = Arc::clone(&next);
            thread::spawn(move || loop {
                let index = next.fetch_add(1, Ordering::Relaxed);
                if index >= providers.len() {
                    break;
                }
                let result = check_one(providers[index].as_ref(), &label);
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

fn check_one(provider: &dyn Provider, label: &str) -> ProviderResult {
    let mut result = ProviderResult {
        id: provider.id().to_string(),
        platform: provider.platform().to_string(),
        platform_name: provider.platform_name(),
        name: provider.name(label),
        url: provider.url(label),
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

    match provider.check(label) {
        Ok(outcome) => apply(&mut result, outcome),
        Err(_) => {
            result.status = Status::Unknown;
            result.detail = "unreachable".to_string();
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
