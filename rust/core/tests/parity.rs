use std::path::{Path, PathBuf};
use std::sync::Arc;

use domain_browser_core::{
    build_status_provider, status_specs, Engine, Fixture, Manifest, Provider, ReplayTransport,
    Status,
};

fn repo_path(relative: &str) -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR"))
        .join("../..")
        .join(relative)
}

fn fixture_path(id: &str) -> PathBuf {
    repo_path(&format!("fixtures/acema/{id}.json"))
}

#[test]
fn manifest_loads() {
    let manifest = Manifest::load(repo_path("fixtures/sources.json")).expect("manifest");
    assert_eq!(manifest.source_count, manifest.sources.len());
    assert_eq!(
        manifest.source_count,
        manifest.enabled_count + manifest.keyed_count
    );
    assert!(manifest.enabled_count > 50);
}

#[test]
fn status_providers_match_fixtures() {
    let manifest = Manifest::load(repo_path("fixtures/sources.json")).expect("manifest");
    let mut checked = 0;

    for spec in status_specs() {
        let path = fixture_path(spec.id);
        if !path.exists() {
            continue;
        }
        let fixture = Fixture::load(&path).unwrap_or_else(|err| panic!("{}: {err}", spec.id));
        let expected = fixture
            .expected_status()
            .unwrap_or_else(|| panic!("{}: missing status", spec.id));

        let transport = Arc::new(ReplayTransport::new(&fixture));
        let provider = build_status_provider(&manifest, spec.id, transport.clone())
            .unwrap_or_else(|err| panic!("{}: {err}", spec.id));

        let outcome = provider
            .check("acema")
            .unwrap_or_else(|err| panic!("{}: {err}", spec.id));

        assert_eq!(
            outcome.status,
            expected,
            "{}: expected {}, got {}",
            spec.id,
            expected.as_str(),
            outcome.status.as_str()
        );
        assert_eq!(
            transport.remaining(),
            0,
            "{}: {} recorded request(s) left",
            spec.id,
            transport.remaining()
        );
        checked += 1;
    }

    assert!(checked >= 30, "only {checked} status providers checked");
}

#[test]
fn engine_collects_every_status_provider_in_order() {
    let manifest = Manifest::load(repo_path("fixtures/sources.json")).expect("manifest");
    let mut providers: Vec<Arc<dyn domain_browser_core::Provider>> = Vec::new();
    let mut expected = Vec::new();

    for spec in status_specs() {
        let path = fixture_path(spec.id);
        if !path.exists() {
            continue;
        }
        let fixture = Fixture::load(&path).unwrap();
        let transport = Arc::new(ReplayTransport::new(&fixture));
        let provider = build_status_provider(&manifest, spec.id, transport).unwrap();
        providers.push(Arc::new(provider));
        expected.push((spec.id, fixture.expected_status().unwrap()));
    }

    let engine = Engine::new(providers).with_concurrency(4);
    let results = engine.scan_all("acema");
    assert_eq!(results.len(), expected.len());

    for (result, (id, status)) in results.iter().zip(expected) {
        assert_eq!(result.id, id);
        assert_eq!(result.status, status, "{id}");
        assert_eq!(
            result.platform_name,
            manifest.source(id).unwrap().platform_name
        );
    }
}

#[test]
fn typed_status_round_trips() {
    for status in [Status::Available, Status::Taken, Status::Unknown] {
        assert_eq!(Status::parse(status.as_str()), Some(status));
    }
}
