//! RDAP domain registries.

use crate::manifest::Source;
use crate::provider::{Error, Outcome, Transport};

use super::*;

pub(crate) const RESERVED_LABELS: &[&str] = &["example", "test", "invalid", "localhost"];

pub(crate) fn check_domain(
    source: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    if RESERVED_LABELS.contains(&label) {
        return Ok(taken());
    }
    let tld = source
        .id
        .strip_prefix("domain.")
        .ok_or_else(|| Error::Invalid(format!("domain source has id {}", source.id)))?;
    let base = rdap_base(tld).ok_or_else(|| Error::Invalid(format!("no RDAP base for .{tld}")))?;
    let url = format!("{base}domain/{label}.{tld}");
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    match response.status {
        200 => Ok(taken()),
        404 => Ok(available()),
        _ => Ok(unknown()),
    }
}

pub(crate) fn rdap_base(tld: &str) -> Option<&'static str> {
    match tld {
        "com" => Some("https://rdap.verisign.com/com/v1/"),
        "net" => Some("https://rdap.verisign.com/net/v1/"),
        "org" => Some("https://rdap.publicinterestregistry.org/rdap/"),
        "dev" | "app" => Some("https://pubapi.registry.google/rdap/"),
        "io" | "ai" => Some("https://rdap.identitydigital.services/rdap/"),
        _ => None,
    }
}
