//! The label grammar shared by every transport: one bare DNS-ish label.

/// valid_label accepts a non-empty label of ASCII letters, digits, and inner
/// hyphens. Sources may impose tighter limits; this is the outer bound.
pub fn valid_label(label: &str) -> bool {
    !label.is_empty()
        && label.len() <= 63
        && label
            .chars()
            .all(|ch| ch.is_ascii_alphanumeric() || ch == '-')
        && !label.starts_with('-')
        && !label.ends_with('-')
}
