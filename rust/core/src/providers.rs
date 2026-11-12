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

/// RegistryStats classifies the inventory against the provider tables. Nothing
/// is stored in `sources.json`, so the numbers cannot drift from the code.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct RegistryStats {
    pub sources: usize,
    pub checked: usize,
    pub keyed: usize,
    pub deferred: usize,
}

/// stats derives the inventory split: checked sources have a provider, keyed
/// ones need credentials, and the rest are deferred.
pub fn stats(manifest: &Manifest) -> RegistryStats {
    let checked: std::collections::HashSet<&str> = crate::status::status_specs()
        .iter()
        .map(|spec| spec.id)
        .chain(crate::custom::custom_specs().iter().map(|spec| spec.id))
        .chain(crate::dns::dns_hosts().iter().map(|(id, _)| *id))
        .collect();
    let mut stats = RegistryStats {
        sources: manifest.sources.len(),
        checked: 0,
        keyed: 0,
        deferred: 0,
    };
    for source in &manifest.sources {
        if checked.contains(source.id.as_str()) {
            stats.checked += 1;
        } else if source.keyed {
            stats.keyed += 1;
        } else {
            stats.deferred += 1;
        }
    }
    stats
}

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

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn stats_partition_the_inventory() {
        let manifest = Manifest::embedded().expect("manifest");
        let stats = stats(&manifest);
        assert_eq!(stats.sources, manifest.sources.len());
        assert_eq!(stats.sources, stats.checked + stats.keyed + stats.deferred);
        assert!(stats.checked > 100, "checked = {}", stats.checked);
        assert!(stats.keyed > 0, "keyed = {}", stats.keyed);
    }
}
