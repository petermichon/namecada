use namecada_core::events::{self, DoneEvent, ManifestEvent, ResultEvent};
use namecada_core::{Engine, Manifest, MemoryCache};
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
}

/// check streams the same three events as the NDJSON server: the manifest, one
/// result per source, then done. Running the scan off the IPC thread keeps the
/// window responsive while slow sources resolve.
#[tauri::command]
async fn check(
    label: String,
    on_event: Channel<Value>,
    state: State<'_, AppState>,
) -> Result<(), String> {
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
        for result in engine.scan(&label) {
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
