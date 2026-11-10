mod providers;
mod serve;

use std::io::{self, Write};
use std::path::PathBuf;
use std::sync::Arc;

use domain_browser_core::{Engine, Manifest, Provider};

const USAGE: &str = "usage: domain-browser <command> [args]

Commands:
  check <label> [--live] [--manifest PATH] [--fixtures DIR] [--json]
      check a label; --live uses real HTTP/DNS, otherwise replay fixtures
  serve [--addr 127.0.0.1:8080] [--origin URL] [--manifest PATH]
      serve GET /check?name=<label> as NDJSON
  sources [--manifest PATH]
      list the source inventory
";

fn main() {
    if let Err(err) = run(std::env::args().skip(1).collect()) {
        if is_broken_pipe(err.as_ref()) {
            return;
        }
        eprintln!("domain-browser: {err}");
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
        .to_ascii_lowercase();
    if !valid_label(&label) {
        return Err(format!("invalid label {label:?}").into());
    }

    let mut manifest_path = PathBuf::from("fixtures/sources.json");
    let mut fixtures_dir = PathBuf::from("fixtures");
    let mut json = false;
    let mut live = false;
    let mut i = 1;
    while i < args.len() {
        match args[i].as_str() {
            "--manifest" => {
                manifest_path = PathBuf::from(args.get(i + 1).ok_or("--manifest requires a path")?);
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

    let manifest = Manifest::load(&manifest_path)?;
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
    let mut manifest_path = PathBuf::from("fixtures/sources.json");
    let mut i = 0;
    while i < args.len() {
        match args[i].as_str() {
            "--manifest" => {
                manifest_path = PathBuf::from(args.get(i + 1).ok_or("--manifest requires a path")?);
                i += 2;
            }
            flag => return Err(format!("unknown flag {flag:?}").into()),
        }
    }

    let manifest = Manifest::load(manifest_path)?;
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

pub(crate) fn valid_label(label: &str) -> bool {
    !label.is_empty()
        && label.len() <= 63
        && label
            .chars()
            .all(|ch| ch.is_ascii_lowercase() || ch.is_ascii_digit() || ch == '-')
        && !label.starts_with('-')
        && !label.ends_with('-')
}
