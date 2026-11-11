use std::path::Path;
use std::sync::Arc;

use namecheck_core::{
    build_custom_provider, build_status_provider, custom_specs, status_specs, Fixture, Manifest,
    Provider, ReplayTransport,
};

/// live builds the complete keyless registry against real HTTP and DNS.
pub fn live(manifest: &Manifest) -> Result<Vec<Arc<dyn Provider>>, Box<dyn std::error::Error>> {
    Ok(namecheck_core::live_providers(manifest)?)
}

/// replay builds the HTTP-backed providers from a label fixture directory.
pub fn replay(
    manifest: &Manifest,
    label_dir: &Path,
) -> Result<Vec<Arc<dyn Provider>>, Box<dyn std::error::Error>> {
    let mut providers: Vec<Arc<dyn Provider>> = Vec::new();

    for spec in status_specs() {
        let path = label_dir.join(format!("{}.json", spec.id));
        if !path.exists() {
            continue;
        }
        let fixture = Fixture::load(&path)?;
        let transport = Arc::new(ReplayTransport::new(&fixture));
        providers.push(Arc::new(build_status_provider(
            manifest, spec.id, transport,
        )?));
    }
    for spec in custom_specs() {
        let path = label_dir.join(format!("{}.json", spec.id));
        if !path.exists() {
            continue;
        }
        let fixture = Fixture::load(&path)?;
        let transport = Arc::new(ReplayTransport::new(&fixture));
        providers.push(Arc::new(build_custom_provider(
            manifest, spec.id, transport,
        )?));
    }
    Ok(providers)
}
