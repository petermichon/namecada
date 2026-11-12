//! Shared plumbing for the two development binaries: argument flags, the
//! compiled registry, and provider construction.

use std::error::Error;
use std::io;
use std::sync::Arc;

use namecheck_core::{
    build_custom_provider, build_dns_provider, build_status_provider, custom_specs, dns_hosts,
    status_specs, valid_label, Manifest, Provider, RecordingTransport, SystemResolver,
};

/// Flags is the whole argument grammar: `--name value`, `--switch`, and
/// positionals. Each binary passes the flags it knows; anything else is an
/// error rather than a silently ignored typo.
pub struct Flags {
    pairs: Vec<(String, Option<String>)>,
    positionals: Vec<String>,
}

impl Flags {
    pub fn parse(args: &[String], spec: &[(&str, bool)]) -> Result<Self, Box<dyn Error>> {
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

    pub fn switch(&self, name: &str) -> bool {
        self.pairs.iter().any(|(known, _)| known == name)
    }

    pub fn value(&self, name: &str) -> Option<&str> {
        self.values(name).into_iter().next_back()
    }

    pub fn values(&self, name: &str) -> Vec<&str> {
        self.pairs
            .iter()
            .filter(|(known, _)| known == name)
            .filter_map(|(_, value)| value.as_deref())
            .collect()
    }

    pub fn label(&self, command: &str) -> Result<&str, Box<dyn Error>> {
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

pub fn load_manifest(path: Option<&str>) -> Result<Manifest, Box<dyn Error>> {
    let manifest = match path {
        Some(path) => Manifest::load(path)?,
        None => Manifest::embedded()?,
    };
    manifest.validate()?;
    Ok(manifest)
}

/// build returns the provider for an id, wired to the recording transport when
/// it makes HTTP requests.
pub fn build(
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

/// run_or_exit prints the error and exits, except for a closed pipe, which is
/// how `... | head` ends.
pub fn run_or_exit(result: Result<(), Box<dyn Error>>) {
    if let Err(err) = result {
        if err
            .downcast_ref::<io::Error>()
            .is_some_and(|err| err.kind() == io::ErrorKind::BrokenPipe)
        {
            return;
        }
        eprintln!("namecheck: {err}");
        std::process::exit(1);
    }
}
