//! The source inventory in `fixtures/sources.json`.

use std::collections::HashSet;
use std::path::Path;

use serde::Deserialize;

use crate::provider::Error;

/// Source is one row of `fixtures/sources.json`.
#[derive(Debug, Clone, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Source {
    pub id: String,
    pub platform: String,
    pub platform_name: String,
    /// Name for the manifest label (e.g. `acema.com`).
    pub name: String,
    /// Name format with `%s` for the label.
    pub name_template: String,
    /// Public page for the manifest label.
    pub url: String,
    /// Public page format with `%s` for the label.
    pub url_template: String,
    pub paid: bool,
    pub min_length: usize,
    pub max_length: usize,
    pub keyed: bool,
    /// Facet tags (e.g. `domains`, `social`) used by the report filters.
    #[serde(default)]
    pub tags: Vec<String>,
}

/// Manifest is the complete source set.
#[derive(Debug, Clone, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Manifest {
    pub generated_at: String,
    pub label: String,
    pub source_count: usize,
    pub enabled_count: usize,
    pub keyed_count: usize,
    pub deferred_count: usize,
    pub sources: Vec<Source>,
}

/// EMBEDDED_SOURCES is the registry compiled into the binary: the packaged app
/// has no `fixtures/` directory to load at runtime.
const EMBEDDED_SOURCES: &str = include_str!("../../../fixtures/sources.json");

impl Manifest {
    pub fn load(path: impl AsRef<Path>) -> Result<Self, Error> {
        let data = std::fs::read(path)?;
        Ok(serde_json::from_slice(&data)?)
    }

    /// embedded returns the registry compiled into the binary.
    pub fn embedded() -> Result<Self, Error> {
        Ok(serde_json::from_str(EMBEDDED_SOURCES)?)
    }

    /// validate catches registry mistakes at startup rather than as a blank
    /// manifest row or a broken URL.
    pub fn validate(&self) -> Result<(), Error> {
        let mut ids = HashSet::new();
        for source in &self.sources {
            if !ids.insert(source.id.as_str()) {
                return Err(Error::Invalid(format!(
                    "duplicate source id {:?}",
                    source.id
                )));
            }
            if !source.url_template.contains("%s") {
                return Err(Error::Invalid(format!(
                    "source {:?} urlTemplate lacks %s",
                    source.id
                )));
            }
            if !source.name_template.contains("%s") {
                return Err(Error::Invalid(format!(
                    "source {:?} nameTemplate lacks %s",
                    source.id
                )));
            }
        }
        if self.source_count != self.sources.len() {
            return Err(Error::Invalid(format!(
                "sourceCount {} does not match {} sources",
                self.source_count,
                self.sources.len()
            )));
        }
        if self.source_count != self.enabled_count + self.keyed_count + self.deferred_count {
            return Err(Error::Invalid(
                "sourceCount does not equal enabledCount + keyedCount + deferredCount".to_string(),
            ));
        }
        Ok(())
    }

    pub fn source(&self, id: &str) -> Option<&Source> {
        self.sources.iter().find(|source| source.id == id)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn embedded_registry_is_valid() {
        let manifest = Manifest::embedded().expect("embedded registry parses");
        manifest.validate().expect("embedded registry is valid");
        assert!(manifest.sources.len() >= 100);
        assert!(manifest
            .sources
            .iter()
            .all(|source| !source.tags.is_empty()));
    }
}
