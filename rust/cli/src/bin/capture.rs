use std::error::Error;
use std::path::PathBuf;
use std::sync::Arc;

use namecheck::{load_manifest, run_or_exit, Flags};
use namecheck_core::{
    build_custom_provider, build_dns_provider, build_status_provider, custom_specs, dns_hosts,
    status_specs, Fixture, HttpTransport, Manifest, Provider, RecordingTransport, SystemResolver,
    Transport,
};

fn main() {
    run_or_exit(run());
}

fn run() -> Result<(), Box<dyn Error>> {
    let args: Vec<String> = std::env::args().skip(1).collect();
    let flags = Flags::parse(
        &args,
        &[("source", true), ("out", true), ("manifest", true)],
    )?;
    let label = flags.label("capture")?;
    let manifest = load_manifest(flags.value("manifest"))?;

    let mut wanted: Vec<String> = flags
        .values("source")
        .into_iter()
        .flat_map(|value| value.split(','))
        .map(|id| id.trim().to_string())
        .filter(|id| !id.is_empty())
        .collect();
    if wanted.is_empty() {
        wanted.extend(status_specs().iter().map(|spec| spec.id.to_string()));
        wanted.extend(custom_specs().iter().map(|spec| spec.id.to_string()));
        wanted.extend(dns_hosts().iter().map(|(id, _)| id.to_string()));
    }

    let dir = PathBuf::from(flags.value("out").unwrap_or("fixtures")).join(label);
    std::fs::create_dir_all(&dir)?;
    let http: Arc<dyn Transport> = Arc::new(HttpTransport::new());
    let mut written = 0;

    for id in &wanted {
        let recorder = Arc::new(RecordingTransport::new(Arc::clone(&http)));
        let provider = build(&manifest, id, Arc::clone(&recorder))?;
        let fixture = Fixture::record(provider.as_ref(), label, &recorder);

        let path = dir.join(format!("{id}.json"));
        fixture.save(&path)?;
        written += 1;
        println!(
            "{:<18} {:<10} {}",
            id,
            fixture.status.as_deref().unwrap_or("error"),
            path.display()
        );
    }
    println!("\n{written} fixture(s) written to {}", dir.display());
    Ok(())
}

/// build returns the provider for an id, wired to the recording transport when
/// it makes HTTP requests.
fn build(
    manifest: &Manifest,
    id: &str,
    recorder: Arc<RecordingTransport>,
) -> Result<Arc<dyn Provider>, Box<dyn Error>> {
    if status_specs().iter().any(|spec| spec.id == id) {
        Ok(Arc::new(build_status_provider(manifest, id, recorder)?))
    } else if custom_specs().iter().any(|spec| spec.id == id) {
        Ok(Arc::new(build_custom_provider(manifest, id, recorder)?))
    } else if dns_hosts().iter().any(|(known, _)| *known == id) {
        Ok(Arc::new(build_dns_provider(
            manifest,
            id,
            Arc::new(SystemResolver),
        )?))
    } else {
        Err(format!("unknown source {id:?}").into())
    }
}
