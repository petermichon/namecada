//! The source inventory produced by `server/cmd/domainbrowser export`.

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
    pub sources: Vec<Source>,
}

impl Manifest {
    pub fn load(path: impl AsRef<Path>) -> Result<Self, Error> {
        let data = std::fs::read(path)?;
        Ok(serde_json::from_slice(&data)?)
    }

    pub fn source(&self, id: &str) -> Option<&Source> {
        self.sources.iter().find(|source| source.id == id)
    }

    /// enabled returns the sources a zero-config scan would run.
    pub fn enabled(&self) -> impl Iterator<Item = &Source> {
        self.sources.iter().filter(|source| !source.keyed)
    }
}
