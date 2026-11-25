use namecada_core::events::{self, DoneEvent, ManifestEvent, ResultEvent};
use namecada_core::{Engine, Manifest, MemoryCache};
use serde::Serialize;
use serde_json::Value;
use std::collections::HashSet;
use std::path::PathBuf;
use std::sync::Arc;
use std::time::Duration;
use tauri::ipc::Channel;
use tauri::{Manager, State};

/// SCAN_BUDGET lets the scan outlive any single request bound; the desktop
/// keeps the channel open until it resolves or this safety stop trips.
const SCAN_BUDGET: Duration = Duration::from_secs(24 * 60 * 60);

/// AppState is built once at startup: the registry, which ids have a provider,
/// a shared engine with its cache, and where that cache is persisted.
struct AppState {
    manifest: Manifest,
    checked: HashSet<String>,
    engine: Arc<Engine>,
    cache: Arc<MemoryCache>,
    cache_path: PathBuf,
}

/// check streams the same three events as the NDJSON server: the manifest, one
/// result per source, then done. Running the scan off the IPC thread keeps the
/// window responsive while slow sources resolve.
///
/// The desktop scans one label on one device, so it streams the whole scan over
/// the one channel instead of paging it: there is nothing to resume, and the
/// page never has to poll. `scan`/`cursor` stay in the signature so the shared
/// client can call the same shape as the HTTP transport; they are unused here.
#[tauri::command]
async fn check(
    label: String,
    scan: Option<String>,
    cursor: Option<usize>,
    on_event: Channel<Value>,
    state: State<'_, AppState>,
) -> Result<(), String> {
    // The shared client sends these for the HTTP transport; this command
    // streams the whole scan, so there is no page to resume.
    let _ = (scan, cursor);
    let label = label.trim().to_string();
    if label.is_empty() {
        return Err("invalid label".to_string());
    }

    let manifest = state.manifest.clone();
    let checked = state.checked.clone();
    let engine = Arc::clone(&state.engine);
    let cache = Arc::clone(&state.cache);
    let cache_path = state.cache_path.clone();

    tauri::async_runtime::spawn_blocking(move || {
        let sources = events::source_events(&manifest, &checked, &label);
        if send(&on_event, &ManifestEvent::new(&label, sources)).is_err() {
            return;
        }
        for result in engine.scan_with_deadline(&label, SCAN_BUDGET) {
            if send(&on_event, &ResultEvent::new(&result)).is_err() {
                return;
            }
        }
        // Persist before the done event so a relaunch starts with these
        // verdicts; a failed write leaves the in-memory cache intact.
        let _ = cache.save(&cache_path);
        let _ = send(&on_event, &DoneEvent::default());
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
            });
            Ok(())
        })
        .invoke_handler(tauri::generate_handler![check])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
