//! DNS providers: a free-subdomain namespace resolves only when claimed, so a
//! successful lookup is taken and NXDOMAIN is available.

use std::sync::Arc;

use dns_lookup::{LookupError, LookupErrorKind};

use crate::manifest::{Manifest, Source};
use crate::provider::{Error, Outcome, Provider, Status};

/// DnsError distinguishes a not-found answer from a transient failure.
#[derive(Debug)]
pub enum DnsError {
    NotFound,
    Other(String),
}

/// Resolver answers host lookups. Production uses the system resolver; tests
/// inject a fake.
pub trait Resolver: Send + Sync {
    fn lookup_host(&self, host: &str) -> Result<Vec<String>, DnsError>;
}

/// SystemResolver uses the operating system's resolver. This is the native
/// path; the web build will use DoH instead.
pub struct SystemResolver;

impl Resolver for SystemResolver {
    fn lookup_host(&self, host: &str) -> Result<Vec<String>, DnsError> {
        // dns-lookup exposes the getaddrinfo code instead of a message: the
        // message from std is localized (GTK calls setlocale), which made a
        // French desktop classify NXDOMAIN as a transient failure.
        match dns_lookup::lookup_host(host) {
            Ok(addrs) => {
                let addresses: Vec<String> = addrs.map(|ip| ip.to_string()).collect();
                if addresses.is_empty() {
                    Err(DnsError::NotFound)
                } else {
                    Ok(addresses)
                }
            }
            Err(err) => Err(classify(&err)),
        }
    }
}

/// classify maps a getaddrinfo error to a verdict: only "no such name" means the
/// name is free; everything else is transient and must not be cached.
fn classify(err: &LookupError) -> DnsError {
    if is_absent(err.kind()) {
        DnsError::NotFound
    } else {
        DnsError::Other(err.to_string())
    }
}

fn is_absent(kind: LookupErrorKind) -> bool {
    matches!(kind, LookupErrorKind::NoName | LookupErrorKind::NoData)
}

/// DnsProvider checks `{label}.{host}` for a namespace without wildcard DNS.
pub struct DnsProvider {
    source: Source,
    host: &'static str,
    resolver: Arc<dyn Resolver>,
}

impl DnsProvider {
    pub fn new(source: Source, host: &'static str, resolver: Arc<dyn Resolver>) -> Self {
        Self {
            source,
            host,
            resolver,
        }
    }
}

impl Provider for DnsProvider {
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
        let host = format!("{label}.{}", self.host);
        match self.resolver.lookup_host(&host) {
            Ok(_) => Ok(Outcome::new(Status::Taken)),
            Err(DnsError::NotFound) => Ok(Outcome::new(Status::Available)),
            Err(DnsError::Other(message)) => Ok(Outcome {
                status: Status::Unknown,
                detail: format!("dns: {message}"),
                url: None,
            }),
        }
    }
}

/// dns_hosts returns the DNS provider ids and their parent domains.
pub fn dns_hosts() -> &'static [(&'static str, &'static str)] {
    &[
        ("euorg", "eu.org"),
        ("duckdns", "duckdns.org"),
        ("noip", "no-ip.org"),
        ("dynv6", "dynv6.net"),
        ("notedlol", "noted.lol"),
        ("isagooddev", "is-a-good.dev"),
        ("pagesdev", "pages.dev"),
        ("fly", "fly.dev"),
        ("ddnsnet", "ddns.net"),
        ("ddnsgeek", "ddnsgeek.com"),
        ("hopto", "hopto.org"),
        ("myftp", "myftp.org"),
        ("zapto", "zapto.org"),
    ]
}

/// build_dns_provider returns the DNS provider for id.
pub fn build_dns_provider(
    manifest: &Manifest,
    id: &str,
    resolver: Arc<dyn Resolver>,
) -> Result<DnsProvider, Error> {
    let host = dns_hosts()
        .iter()
        .find(|(source, _)| *source == id)
        .map(|(_, host)| *host)
        .ok_or_else(|| Error::NotFound(format!("dns source {id}")))?;
    let source = manifest
        .source(id)
        .ok_or_else(|| Error::NotFound(format!("manifest source {id}")))?
        .clone();
    Ok(DnsProvider::new(source, host, resolver))
}

fn render(template: &str, label: &str) -> String {
    template.replace("%s", label)
}

#[cfg(test)]
mod tests {
    use super::*;

    struct Failing;

    impl Resolver for Failing {
        fn lookup_host(&self, _: &str) -> Result<Vec<String>, DnsError> {
            Err(DnsError::Other("temporary failure".to_string()))
        }
    }

    #[test]
    fn reports_transient_dns_failures() {
        let manifest = Manifest::embedded().expect("manifest");
        let provider = build_dns_provider(&manifest, "duckdns", Arc::new(Failing)).unwrap();
        let outcome = provider.check("abcdx").expect("check");
        assert_eq!(outcome.status, Status::Unknown);
        assert_eq!(outcome.detail, "dns: temporary failure");
    }

    #[test]
    fn classifies_absent_names_as_not_found() {
        assert!(is_absent(LookupErrorKind::NoName));
        assert!(is_absent(LookupErrorKind::NoData));
        assert!(!is_absent(LookupErrorKind::Again));
        assert!(!is_absent(LookupErrorKind::Fail));
    }
}
