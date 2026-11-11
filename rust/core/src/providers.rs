//! Builds the live provider set from the compiled registry.
//!
//! Every transport (the dev server, the Tauri app) needs the same wiring, so it
//! lives here instead of in a binary.

use std::sync::Arc;

use crate::custom::build_custom_provider;
use crate::dns::{build_dns_provider, SystemResolver};
use crate::http::HttpTransport;
use crate::manifest::Manifest;
use crate::provider::{Error, Provider, Transport};
use crate::status::build_status_provider;

/// live returns every keyless provider the registry and the adapters know
/// about, in a stable order (status, custom, then DNS).
pub fn live(manifest: &Manifest) -> Result<Vec<Arc<dyn Provider>>, Error> {
    let transport: Arc<dyn Transport> = Arc::new(HttpTransport::new());
    let mut providers: Vec<Arc<dyn Provider>> = Vec::new();

    for spec in crate::status::status_specs() {
        providers.push(Arc::new(build_status_provider(
            manifest,
            spec.id,
            Arc::clone(&transport),
        )?));
    }
    for spec in crate::custom::custom_specs() {
        providers.push(Arc::new(build_custom_provider(
            manifest,
            spec.id,
            Arc::clone(&transport),
        )?));
    }
    let resolver: Arc<dyn crate::dns::Resolver> = Arc::new(SystemResolver);
    for (id, _) in crate::dns::dns_hosts() {
        providers.push(Arc::new(build_dns_provider(
            manifest,
            id,
            Arc::clone(&resolver),
        )?));
    }
    Ok(providers)
}
