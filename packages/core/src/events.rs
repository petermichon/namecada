//! Stream events shared by the NDJSON dev server and the Tauri bridge.
//!
//! Both transports send the same three events in order: `manifest`, one
//! `result` per source, then `done`. Keeping the shapes here stops the two
//! clients from drifting apart.

use std::collections::HashSet;

use serde::Serialize;

use crate::engine::ProviderResult;
use crate::manifest::Manifest;

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ManifestEvent<'a> {
    #[serde(rename = "type")]
    pub kind: &'static str,
    pub label: &'a str,
    pub sources: Vec<SourceEvent<'a>>,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct SourceEvent<'a> {
    pub id: &'a str,
    pub platform_name: &'a str,
    pub url: String,
    pub paid: bool,
    pub checked: bool,
}

#[derive(Debug, Clone, Serialize)]
pub struct ResultEvent<'a> {
    #[serde(rename = "type")]
    pub kind: &'static str,
    #[serde(flatten)]
    pub result: &'a ProviderResult,
}

#[derive(Debug, Clone, Serialize)]
pub struct DoneEvent {
    #[serde(rename = "type")]
    pub kind: &'static str,
}

impl<'a> ManifestEvent<'a> {
    pub fn new(label: &'a str, sources: Vec<SourceEvent<'a>>) -> Self {
        Self {
            kind: "manifest",
            label,
            sources,
        }
    }
}

impl<'a> SourceEvent<'a> {
    pub fn new(
        id: &'a str,
        platform_name: &'a str,
        url_template: &str,
        paid: bool,
        checked: bool,
        label: &str,
    ) -> Self {
        Self {
            id,
            platform_name,
            url: url_template.replace("%s", label),
            paid,
            checked,
        }
    }
}

impl<'a> ResultEvent<'a> {
    pub fn new(result: &'a ProviderResult) -> Self {
        Self {
            kind: "result",
            result,
        }
    }
}

impl Default for DoneEvent {
    fn default() -> Self {
        Self { kind: "done" }
    }
}

/// source_events lists the whole registry, marking the ids that have a
/// provider. `checked` is derived, never stored, so it cannot drift.
pub fn source_events<'a>(
    manifest: &'a Manifest,
    checked: &HashSet<String>,
    label: &'a str,
) -> Vec<SourceEvent<'a>> {
    manifest
        .sources
        .iter()
        .map(|source| {
            SourceEvent::new(
                &source.id,
                &source.platform_name,
                &source.url_template,
                source.paid,
                checked.contains(&source.id),
                label,
            )
        })
        .collect()
}
