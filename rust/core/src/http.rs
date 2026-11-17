//! Production HTTP transport. Redirects are left to the providers so their
//! final-URL handling (Giphy, etc.) stays explicit.

use std::collections::BTreeMap;
use std::io;
use std::time::Duration;

use ureq::config::IpFamily;

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
///
/// It connects over IPv4 first and only falls back to every address family
/// when that cannot connect. The resolver hands ureq the AAAA records first on
/// most hosts, and ureq tries the addresses one at a time, sharing the connect
/// budget in address order. Where a VPN or ISP blackholes IPv6 — the packets
/// vanish instead of being refused — the AAAA attempts spend the whole budget
/// and a perfectly reachable host reports `unknown`. Browsers hide that with
/// Happy Eyeballs; ureq has no such racing, so preferring A records is the fix.
/// A host with no A record still works, through the fallback.
pub struct HttpTransport {
    /// Address-family order: IPv4 first, the common case.
    agent: ureq::Agent,
    /// Same bounds, every address family, for IPv6-only or IPv4-blocked hosts.
    fallback: ureq::Agent,
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
        Self {
            agent: agent(global, connect, read, IpFamily::Ipv4Only),
            fallback: agent(global, connect, read, IpFamily::Any),
        }
    }

    fn send(
        &self,
        agent: &ureq::Agent,
        request: &HttpRequest,
    ) -> Result<HttpResponse, ureq::Error> {
        let result = match request.method.as_str() {
            "GET" => {
                let mut builder = agent.get(&request.url);
                for (key, value) in &request.headers {
                    builder = builder.header(key, value);
                }
                builder.call()
            }
            "POST" => {
                let mut builder = agent.post(&request.url);
                for (key, value) in &request.headers {
                    builder = builder.header(key, value);
                }
                match &request.body {
                    Some(body) => builder.send(body.as_slice()),
                    None => builder.send(""),
                }
            }
            // execute() rejects anything else before reaching here.
            other => unreachable!("method {other:?} was not checked"),
        };

        let mut response = result?;
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
            .map_err(|err| ureq::Error::Io(io::Error::other(err)))?;
        Ok(HttpResponse {
            status,
            headers,
            body,
        })
    }
}

/// agent builds one client. The family filter is the only difference between
/// the two an HttpTransport holds.
fn agent(global: Duration, connect: Duration, read: Duration, family: IpFamily) -> ureq::Agent {
    let config = ureq::Agent::config_builder()
        .timeout_global(Some(global))
        .timeout_resolve(Some(connect))
        .timeout_connect(Some(connect))
        .timeout_recv_response(Some(read))
        .timeout_recv_body(Some(read))
        .http_status_as_error(false)
        .ip_family(family)
        .max_redirects(0)
        .build();
    config.into()
}

/// family_may_be_why reports whether a failure could be caused by the address
/// family rather than by the host: nothing resolved, nothing connected, or the
/// connect phase ran out of time. A response, a TLS error or a protocol error
/// would fail the same way over either family, so they are not retried.
fn family_may_be_why(err: &ureq::Error) -> bool {
    match err {
        ureq::Error::HostNotFound | ureq::Error::ConnectionFailed => true,
        ureq::Error::Timeout(reason) => {
            matches!(reason, ureq::Timeout::Resolve | ureq::Timeout::Connect)
        }
        ureq::Error::Io(err) => matches!(
            err.kind(),
            io::ErrorKind::ConnectionRefused
                | io::ErrorKind::ConnectionReset
                | io::ErrorKind::ConnectionAborted
                | io::ErrorKind::HostUnreachable
                | io::ErrorKind::NetworkUnreachable
                | io::ErrorKind::AddrNotAvailable
                | io::ErrorKind::TimedOut
        ),
        _ => false,
    }
}

impl Default for HttpTransport {
    fn default() -> Self {
        Self::new()
    }
}

impl Transport for HttpTransport {
    fn execute(&self, request: HttpRequest) -> Result<HttpResponse, Error> {
        if !matches!(request.method.as_str(), "GET" | "POST") {
            return Err(Error::Invalid(format!(
                "unsupported method {:?}",
                request.method
            )));
        }
        match self.send(&self.agent, &request) {
            Ok(response) => Ok(response),
            Err(err) if family_may_be_why(&err) => self
                .send(&self.fallback, &request)
                .map_err(|err| Error::Transport(err.to_string())),
            Err(err) => Err(Error::Transport(err.to_string())),
        }
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

    /// The transport connects over IPv4 first, so an IPv6-only host is only
    /// reachable through the fallback. A loopback listener keeps that offline.
    #[test]
    fn an_ipv6_only_host_reaches_the_fallback() {
        let Ok(listener) = TcpListener::bind("[::1]:0") else {
            return; // no IPv6 loopback on this machine
        };
        let addr = listener.local_addr().expect("addr");
        let handle = thread::spawn(move || {
            let (mut stream, _) = listener.accept().expect("accept");
            let mut buf = [0_u8; 1024];
            let _ = stream.read(&mut buf);
            stream
                .write_all(b"HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nok")
                .expect("write");
        });

        let transport = HttpTransport::new();
        let response = transport
            .execute(HttpRequest {
                method: "GET".to_string(),
                url: format!("http://{addr}/"),
                headers: BTreeMap::new(),
                body: None,
            })
            .expect("fallback request");
        assert_eq!(response.status, 200);
        assert_eq!(response.body, b"ok");
        handle.join().expect("server");
    }

    #[test]
    fn rejects_an_unsupported_method() {
        let error = HttpTransport::new()
            .execute(HttpRequest {
                method: "DELETE".to_string(),
                url: "http://127.0.0.1:1/".to_string(),
                headers: BTreeMap::new(),
                body: None,
            })
            .expect_err("DELETE has no mapping");
        assert!(matches!(error, Error::Invalid(_)), "{error:?}");
    }
}
