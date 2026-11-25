//! Resumable scan jobs: one scan that outlives any single response.
//!
//! A request must answer inside a fixed budget, but the scan does not have to
//! finish inside it. A job runs the whole source list in the background and
//! keeps every verdict it resolves; each request returns the slice the caller
//! has not seen yet, plus the job id so it can come back. The user waits for a
//! page, not for the catalogue.

use std::collections::HashMap;
use std::sync::atomic::{AtomicBool, AtomicUsize, Ordering};
use std::sync::{Arc, Mutex};
use std::time::{Duration, Instant};

use crate::engine::{Engine, ProviderResult};

/// JOB_BUDGET is a safety stop for a runaway job, not a product deadline. The
/// user is never blocked on it, so it can be far longer than any request.
const JOB_BUDGET: Duration = Duration::from_secs(24 * 60 * 60);

/// JOBS keep a finished job around for this long, so a slow client can still
/// collect the tail after the scan completed.
pub const JOB_TTL: Duration = Duration::from_secs(10 * 60);

static NEXT: AtomicUsize = AtomicUsize::new(1);

/// ScanJob is one label's scan, running off the request path.
pub struct ScanJob {
    id: String,
    label: String,
    results: Arc<Mutex<Vec<ProviderResult>>>,
    done: Arc<AtomicBool>,
    created: Instant,
}

impl ScanJob {
    /// start spawns the scan and returns immediately. The job keeps running if
    /// every client goes away; its results stay collectable until the TTL.
    pub fn start(engine: Arc<Engine>, label: String) -> Self {
        let id = format!(
            "{}-{}",
            std::process::id(),
            NEXT.fetch_add(1, Ordering::Relaxed)
        );
        let results: Arc<Mutex<Vec<ProviderResult>>> = Arc::new(Mutex::new(Vec::new()));
        let done = Arc::new(AtomicBool::new(false));
        let thread_results = Arc::clone(&results);
        let thread_done = Arc::clone(&done);
        let thread_label = label.clone();
        std::thread::spawn(move || {
            for result in engine.scan_with_deadline(&thread_label, JOB_BUDGET) {
                thread_results
                    .lock()
                    .expect("job mutex poisoned")
                    .push(result);
            }
            thread_done.store(true, Ordering::Relaxed);
        });
        Self {
            id,
            label,
            results,
            done,
            created: Instant::now(),
        }
    }

    pub fn id(&self) -> &str {
        &self.id
    }

    pub fn label(&self) -> &str {
        &self.label
    }

    pub fn len(&self) -> usize {
        self.results.lock().expect("job mutex poisoned").len()
    }

    pub fn is_empty(&self) -> bool {
        self.len() == 0
    }

    pub fn is_done(&self) -> bool {
        self.done.load(Ordering::Relaxed)
    }

    pub fn age(&self) -> Duration {
        self.created.elapsed()
    }

    /// results_from clones the verdicts the caller has not seen yet. Cloning
    /// only the slice keeps a poll proportional to what is new, not to the
    /// whole catalogue.
    pub fn results_from(&self, cursor: usize) -> Vec<ProviderResult> {
        let results = self.results.lock().expect("job mutex poisoned");
        if cursor >= results.len() {
            return Vec::new();
        }
        results[cursor..].to_vec()
    }
}

/// Jobs is the transport-side registry of running scans.
#[derive(Default)]
pub struct Jobs {
    jobs: Mutex<HashMap<String, Arc<ScanJob>>>,
}

impl Jobs {
    pub fn new() -> Self {
        Self::default()
    }

    /// start registers a fresh job for a label.
    pub fn start(&self, engine: Arc<Engine>, label: &str) -> Arc<ScanJob> {
        let job = Arc::new(ScanJob::start(engine, label.to_string()));
        self.jobs
            .lock()
            .expect("jobs mutex poisoned")
            .insert(job.id().to_string(), Arc::clone(&job));
        job
    }

    /// get returns a running (or recently finished) job.
    pub fn get(&self, id: &str) -> Option<Arc<ScanJob>> {
        self.jobs
            .lock()
            .expect("jobs mutex poisoned")
            .get(id)
            .cloned()
    }

    /// reap drops jobs older than the TTL so a long-lived server does not keep
    /// every finished scan in memory.
    pub fn reap(&self) {
        let mut jobs = self.jobs.lock().expect("jobs mutex poisoned");
        jobs.retain(|_, job| job.age() < JOB_TTL);
    }
}
