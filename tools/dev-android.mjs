#!/usr/bin/env node
/**
 * Start the Namecheck dev stack and expose it to a USB-connected Android phone.
 *
 * `adb reverse` maps a port on the phone back to this machine's loopback, so the
 * phone can open http://127.0.0.1:<port> without Wi-Fi, LAN addresses, or
 * --host.
 *
 * If a Namecheck dev server is already running, attach to it instead of starting
 * a second one. Otherwise pick the first free port from 5173 so the tunnel and
 * Vite always agree on the same port. The Rust server on :8080 is reused when
 * it is already listening.
 */
import { spawn, spawnSync } from "node:child_process";
import { existsSync } from "node:fs";
import net from "node:net";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const ROOT = join(dirname(fileURLToPath(import.meta.url)), "..");
const WEB = join(ROOT, "web");
const FIRST_PORT = 5173;
const LAST_PORT = FIRST_PORT + 49;
const API_PORT = 8080;
const VITE_BIN = [
	join(ROOT, "node_modules", "vite", "bin", "vite.js"),
	join(WEB, "node_modules", "vite", "bin", "vite.js"),
].find((path) => existsSync(path));

function info(message) {
	console.log(`[dev:android] ${message}`);
}

function fail(message) {
	console.error(`[dev:android] ${message}`);
	process.exit(1);
}

function adb(args) {
	try {
		const proc = spawnSync("adb", args, { encoding: "utf8" });
		return {
			ok: proc.status === 0,
			out: (proc.stdout ?? "").trim(),
			err:
				proc.error?.code === "ENOENT"
					? "adb not found"
					: (proc.stderr ?? "").trim() ||
						(proc.error ? String(proc.error.message) : ""),
		};
	} catch (error) {
		return { ok: false, out: "", err: String(error?.message ?? error) };
	}
}

function pickDevice() {
	const listed = adb(["devices"]);
	if (!listed.ok && listed.err === "adb not found") {
		fail("adb not found. Install Android platform-tools and put adb on PATH.");
	}

	const devices = listed.out
		.split("\n")
		.map((line) => line.trim().split(/\s+/))
		.filter(([, state]) => state === "device")
		.map(([serial]) => serial);

	if (devices.length === 0) {
		if (listed.out) console.error(listed.out);
		fail(
			"No authorized device. Connect over USB, enable USB debugging, and accept the RSA prompt on the phone.",
		);
	}

	const requested = process.env.ANDROID_SERIAL;
	if (requested) {
		if (!devices.includes(requested)) {
			fail(
				`ANDROID_SERIAL=${requested} is not connected (found: ${devices.join(", ")}).`,
			);
		}
		return requested;
	}

	if (devices.length > 1) {
		fail(
			`Multiple devices connected: ${devices.join(", ")}. Set ANDROID_SERIAL to pick one.`,
		);
	}

	return devices[0];
}

function canConnect(host, port) {
	return new Promise((resolve) => {
		const socket = net.connect({ host, port });
		let settled = false;
		const finish = (value) => {
			if (settled) return;
			settled = true;
			socket.destroy();
			resolve(value);
		};

		socket.setTimeout(250);
		socket.once("connect", () => finish(true));
		socket.once("timeout", () => finish(false));
		socket.once("error", () => finish(false));
	});
}

async function isBusy(port) {
	const [ipv4, ipv6] = await Promise.all([
		canConnect("127.0.0.1", port),
		canConnect("::1", port),
	]);
	return ipv4 || ipv6;
}

async function devServerHtml(host, port) {
	const hostPart = host.includes(":") ? `[${host}]` : host;
	try {
		const response = await fetch(`http://${hostPart}:${port}/`, {
			signal: AbortSignal.timeout(4000),
		});
		if (!response.ok) return null;
		return await response.text();
	} catch {
		// Not reachable on this address, or no response in time.
	}
	return null;
}

/** Dev servers are only trusted if they actually serve this site in dev mode. */
async function isNamecheckDevServer(port) {
	for (const host of ["127.0.0.1", "::1"]) {
		const html = await devServerHtml(host, port);
		if (html === null) continue;
		if (/<title>[^<]*Namecheck/i.test(html) && html.includes("@vite/client"))
			return true;
	}
	return false;
}

async function resolveTarget() {
	const configured = process.env.PORT;
	if (configured) {
		const port = Number(configured);
		if (!Number.isInteger(port) || port < 1 || port > 65535) {
			fail(`PORT=${configured} is not a valid port.`);
		}
		if (!(await isBusy(port))) return { mode: "start", port };
		if (await isNamecheckDevServer(port)) return { mode: "attach", port };
		fail(
			`Port ${port} is used by another app. Stop it or choose a different PORT.`,
		);
	}

	for (let port = FIRST_PORT; port <= LAST_PORT; port++) {
		if (!(await isBusy(port))) return { mode: "start", port };
		if (await isNamecheckDevServer(port)) return { mode: "attach", port };
	}

	fail(
		`No free port between ${FIRST_PORT} and ${LAST_PORT}. Free one up or set PORT.`,
	);
}

async function resolveApi() {
	// Vite proxies /v1 to 127.0.0.1:8080, so only the IPv4 listener matters.
	if (await canConnect("127.0.0.1", API_PORT)) return "attach";
	return "start";
}

const serial = pickDevice();
if (!VITE_BIN) fail("Vite is not installed. Run npm install first.");
const target = await resolveTarget();
const api = await resolveApi();
let activePort;

const children = new Set();
let shuttingDown = false;

function removeReverse() {
	if (activePort === undefined) return;
	adb(["-s", serial, "reverse", "--remove", `tcp:${activePort}`]);
	activePort = undefined;
}

function killChild(child, signal) {
	if (child.exitCode !== null || child.signalCode !== null) return;
	if (process.platform === "win32") {
		spawnSync("taskkill", ["/pid", String(child.pid), "/T", "/F"], {
			stdio: "ignore",
		});
		return;
	}
	// The child is its own process group, so this also stops npm's Vite child.
	try {
		process.kill(-child.pid, signal);
	} catch {
		child.kill(signal);
	}
}

function shutdown(signal, code) {
	if (shuttingDown) return;
	shuttingDown = true;
	removeReverse();
	for (const child of children) killChild(child, signal);
	process.exit(code);
}

function startChild(name, command, args, cwd = ROOT) {
	const child = spawn(command, args, {
		cwd,
		stdio: "inherit",
		env: process.env,
		detached: process.platform !== "win32",
	});
	children.add(child);

	child.on("error", (error) => {
		console.error(`[dev:android] ${name} failed to start: ${error.message}`);
		shutdown("SIGTERM", 1);
	});
	child.on("exit", (code, signal) => {
		children.delete(child);
		if (shuttingDown) return;
		console.error(`[dev:android] ${name} exited (${signal ?? code}).`);
		shutdown("SIGTERM", code ?? 1);
	});

	return child;
}

process.on("SIGINT", () => shutdown("SIGINT", 130));
process.on("SIGTERM", () => shutdown("SIGTERM", 143));
process.on("exit", removeReverse);

adb(["-s", serial, "reverse", "--remove", `tcp:${target.port}`]);
const reverse = adb([
	"-s",
	serial,
	"reverse",
	`tcp:${target.port}`,
	`tcp:${target.port}`,
]);
if (!reverse.ok) fail(`adb reverse failed: ${reverse.err || reverse.out}`);
activePort = target.port;

info(`Device ${serial} ready`);
if (target.mode === "attach") {
	info(
		`Attaching to the Namecheck dev server already running on port ${target.port}`,
	);
} else {
	info(`Starting the Namecheck dev server on port ${target.port}`);
}
if (api === "attach") {
	info(`Reusing the Namecheck API already listening on 127.0.0.1:${API_PORT}`);
}

if (api === "start") {
	startChild("serve", "cargo", [
		"run",
		"--manifest-path",
		"rust/Cargo.toml",
		"-p",
		"namecheck",
		"--",
		"serve",
	]);
}

if (target.mode === "start") {
	// --strictPort keeps the URL predictable: if the port was taken in the
	// meantime, fail instead of silently moving and breaking the mapping. Run
	// the Vite binary directly so a project-specific `dev` script cannot pin a
	// different port.
	startChild(
		"web",
		process.execPath,
		[
			VITE_BIN,
			"--port",
			String(target.port),
			"--host",
			"127.0.0.1",
			"--strictPort",
		],
		WEB,
	);
}

info(`Open http://127.0.0.1:${target.port} on the phone (Ctrl-C to stop)`);
info(`Change the web port with PORT=5180 npm run dev:android`);

// Keep this process alive to own the reverse mapping. Child processes would
// keep the event loop alive on their own; in attach mode nothing else would.
setInterval(() => {}, 1 << 30);
