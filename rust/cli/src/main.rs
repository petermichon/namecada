mod serve;

use std::error::Error;
use std::io::{self, Write};
use std::path::PathBuf;
use std::sync::Arc;

use namecheck_core::{
    build_custom_provider, build_dns_provider, build_status_provider, custom_specs, dns_hosts,
    live_providers, status_specs, valid_label, Engine, Fixture, HttpTransport, Manifest, Provider,
    RecordingTransport, SystemResolver, Transport,
};

const USAGE: &str = "usage: namecheck <command> [args]

Commands:
  check <label> [--manifest PATH] [--json]
      check a label against every live source
  serve [--addr 127.0.0.1:8080] [--manifest PATH]
      serve GET /v1/check?name=<label> as NDJSON for local development
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

fn is_broken_pipe(err: &(dyn Error + 'static)) -> bool {
    err.downcast_ref::<io::Error>()
        .is_some_and(|err| err.kind() == io::ErrorKind::BrokenPipe)
}

fn run(args: Vec<String>) -> Result<(), Box<dyn Error>> {
    let Some((command, rest)) = args.split_first() else {
        print!("{USAGE}");
        return Ok(());
    };
    match command.as_str() {
        "check" => check(rest),
        "capture" => capture(rest),
        "serve" => serve::run(rest),
        "help" | "-h" | "--help" => {
            print!("{USAGE}");
            Ok(())
        }
        other => Err(format!("unknown command {other:?}\n\n{USAGE}").into()),
    }
}

/// Flags is the whole argument grammar: `--name value`, `--switch`, and
/// positionals. Each command passes the flags it knows; anything else is an
/// error rather than a silently ignored typo.
pub(crate) struct Flags {
    pairs: Vec<(String, Option<String>)>,
    positionals: Vec<String>,
}

impl Flags {
    pub(crate) fn parse(args: &[String], spec: &[(&str, bool)]) -> Result<Self, Box<dyn Error>> {
        let mut pairs = Vec::new();
        let mut positionals = Vec::new();
        let mut i = 0;
        while i < args.len() {
            let arg = args[i].as_str();
            let Some(name) = arg.strip_prefix("--") else {
                positionals.push(arg.to_string());
                i += 1;
                continue;
            };
            let takes_value = spec
                .iter()
                .find(|(known, _)| *known == name)
                .map(|(_, takes_value)| *takes_value)
                .ok_or_else(|| format!("unknown flag {arg:?}"))?;
            if takes_value {
                let value = args
                    .get(i + 1)
                    .ok_or_else(|| format!("--{name} requires a value"))?;
                pairs.push((name.to_string(), Some(value.clone())));
                i += 2;
            } else {
                pairs.push((name.to_string(), None));
                i += 1;
            }
        }
        Ok(Self { pairs, positionals })
    }

    pub(crate) fn switch(&self, name: &str) -> bool {
        self.pairs.iter().any(|(known, _)| known == name)
    }

    pub(crate) fn value(&self, name: &str) -> Option<&str> {
        self.values(name).into_iter().next_back()
    }

    fn values(&self, name: &str) -> Vec<&str> {
        self.pairs
            .iter()
            .filter(|(known, _)| known == name)
            .filter_map(|(_, value)| value.as_deref())
            .collect()
    }

    fn label(&self, command: &str) -> Result<&str, Box<dyn Error>> {
        let label = match self.positionals.as_slice() {
            [label] => label,
            [] => return Err(format!("{command} requires a <label> argument").into()),
            _ => return Err(format!("{command} takes one label").into()),
        };
        if !valid_label(label) {
            return Err(format!("invalid label {label:?}").into());
        }
        Ok(label)
    }
}

pub(crate) fn load_manifest(path: Option<&str>) -> Result<Manifest, Box<dyn Error>> {
    let manifest = match path {
        Some(path) => Manifest::load(path)?,
        None => Manifest::embedded()?,
    };
    manifest.validate()?;
    Ok(manifest)
}

fn check(args: &[String]) -> Result<(), Box<dyn Error>> {
    let flags = Flags::parse(args, &[("manifest", true), ("json", false)])?;
    let label = flags.label("check")?;
    let manifest = load_manifest(flags.value("manifest"))?;

    let results = Engine::new(live_providers(&manifest)?).scan_all(label);
    let stdout = io::stdout();
    let mut out = io::BufWriter::new(stdout.lock());
    if flags.switch("json") {
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

fn capture(args: &[String]) -> Result<(), Box<dyn Error>> {
    let flags = Flags::parse(args, &[("source", true), ("out", true), ("manifest", true)])?;
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
