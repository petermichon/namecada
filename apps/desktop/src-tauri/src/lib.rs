use namecada_core::events::{self, DoneEvent, ManifestEvent, PendingEvent, ResultEvent};
use namecada_core::{Engine, Jobs, Manifest, MemoryCache};
use serde::Serialize;
use serde_json::Value;
use std::collections::HashSet;
use std::path::PathBuf;
use std::sync::Arc;
use tauri::ipc::Channel;
use tauri::{Manager, State};

/// AppState is built once at startup: the registry, which ids have a provider,
/// a shared engine with its cache, and where that cache is persisted.
struct AppState {
    manifest: Manifest,
    checked: HashSet<String>,
    engine: Arc<Engine>,
    cache: Arc<MemoryCache>,
    cache_path: PathBuf,
    /// jobs survive the IPC call: the scan keeps running while the frontend
    /// polls for the next page.
    jobs: Arc<Jobs>,
}

/// check streams the same three events as the NDJSON server: the manifest, one
/// result per source, then done. Running the scan off the IPC thread keeps the
/// window responsive while slow sources resolve.
#[tauri::command]
async fn check(
    label: String,
    scan: Option<String>,
    cursor: Option<usize>,
    on_event: Channel<Value>,
    state: State<'_, AppState>,
) -> Result<(), String> {
    let label = label.trim().to_string();
    if label.is_empty() {
        return Err("invalid label".to_string());
    }
    let cursor = cursor.unwrap_or(0);

    // Reuse a running job for this label, or start one. The scan is not bound
    // to this call: it keeps resolving after the IPC message returns.
    state.jobs.reap();
    let job = scan
        .as_deref()
        .and_then(|id| state.jobs.get(id))
        .filter(|job| job.label() == label)
        .unwrap_or_else(|| state.jobs.start(Arc::clone(&state.engine), &label));

    let manifest = state.manifest.clone();
    let checked = state.checked.clone();
    let cache = Arc::clone(&state.cache);
    let cache_path = state.cache_path.clone();

    tauri::async_runtime::spawn_blocking(move || {
        let sources = events::source_events(&manifest, &checked, &label);
        if send(&on_event, &ManifestEvent::new(&label, sources)).is_err() {
            return;
        }
        let results = job.results_from(cursor);
        let next = cursor + results.len();
        for result in &results {
            if send(&on_event, &ResultEvent::new(result)).is_err() {
                return;
            }
        }
        if job.is_done() {
            // Persist before the done event so a relaunch starts with these
            // verdicts; a failed write leaves the in-memory cache intact.
            let _ = cache.save(&cache_path);
            let _ = send(&on_event, &DoneEvent::default());
        } else {
            let _ = send(&on_event, &PendingEvent::new(job.id(), next));
        }
    })
    .await
    .map_err(|err| err.to_string())
}

fn send<T: Serialize>(channel: &Channel<Value>, event: &T) -> Result<(), ()> {
    let value = serde_json::to_value(event).map_err(|_| ())?;
    channel.send(value).map_err(|_| ())
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    let manifest = Manifest::embedded().expect("embedded registry");
    manifest.validate().expect("embedded registry is valid");
    let providers = namecada_core::live_providers(&manifest).expect("live providers");
    let checked: HashSet<String> = providers
        .iter()
        .map(|provider| provider.id().to_string())
        .collect();

    tauri::Builder::default()
        .setup(move |app| {
            let cache_path = app.path().app_data_dir()?.join("cache.json");
            let cache = Arc::new(MemoryCache::load(&cache_path));
            // Drop verdicts for sources that no longer ship.
            cache.retain_providers(|id| checked.contains(id));
            let engine = Arc::new(Engine::new(providers).with_cache(Arc::clone(&cache)));
            app.manage(AppState {
                manifest,
                checked,
                engine,
                cache,
                cache_path,
                jobs: Arc::new(Jobs::new()),
            });
            Ok(())
        })
        .invoke_handler(tauri::generate_handler![check])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
