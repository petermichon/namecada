//! The label grammar shared by every transport: one bare DNS-ish label.

/// valid_label accepts a non-empty label of ASCII letters, digits, and inner
/// hyphens. Sources may impose tighter limits; this is the outer bound.
pub fn valid_label(label: &str) -> bool {
    let length_ok = (1..=63).contains(&label.len());
    let edges_ok = !label.starts_with('-') && !label.ends_with('-');
    let chars_ok = label.chars().all(allowed);
    length_ok && edges_ok && chars_ok
}

fn allowed(ch: char) -> bool {
    ch.is_ascii_alphanumeric() || ch == '-'
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
    fn rejects_words_outside_the_grammar() {
        for label in ["", "-a", "a-", "a.b", "a_b", "É", &"a".repeat(64)] {
            assert!(!valid_label(label), "{label}");
        }
    }
}
