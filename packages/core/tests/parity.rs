use std::path::{Path, PathBuf};
use std::sync::Arc;

use namecada_core::{
    build_custom_provider, build_dns_provider, build_status_provider, custom_specs, dns_hosts,
    status_specs, DnsError, Engine, Fixture, Manifest, Provider, ReplayTransport, Resolver, Status,
};

fn repo_path(relative: &str) -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR"))
        .join("../..")
        .join(relative)
}

fn fixture_path(label: &str, id: &str) -> PathBuf {
    repo_path(&format!("fixtures/{label}/{id}.json"))
}

const LABELS: &[&str] = &["acema", "google"];

#[test]
fn manifest_loads() {
    let manifest = Manifest::load(repo_path("fixtures/sources.json")).expect("manifest");
    let stats = namecada_core::registry_stats(&manifest);
    assert_eq!(stats.sources, manifest.sources.len());
    assert_eq!(stats.sources, stats.checked + stats.keyed + stats.deferred);
    assert!(stats.checked > 100);
}

#[test]
fn providers_match_fixtures() {
    let manifest = Manifest::load(repo_path("fixtures/sources.json")).expect("manifest");
    let mut checked = 0;

    for label in LABELS {
        for spec in status_specs() {
            let path = fixture_path(label, spec.id);
            if !path.exists() {
                continue;
            }
            let fixture =
                Fixture::load(&path).unwrap_or_else(|err| panic!("{} {label}: {err}", spec.id));
            let transport = Arc::new(ReplayTransport::new(&fixture));
            let provider = build_status_provider(&manifest, spec.id, transport.clone())
                .unwrap_or_else(|err| panic!("{} {label}: {err}", spec.id));
            assert_fixture(spec.id, label, &fixture, transport, || {
                provider.check(label)
            });
            checked += 1;
        }

        for spec in custom_specs() {
            let path = fixture_path(label, spec.id);
            if !path.exists() {
                continue;
            }
            let fixture =
                Fixture::load(&path).unwrap_or_else(|err| panic!("{} {label}: {err}", spec.id));
            let transport = Arc::new(ReplayTransport::new(&fixture));
            let provider = build_custom_provider(&manifest, spec.id, transport.clone())
                .unwrap_or_else(|err| panic!("{} {label}: {err}", spec.id));
            assert_fixture(spec.id, label, &fixture, transport, || {
                provider.check(label)
            });
            checked += 1;
        }
    }

    assert!(checked >= 120, "only {checked} provider checks ran");
}

fn assert_fixture(
    id: &str,
    label: &str,
    fixture: &Fixture,
    transport: Arc<ReplayTransport>,
    check: impl FnOnce() -> Result<namecada_core::Outcome, namecada_core::Error>,
) {
    if let Some(expected) = fixture.expected_status() {
        let outcome = check().unwrap_or_else(|err| panic!("{id} {label}: {err}"));
        assert_eq!(
            outcome.status,
            expected,
            "{id} {label}: expected {}, got {}",
            expected.as_str(),
            outcome.status.as_str()
        );
    } else {
        assert!(
            fixture.error.is_some(),
            "{id} {label}: missing status without a transport error"
        );
        assert!(
            check().is_err(),
            "{id} {label}: transport error did not propagate"
        );
    }
    assert_eq!(
        transport.remaining(),
        0,
        "{id} {label}: {} recorded request(s) left",
        transport.remaining()
    );
}

#[test]
fn engine_collects_every_provider_in_order() {
    let manifest = Manifest::load(repo_path("fixtures/sources.json")).expect("manifest");

    for label in LABELS {
        let mut providers: Vec<Arc<dyn Provider>> = Vec::new();
        let mut expected: Vec<(&str, Option<Status>)> = Vec::new();

        for spec in status_specs() {
            let path = fixture_path(label, spec.id);
            if !path.exists() {
                continue;
            }
            let fixture = Fixture::load(&path).unwrap();
            let transport = Arc::new(ReplayTransport::new(&fixture));
            let provider = build_status_provider(&manifest, spec.id, transport).unwrap();
            providers.push(Arc::new(provider));
            expected.push((spec.id, fixture.expected_status()));
        }
        for spec in custom_specs() {
            let path = fixture_path(label, spec.id);
            if !path.exists() {
                continue;
            }
            let fixture = Fixture::load(&path).unwrap();
            let transport = Arc::new(ReplayTransport::new(&fixture));
            let provider = build_custom_provider(&manifest, spec.id, transport).unwrap();
            providers.push(Arc::new(provider));
            expected.push((spec.id, fixture.expected_status()));
        }

        let engine = Engine::new(providers).with_concurrency(4);
        let results = engine.scan_all(label);
        assert_eq!(results.len(), expected.len());

        for (result, (id, status)) in results.iter().zip(expected) {
            assert_eq!(result.id, id);
            match status {
                Some(status) => assert_eq!(result.status, status, "{id} {label}"),
                None => {
                    assert_eq!(result.status, Status::Unknown, "{id} {label}");
                    assert_eq!(result.detail, "unreachable", "{id} {label}");
                }
            }
            assert_eq!(
                result.platform_name,
                manifest.source(id).unwrap().platform_name
            );
        }
    }
}

#[test]
fn dns_providers_match_fixtures() {
    let manifest = Manifest::load(repo_path("fixtures/sources.json")).expect("manifest");
    let mut checked = 0;

    for label in LABELS {
        for (id, host) in dns_hosts() {
            let path = fixture_path(label, id);
            if !path.exists() {
                continue;
            }
            let fixture = Fixture::load(&path).unwrap();
            let expected = fixture
                .expected_status()
                .unwrap_or_else(|| panic!("{id} {label}: missing status"));

            let resolver = Arc::new(FixtureResolver { expected, host });
            let provider = build_dns_provider(&manifest, id, resolver)
                .unwrap_or_else(|err| panic!("{id} {label}: {err}"));
            assert_eq!(
                provider.name(label),
                render(&manifest.source(id).unwrap().name_template, label)
            );
            let outcome = provider
                .check(label)
                .unwrap_or_else(|err| panic!("{id} {label}: {err}"));
            assert_eq!(outcome.status, expected, "{id} {label}");
            checked += 1;
        }
    }
    assert!(checked >= 16, "only {checked} DNS checks ran");
}

/// FixtureResolver maps the captured verdict back to a resolver answer. It
/// exercises the DNS provider's available/taken/unknown mapping; the real
/// resolver is covered by SystemResolver.
struct FixtureResolver {
    expected: Status,
    host: &'static str,
}

impl Resolver for FixtureResolver {
    fn lookup_host(&self, host: &str) -> Result<Vec<String>, DnsError> {
        assert!(host.ends_with(&format!(".{}", self.host)), "{host}");
        match self.expected {
            Status::Available => Err(DnsError::NotFound),
            Status::Taken => Ok(vec!["192.0.2.1".to_string()]),
            Status::Unknown => Err(DnsError::Other("timeout".to_string())),
            // A pending source never reaches the resolver: the deadline cut it.
            Status::Pending => Err(DnsError::Other("pending".to_string())),
            // DNS providers have no alphabet rules, so no fixture is invalid.
            Status::Invalid => Err(DnsError::Other("invalid".to_string())),
        }
    }
}

fn render(template: &str, label: &str) -> String {
    template.replace("%s", label)
}

#[test]
fn typed_status_round_trips() {
    for status in [Status::Available, Status::Taken, Status::Unknown] {
        assert_eq!(Status::parse(status.as_str()), Some(status));
    }
}
