//! The label grammar shared by every transport: one bare name.

use serde::Deserialize;

/// Chars is the alphabet a source accepts. Sources may be stricter than the
/// input grammar; a name outside a source's alphabet simply cannot be
/// registered there.
#[derive(Debug, Clone, Copy, Default, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum Chars {
    /// Letters, digits and hyphens: domains and code hosts.
    #[default]
    Ldh,
    /// Adds underscores.
    Underscore,
    /// Adds dots and underscores: most social handles.
    Handle,
}

impl Chars {
    pub fn allows(self, label: &str) -> bool {
        label.chars().all(|ch| match self {
            Chars::Ldh => ch.is_ascii_alphanumeric() || ch == '-',
            Chars::Underscore => ch.is_ascii_alphanumeric() || matches!(ch, '-' | '_'),
            Chars::Handle => ch.is_ascii_alphanumeric() || matches!(ch, '-' | '_' | '.'),
        })
    }
}

/// valid_label accepts a non-empty name of ASCII letters, digits, dots,
/// underscores and hyphens, starting and ending with a letter or digit.
pub fn valid_label(label: &str) -> bool {
    let length_ok = (1..=63).contains(&label.len());
    let edges_ok = matches!(label.chars().next(), Some(ch) if ch.is_ascii_alphanumeric())
        && matches!(label.chars().last(), Some(ch) if ch.is_ascii_alphanumeric());
    let chars_ok = label.chars().all(allowed);
    length_ok && edges_ok && chars_ok
}

fn allowed(ch: char) -> bool {
    ch.is_ascii_alphanumeric() || matches!(ch, '-' | '_' | '.')
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn accepts_letters_numbers_and_inner_hyphens() {
        for label in ["a", "acema", "my-name", "a1-b2", &"a".repeat(63)] {
            assert!(valid_label(label), "{label}");
        }
    }

    #[test]
    fn accepts_dots_and_underscores_inside_a_name() {
        for label in ["example.com", "example_com", "a.b_c-d"] {
            assert!(valid_label(label), "{label}");
        }
    }

    #[test]
    fn rejects_words_outside_the_grammar() {
        for label in [
            "",
            "-a",
            "a-",
            ".a",
            "a.",
            "_a",
            "a_",
            "a b",
            "a/b",
            "É",
            &"a".repeat(64),
        ] {
            assert!(!valid_label(label), "{label}");
        }
    }
}
