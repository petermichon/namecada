//! Sources read from server-rendered HTML: titles, meta tags and redirects.

use crate::manifest::Source;
use crate::provider::{Error, HttpResponse, Outcome, Transport};

const HTML_ANY_ACCEPT: &str = "text/html,*/*";

use super::shared::*;

pub(crate) fn check_threads(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    let url = format!("https://www.threads.net/@{label}");
    let response = get(transport, &url, HTML_ANY_ACCEPT)?;
    if response.status != 200 {
        return Ok(unknown_response(&response));
    }
    let Some(title) = html_title(&response) else {
        return Ok(unknown());
    };
    // A profile title names the handle, either after a display name
    // ("Sofyan Boudouni (@sofyan) • Threads") or on its own
    // ("@acema • Threads"). The missing page is "Threads • Log in".
    let title = title.to_ascii_lowercase();
    let named = title.contains(&format!("(@{label})"))
        || title.contains(&format!("(&#064;{label})"))
        || title.starts_with(&format!("@{label} "))
        || title.starts_with(&format!("&#064;{label} "));
    if named {
        return Ok(taken());
    }
    if title.contains("log in") {
        return Ok(available());
    }
    Ok(unknown())
}

pub(crate) fn check_freelancer(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    let url = format!("https://www.freelancer.com/u/{label}");
    let response = get(transport, &url, HTML_ANY_ACCEPT)?;
    if response.status == 404 {
        return Ok(available());
    }
    if response.status != 200 {
        return Ok(unknown_response(&response));
    }
    let Some(title) = html_title(&response) else {
        return Ok(unknown());
    };
    // A claimed profile is titled "{name} Profile | Freelancer". An unclaimed
    // name can still answer 200 with a *different* profile ("Mdk Profile"), so
    // the requested name has to appear in the title.
    let title = title.to_ascii_lowercase();
    let label = label.to_ascii_lowercase();
    if title.starts_with(&format!("{label} profile")) {
        return Ok(taken());
    }
    if title.ends_with("| freelancer") {
        return Ok(available());
    }
    Ok(unknown())
}

pub(crate) fn check_twitch(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    let url = format!("https://www.twitch.tv/{label}");
    let response = get(transport, &url, HTML_ANY_ACCEPT)?;
    if response.status != 200 {
        return Ok(unknown_response(&response));
    }
    let Some(title) = html_title(&response) else {
        return Ok(unknown());
    };
    // A channel title is "{name} - Twitch"; a free name falls back to "Twitch".
    if title.to_ascii_lowercase().ends_with(" - twitch") {
        return Ok(taken());
    }
    if title.eq_ignore_ascii_case("twitch") {
        return Ok(available());
    }
    Ok(unknown())
}

pub(crate) fn check_duolingo(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    let url = format!("https://www.duolingo.com/profile/{label}");
    let response = get(transport, &url, HTML_ANY_ACCEPT)?;
    if response.status != 200 {
        return Ok(unknown_response(&response));
    }
    // Only og:title carries the name; <title> is "Duolingo" for every profile.
    let Some(title) = og_title(&response) else {
        return Ok(unknown());
    };
    let title = title.to_ascii_lowercase();
    if title.ends_with("on duolingo") {
        return Ok(taken());
    }
    if title.contains("learn a language") {
        return Ok(available());
    }
    Ok(unknown())
}

pub(crate) fn check_telegram(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    let url = format!("https://t.me/{label}");
    let response = get(transport, &url, HTML_ANY_ACCEPT)?;
    if response.status != 200 {
        return Ok(unknown_response(&response));
    }
    let Some(title) = html_title(&response) else {
        return Ok(unknown());
    };
    let title = title.to_ascii_lowercase();
    // "Telegram: View @name" is a real account; the contact card is the miss.
    if title.starts_with("telegram: view @") {
        return Ok(taken());
    }
    if title.starts_with("telegram: contact @") {
        return Ok(available());
    }
    Ok(unknown())
}

pub(crate) fn check_instagram(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    let url = format!("https://www.instagram.com/{label}/");
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    if response.status != 200 {
        return Ok(unknown_response(&response));
    }
    let text = text_body(&response);
    if text.contains(&format!("\"username\":\"{label}\"")) {
        return Ok(taken());
    }
    if let Some(index) = text.find("<title>") {
        if text[index + "<title>".len()..].starts_with(label) {
            return Ok(taken());
        }
    }
    if text.contains("<title>Instagram</title>") {
        return Ok(available());
    }
    Ok(unknown())
}

pub(crate) fn check_dribbble(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    let url = format!("https://dribbble.com/{label}");
    let response = get(transport, &url, HTML_ANY_ACCEPT)?;
    match response.status {
        200 => Ok(taken()),
        404 => {
            if text_body(&response).contains("public yet") {
                Ok(taken())
            } else {
                Ok(available())
            }
        }
        _ => Ok(unknown()),
    }
}

pub(crate) fn check_giphy(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    let url = format!("https://giphy.com/{label}");
    let fetched = get_follow(transport, &url, HTML_ANY_ACCEPT)?;
    if fetched.response.status != 200 {
        return Ok(unknown());
    }
    let path = url::Url::parse(&fetched.url)
        .map(|url| url.path().to_string())
        .unwrap_or_default();
    if path.starts_with("/explore/") {
        return Ok(available());
    }
    if path == format!("/{label}") {
        return Ok(taken());
    }
    if text_body(&fetched.response)
        .to_ascii_lowercase()
        .contains("gifs on giphy")
    {
        return Ok(taken());
    }
    Ok(unknown())
}

/// Bandcamp sends a free artist subdomain to `/signup` and a claimed one into
/// its own content, so the redirect target is the verdict and must not be
/// followed.
pub(crate) fn check_bandcamp(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    let url = format!("https://{label}.bandcamp.com");
    let response = get_once(transport, &url, DEFAULT_ACCEPT)?;
    match (response.status, response.headers.get("location")) {
        (200, _) => Ok(taken()),
        (301 | 302 | 303 | 307 | 308, Some(location)) => {
            if location.to_ascii_lowercase().contains("/signup") {
                Ok(available())
            } else {
                Ok(taken())
            }
        }
        _ => Ok(unknown()),
    }
}

pub(crate) fn normalize_space(text: &str) -> String {
    text.split_whitespace().collect::<Vec<_>>().join(" ")
}

/// html_title reads the document title, which profile pages set for link
/// previews even when the body is a client-rendered shell.
pub(crate) fn html_title(response: &HttpResponse) -> Option<String> {
    let html = text_body(response);
    let lower = html.to_ascii_lowercase();
    let start = lower.find("<title")?;
    let open_end = lower[start..].find('>')? + start;
    let close = lower[open_end..].find("</title>")? + open_end;
    Some(normalize_space(&html[open_end + 1..close]))
}

/// og_title reads the og:title meta tag, which some shells only render there.
pub(crate) fn og_title(response: &HttpResponse) -> Option<String> {
    let html = text_body(response);
    let lower = html.to_ascii_lowercase();
    let mut from = 0;
    while let Some(rel) = lower[from..].find("og:title") {
        let index = from + rel;
        let tag_start = lower[..index].rfind("<meta").unwrap_or(index);
        let tag_end = lower[index..].find('>')? + index;
        if let Some(value) = attribute(&html[tag_start..tag_end], "content") {
            return Some(normalize_space(&value));
        }
        from = index + "og:title".len();
    }
    None
}

pub(crate) fn attribute(tag: &str, name: &str) -> Option<String> {
    let lower = tag.to_ascii_lowercase();
    let key = format!("{name}=");
    let start = lower.find(&key)? + key.len();
    let rest = &tag[start..];
    let quote = rest.chars().next()?;
    if quote != '"' && quote != '\'' {
        return None;
    }
    let value = &rest[1..];
    Some(value[..value.find(quote)?].to_string())
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::fixture::{Fixture, ReplayTransport};
    use crate::manifest::Manifest;
    use crate::provider::Status;

    /// Freelancer serves 200 for some unclaimed names by showing a *different*
    /// profile, so only a title that names the requested handle counts.
    #[test]
    fn freelancer_only_counts_a_title_that_names_the_handle() {
        fn outcome(title: &str) -> Status {
            let fixture: Fixture = serde_json::from_value(serde_json::json!({
                "source": "freelancer",
                "platform": "freelancer",
                "label": "acema",
                "name": "acema",
                "url": "https://www.freelancer.com/u/acema",
                "requests": [{
                    "method": "GET",
                    "url": "https://www.freelancer.com/u/acema",
                    "response": {
                        "status": 200,
                        "body": format!("<html><head><title>{title}</title></head></html>")
                    }
                }]
            }))
            .expect("fixture");
            let transport = ReplayTransport::new(&fixture);
            let source = Manifest::embedded()
                .expect("manifest")
                .source("freelancer")
                .expect("freelancer source")
                .clone();
            check_freelancer(&source, &transport, "acema")
                .expect("check")
                .status
        }

        assert_eq!(outcome("Acema Profile | Freelancer"), Status::Taken);
        // The false 200: a missing name resolved to somebody else's profile.
        assert_eq!(outcome("Mdk Profile | Freelancer"), Status::Available);
    }

    /// A profile title names the handle whether or not the account has a
    /// display name; the missing page is the login shell.
    #[test]
    fn threads_reads_the_handle_out_of_the_title() {
        fn outcome(title: &str) -> Status {
            let fixture: Fixture = serde_json::from_value(serde_json::json!({
                "source": "threads",
                "platform": "threads",
                "label": "acema",
                "name": "acema",
                "url": "https://www.threads.net/@acema",
                "requests": [{
                    "method": "GET",
                    "url": "https://www.threads.net/@acema",
                    "response": {
                        "status": 200,
                        "body": format!("<html><head><title>{title}</title></head></html>")
                    }
                }]
            }))
            .expect("fixture");
            let transport = ReplayTransport::new(&fixture);
            let source = Manifest::embedded()
                .expect("manifest")
                .source("threads")
                .expect("threads source")
                .clone();
            check_threads(&source, &transport, "acema")
                .expect("check")
                .status
        }

        assert_eq!(
            outcome("Acema (&#064;acema) &#x2022; Threads, Say more"),
            Status::Taken
        );
        // An account with no display name puts the handle first.
        assert_eq!(
            outcome("&#064;acema &#x2022; Threads, Say more"),
            Status::Taken
        );
        assert_eq!(outcome("Threads &#x2022; Log in"), Status::Available);
        assert_eq!(outcome("Threads"), Status::Unknown);
    }
}
