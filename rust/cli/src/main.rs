mod providers;
mod serve;

use std::io::{self, Write};
use std::path::PathBuf;
use std::sync::Arc;

use namecheck_core::{
    build_custom_provider, build_dns_provider, build_status_provider, custom_specs, dns_hosts,
    status_specs, valid_label, Engine, Fixture, HttpTransport, Manifest, Provider,
    RecordingTransport, SystemResolver, Transport,
};

const USAGE: &str = "usage: namecheck <command> [args]

Commands:
  check <label> [--live] [--manifest PATH] [--fixtures DIR] [--json]
      check a label; --live uses real HTTP/DNS, otherwise replay fixtures
  serve [--addr 127.0.0.1:8080] [--manifest PATH] [--budget-ms MILLIS]
      serve GET /v1/check?name=<label> as NDJSON for local development
      (the registry is compiled in; --manifest overrides it; --budget-ms caps
      the stream and leaves unfinished sources unchecked)
  sources [--manifest PATH]
      list the source inventory
  capture <label> [--source ID[,ID]] [--out DIR] [--manifest PATH]
      run providers live and write their fixtures (default: every source)
";

fn main() {
    if let Err(err) = run(std::env::args().skip(1).collect()) {
        if is_broken_pipe(err.as_ref()) {
            return;
        }
        eprintln!("namecheck: {err}");
        std::process::exit(1);
    }
}

fn is_broken_pipe(err: &(dyn std::error::Error + 'static)) -> bool {
    err.downcast_ref::<io::Error>()
        .is_some_and(|err| err.kind() == io::ErrorKind::BrokenPipe)
}

fn run(args: Vec<String>) -> Result<(), Box<dyn std::error::Error>> {
    let Some(command) = args.first().map(String::as_str) else {
        print!("{USAGE}");
        return Ok(());
    };
    match command {
        "check" => check(&args[1..]),
        "capture" => capture(&args[1..]),
        "serve" => serve::run(&args[1..]),
        "sources" => sources(&args[1..]),
        "help" | "-h" | "--help" => {
            print!("{USAGE}");
            Ok(())
        }
        other => Err(format!("unknown command {other:?}\n\n{USAGE}").into()),
    }
}

fn check(args: &[String]) -> Result<(), Box<dyn std::error::Error>> {
    let label = args
        .first()
        .filter(|arg| !arg.starts_with("--"))
        .ok_or("check requires a <label> argument")?
        .to_string();
    if !namecheck_core::valid_label(&label) {
        return Err(format!("invalid label {label:?}").into());
    }

    let mut manifest_path: Option<PathBuf> = None;
    let mut fixtures_dir = PathBuf::from("fixtures");
    let mut json = false;
    let mut live = false;
    let mut i = 1;
    while i < args.len() {
        match args[i].as_str() {
            "--manifest" => {
                manifest_path = Some(PathBuf::from(
                    args.get(i + 1).ok_or("--manifest requires a path")?,
                ));
                i += 2;
            }
            "--fixtures" => {
                fixtures_dir =
                    PathBuf::from(args.get(i + 1).ok_or("--fixtures requires a directory")?);
                i += 2;
            }
            "--json" => {
                json = true;
                i += 1;
            }
            "--live" => {
                live = true;
                i += 1;
            }
            flag => return Err(format!("unknown flag {flag:?}").into()),
        }
    }

    let manifest = match manifest_path {
        Some(path) => Manifest::load(path)?,
        None => Manifest::embedded()?,
    };
    manifest.validate()?;
    let providers: Vec<Arc<dyn Provider>> = if live {
        providers::live(&manifest)?
    } else {
        providers::replay(&manifest, &fixtures_dir.join(&label))?
    };
    if providers.is_empty() {
        return Err(format!("no providers for {label}").into());
    }

    let results = Engine::new(providers).scan_all(&label);
    let stdout = io::stdout();
    let mut out = io::BufWriter::new(stdout.lock());
    if json {
        serde_json::to_writer_pretty(&mut out, &results)?;
        writeln!(out)?;
    } else {
        for result in results {
            writeln!(
                out,
                "{:<14} {:<9} {:<24} {}",
                result.id,
                result.status.as_str(),
                result.name,
                result.url
            )?;
        }
    }
    out.flush()?;
    Ok(())
}

fn sources(args: &[String]) -> Result<(), Box<dyn std::error::Error>> {
    let mut manifest_path: Option<PathBuf> = None;
    let mut i = 0;
    while i < args.len() {
        match args[i].as_str() {
            "--manifest" => {
                manifest_path = Some(PathBuf::from(
                    args.get(i + 1).ok_or("--manifest requires a path")?,
                ));
                i += 2;
            }
            flag => return Err(format!("unknown flag {flag:?}").into()),
        }
    }

    let manifest = match manifest_path {
        Some(path) => Manifest::load(path)?,
        None => Manifest::embedded()?,
    };
    manifest.validate()?;
    let stdout = io::stdout();
    let mut out = io::BufWriter::new(stdout.lock());
    for source in &manifest.sources {
        let keyed = if source.keyed { " keyed" } else { "" };
        writeln!(
            out,
            "{:<18} {:<14} {}{}",
            source.id, source.platform, source.name, keyed
        )?;
    }
    out.flush()?;
    Ok(())
}

fn capture(args: &[String]) -> Result<(), Box<dyn std::error::Error>> {
    let mut label = None;
    let mut wanted: Vec<String> = Vec::new();
    let mut out_dir = PathBuf::from("fixtures");
    let mut manifest_path: Option<PathBuf> = None;
    let mut i = 0;
    while i < args.len() {
        match args[i].as_str() {
            "--source" => {
                let value = args.get(i + 1).ok_or("--source requires an id")?;
                for id in value.split(',') {
                    wanted.push(id.trim().to_string());
                }
                i += 2;
            }
            "--out" => {
                out_dir = PathBuf::from(args.get(i + 1).ok_or("--out requires a directory")?);
                i += 2;
            }
            "--manifest" => {
                manifest_path = Some(PathBuf::from(
                    args.get(i + 1).ok_or("--manifest requires a path")?,
                ));
                i += 2;
            }
            flag if flag.starts_with("--") => return Err(format!("unknown flag {flag:?}").into()),
            value => {
                if label.is_some() {
                    return Err(format!("unexpected argument {value:?}").into());
                }
                label = Some(value.to_string());
                i += 1;
            }
        }
    }
    let label = label.ok_or("capture requires a label")?;
    if !valid_label(&label) {
        return Err(format!("invalid label {label:?}").into());
    }
    let manifest = match manifest_path {
        Some(path) => Manifest::load(path)?,
        None => Manifest::embedded()?,
    };
    manifest.validate()?;

    if wanted.is_empty() {
        for spec in status_specs() {
            wanted.push(spec.id.to_string());
        }
        for spec in custom_specs() {
            wanted.push(spec.id.to_string());
        }
        for (id, _) in dns_hosts() {
            wanted.push(id.to_string());
        }
    }

    let dir = out_dir.join(&label);
    std::fs::create_dir_all(&dir)?;
    let http: Arc<dyn Transport> = Arc::new(HttpTransport::new());
    let mut written = 0;

    for id in &wanted {
        let recorder = Arc::new(RecordingTransport::new(Arc::clone(&http)));
        let provider: Arc<dyn Provider> = if status_specs().iter().any(|s| s.id == id) {
            Arc::new(build_status_provider(&manifest, id, recorder.clone())?)
        } else if custom_specs().iter().any(|s| s.id == id) {
            Arc::new(build_custom_provider(&manifest, id, recorder.clone())?)
        } else if dns_hosts().iter().any(|(known, _)| known == id) {
            Arc::new(build_dns_provider(&manifest, id, Arc::new(SystemResolver))?)
        } else {
            return Err(format!("unknown source {id:?}").into());
        };

        let started = std::time::Instant::now();
        let checked = provider.check(&label);
        let duration_ms = started.elapsed().as_millis() as u64;

        let (status, detail, error) = match &checked {
            Ok(outcome) => (
                Some(outcome.status.as_str().to_string()),
                (!outcome.detail.is_empty()).then(|| outcome.detail.clone()),
                None,
            ),
            Err(err) => (None, None, Some(err.to_string())),
        };
        let fixture = Fixture {
            source: id.clone(),
            platform: provider.platform().to_string(),
            platform_name: provider.platform_name(),
            label: label.clone(),
            name: provider.name(&label),
            url: checked
                .as_ref()
                .ok()
                .and_then(|outcome| outcome.url.clone())
                .unwrap_or_else(|| provider.url(&label)),
            paid: provider.paid(),
            status: status.clone(),
            detail,
            error,
            duration_ms,
            requests: recorder.take(),
        };
        let path = dir.join(format!("{id}.json"));
        fixture.save(&path)?;
        written += 1;
        let status = status.unwrap_or_else(|| "error".to_string());
        println!("{:<18} {:<10} {}", id, status, path.display());
    }
    println!("\n{written} fixture(s) written to {}", dir.display());
    Ok(())
}
