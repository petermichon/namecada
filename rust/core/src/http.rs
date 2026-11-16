//! Production HTTP transport. Redirects are left to the providers so their
//! final-URL handling (Giphy, etc.) stays explicit.

use std::collections::BTreeMap;
use std::time::Duration;

use crate::provider::{Error, HttpRequest, HttpResponse, Transport};

/// DEFAULT_TIMEOUT bounds one request, including reading the body.
pub const DEFAULT_TIMEOUT: Duration = Duration::from_secs(10);
/// DEFAULT_CONNECT_TIMEOUT bounds name resolution and connection setup,
/// including proxy negotiation and the TLS handshake.
pub const DEFAULT_CONNECT_TIMEOUT: Duration = Duration::from_secs(4);
/// DEFAULT_READ_TIMEOUT bounds the wait for response headers and for the body,
/// so a host that accepts and then stalls releases its worker early.
pub const DEFAULT_READ_TIMEOUT: Duration = Duration::from_secs(5);

/// HttpTransport executes requests with the system's HTTP stack.
pub struct HttpTransport {
    agent: ureq::Agent,
}

impl HttpTransport {
    pub fn new() -> Self {
        Self::with_timeout(DEFAULT_TIMEOUT)
    }

    /// with_timeout bounds one request, including reading the body, and keeps
    /// the default connect and read bounds.
    pub fn with_timeout(timeout: Duration) -> Self {
        Self::with_timeouts(timeout, DEFAULT_CONNECT_TIMEOUT, DEFAULT_READ_TIMEOUT)
    }

    /// with_timeouts also sets the phase bounds: `connect` covers resolution,
    /// socket setup and TLS, `read` covers the headers and the body. All three
    /// are capped by `global` when it is shorter.
    pub fn with_timeouts(global: Duration, connect: Duration, read: Duration) -> Self {
        let config = ureq::Agent::config_builder()
            .timeout_global(Some(global))
            .timeout_resolve(Some(connect))
            .timeout_connect(Some(connect))
            .timeout_recv_response(Some(read))
            .timeout_recv_body(Some(read))
            .http_status_as_error(false)
            .max_redirects(0)
            .build();
        Self {
            agent: config.into(),
        }
    }
}

impl Default for HttpTransport {
    fn default() -> Self {
        Self::new()
    }
}

impl Transport for HttpTransport {
    fn execute(&self, request: HttpRequest) -> Result<HttpResponse, Error> {
        let result = match request.method.as_str() {
            "GET" => {
                let mut builder = self.agent.get(&request.url);
                for (key, value) in &request.headers {
                    builder = builder.header(key, value);
                }
                builder.call()
            }
            "POST" => {
                let mut builder = self.agent.post(&request.url);
                for (key, value) in &request.headers {
                    builder = builder.header(key, value);
                }
                match request.body {
                    Some(body) => builder.send(body),
                    None => builder.send(""),
                }
            }
            other => {
                return Err(Error::Invalid(format!("unsupported method {other:?}")));
            }
        };

        let mut response = result.map_err(|err| Error::Transport(err.to_string()))?;
        let status = response.status().as_u16();
        let headers: BTreeMap<String, String> = response
            .headers()
            .iter()
            .map(|(key, value)| {
                (
                    key.as_str().to_ascii_lowercase(),
                    String::from_utf8_lossy(value.as_bytes()).into_owned(),
                )
            })
            .collect();
        let body = response
            .body_mut()
            .read_to_vec()
            .map_err(|err| Error::Transport(err.to_string()))?;
        Ok(HttpResponse {
            status,
            headers,
            body,
        })
    }
}

#[cfg(test)]
mod tests {
    use std::io::{Read, Write};
    use std::net::TcpListener;
    use std::thread;
    use std::time::Instant;

    use super::*;

    #[test]
    fn returns_4xx_as_a_response() {
        let listener = TcpListener::bind("127.0.0.1:0").expect("bind");
        let addr = listener.local_addr().expect("addr");
        let handle = thread::spawn(move || {
            let (mut stream, _) = listener.accept().expect("accept");
            let mut buf = [0_u8; 1024];
            let _ = stream.read(&mut buf);
            stream
                .write_all(b"HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\n\r\n")
                .expect("write");
        });

        let transport = HttpTransport::new();
        let response = transport
            .execute(HttpRequest {
                method: "GET".to_string(),
                url: format!("http://{addr}/missing"),
                headers: BTreeMap::new(),
                body: None,
            })
            .expect("request");
        assert_eq!(response.status, 404);
        assert!(response.body.is_empty());
        handle.join().expect("server");
    }

    #[test]
    fn a_stalled_response_releases_the_worker() {
        let listener = TcpListener::bind("127.0.0.1:0").expect("bind");
        let addr = listener.local_addr().expect("addr");
        let handle = thread::spawn(move || {
            let (mut stream, _) = listener.accept().expect("accept");
            let mut buf = [0_u8; 1024];
            let _ = stream.read(&mut buf);
            // Accept the request, then never answer: only the read bound can
            // end this call. Hold the socket until the client gives up.
            let mut sink = [0_u8; 256];
            while let Ok(read) = stream.read(&mut sink) {
                if read == 0 {
                    break;
                }
            }
        });

        let transport = HttpTransport::with_timeouts(
            Duration::from_secs(5),
            Duration::from_millis(200),
            Duration::from_millis(200),
        );
        let started = Instant::now();
        let error = transport
            .execute(HttpRequest {
                method: "GET".to_string(),
                url: format!("http://{addr}/stall"),
                headers: BTreeMap::new(),
                body: None,
            })
            .expect_err("stalled request must fail");
        assert!(started.elapsed() < Duration::from_secs(1), "{error:?}");
        handle.join().expect("server");
    }
}
