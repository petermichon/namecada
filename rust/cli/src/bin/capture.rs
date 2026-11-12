use std::error::Error;
use std::path::PathBuf;
use std::sync::Arc;

use namecheck::{build, load_manifest, run_or_exit, Flags};
use namecheck_core::{
    custom_specs, dns_hosts, status_specs, Fixture, HttpTransport, RecordingTransport, Transport,
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
