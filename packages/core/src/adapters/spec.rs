//! The custom-provider spec table: one row per source, mapping an id to its
//! check function.

use std::sync::Arc;

use crate::manifest::{Manifest, Source};
use crate::provider::{Error, Outcome, Provider, Transport};

use super::shared::render;

use super::api::*;
use super::domains::*;
use super::page::*;

pub type CheckFn = fn(&Source, &dyn Transport, &str) -> Result<Outcome, Error>;

/// CustomSpec ties a manifest id to its check function.
pub struct CustomSpec {
    pub id: &'static str,
    pub check: CheckFn,
}

/// custom_specs returns every provider this adapter currently covers.
pub fn custom_specs() -> &'static [CustomSpec] {
    CUSTOM_SPECS
}

pub(crate) static CUSTOM_SPECS: &[CustomSpec] = &[
    CustomSpec {
        id: "ens",
        check: check_ens,
    },
    CustomSpec {
        id: "sns",
        check: check_sns,
    },
    CustomSpec {
        id: "hns",
        check: check_hns,
    },
    CustomSpec {
        id: "namecoin",
        check: check_namecoin,
    },
    CustomSpec {
        id: "lens",
        check: check_lens,
    },
    CustomSpec {
        id: "anilist",
        check: check_anilist,
    },
    CustomSpec {
        id: "roblox",
        check: check_roblox,
    },
    CustomSpec {
        id: "codeberg",
        check: check_codeberg,
    },
    CustomSpec {
        id: "wikipedia",
        check: check_wikipedia,
    },
    CustomSpec {
        id: "devto",
        check: check_devto,
    },
    CustomSpec {
        id: "minehut",
        check: check_minehut,
    },
    CustomSpec {
        id: "bluesky",
        check: check_bluesky,
    },
    CustomSpec {
        id: "opencollective",
        check: check_opencollective,
    },
    CustomSpec {
        id: "instagram",
        check: check_instagram,
    },
    CustomSpec {
        id: "dribbble",
        check: check_dribbble,
    },
    CustomSpec {
        id: "hackernews",
        check: check_hackernews,
    },
    CustomSpec {
        id: "giphy",
        check: check_giphy,
    },
    CustomSpec {
        id: "minds",
        check: check_minds,
    },
    CustomSpec {
        id: "snap",
        check: check_snap,
    },
    CustomSpec {
        id: "bandcamp",
        check: check_bandcamp,
    },
    CustomSpec {
        id: "steam",
        check: check_steam,
    },
    CustomSpec {
        id: "wordpress",
        check: check_wordpress,
    },
    CustomSpec {
        id: "gitlab",
        check: check_gitlab,
    },
    CustomSpec {
        id: "peertube",
        check: check_peertube,
    },
    CustomSpec {
        id: "domain.com",
        check: check_domain,
    },
    CustomSpec {
        id: "domain.org",
        check: check_domain,
    },
    CustomSpec {
        id: "domain.dev",
        check: check_domain,
    },
    CustomSpec {
        id: "domain.io",
        check: check_domain,
    },
    CustomSpec {
        id: "domain.net",
        check: check_domain,
    },
    CustomSpec {
        id: "domain.ai",
        check: check_domain,
    },
    CustomSpec {
        id: "domain.app",
        check: check_domain,
    },
    CustomSpec {
        id: "tiktok",
        check: check_tiktok,
    },
    CustomSpec {
        id: "freelancer",
        check: check_freelancer,
    },
    CustomSpec {
        id: "threads",
        check: check_threads,
    },
    CustomSpec {
        id: "twitch",
        check: check_twitch,
    },
    CustomSpec {
        id: "duolingo",
        check: check_duolingo,
    },
    CustomSpec {
        id: "telegram",
        check: check_telegram,
    },
];

pub struct CustomProvider {
    source: Source,
    check: CheckFn,
    transport: Arc<dyn Transport>,
}

impl CustomProvider {
    pub fn new(source: Source, check: CheckFn, transport: Arc<dyn Transport>) -> Self {
        Self {
            source,
            check,
            transport,
        }
    }
}

impl Provider for CustomProvider {
    fn id(&self) -> &str {
        &self.source.id
    }

    fn platform(&self) -> &str {
        &self.source.platform
    }

    fn platform_name(&self) -> String {
        self.source.platform_name.clone()
    }

    fn name(&self, label: &str) -> String {
        render(&self.source.name_template, label)
    }

    fn url(&self, label: &str) -> String {
        render(&self.source.url_template, label)
    }

    fn paid(&self) -> bool {
        self.source.paid
    }

    fn normalize_label(&self, label: &str) -> String {
        if self.source.case_sensitive {
            label.to_string()
        } else {
            label.to_ascii_lowercase()
        }
    }

    fn chars(&self) -> crate::label::Chars {
        self.source.chars
    }

    fn min_length(&self) -> usize {
        self.source.min_length
    }

    fn max_length(&self) -> Option<usize> {
        (self.source.max_length > 0).then_some(self.source.max_length)
    }

    fn check(&self, label: &str) -> Result<Outcome, Error> {
        (self.check)(&self.source, self.transport.as_ref(), label)
    }
}

/// build_custom_provider returns the provider for id when the manifest and a
/// custom spec both know it.
pub fn build_custom_provider(
    manifest: &Manifest,
    id: &str,
    transport: Arc<dyn Transport>,
) -> Result<CustomProvider, Error> {
    let spec = custom_specs()
        .iter()
        .find(|spec| spec.id == id)
        .ok_or_else(|| Error::NotFound(format!("custom spec {id}")))?;
    let source = manifest
        .source(id)
        .ok_or_else(|| Error::NotFound(format!("manifest source {id}")))?
        .clone();
    Ok(CustomProvider::new(source, spec.check, transport))
}
