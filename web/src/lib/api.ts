import type { ProviderResult, Status } from './platforms';

// The backend `serve` command defaults to :8080 (see backend/internal/api).
export const DEFAULT_API_BASE = 'http://localhost:8080';

const STATUSES: Status[] = ['available', 'taken', 'unknown'];

// apiBase resolves the backend origin. VITE_API_BASE overrides the default so a
// deployment can point the app at a remote `serve`.
export function apiBase(): string {
	const base = import.meta.env.VITE_API_BASE?.trim();
	return (base || DEFAULT_API_BASE).replace(/\/+$/, '');
}

export function isAbort(error: unknown): boolean {
	return error instanceof DOMException && error.name === 'AbortError';
}

// parseResult turns one NDJSON line into a ProviderResult, or null when the
// line is blank or not a well-formed result (a malformed line is skipped rather
// than failing the whole stream).
export function parseResult(line: string): ProviderResult | null {
	const trimmed = line.trim();
	if (!trimmed) return null;
	let data: unknown;
	try {
		data = JSON.parse(trimmed);
	} catch {
		return null;
	}
	if (typeof data !== 'object' || data === null) return null;
	const row = data as Record<string, unknown>;
	if (typeof row.id !== 'string' || !row.id) return null;
	if (typeof row.name !== 'string' || typeof row.url !== 'string') return null;
	if (typeof row.status !== 'string' || !STATUSES.includes(row.status as Status)) return null;
	const platform = typeof row.platform === 'string' ? row.platform : '';
	return {
		id: row.id,
		platform,
		platformName:
			typeof row.platformName === 'string' && row.platformName ? row.platformName : platform,
		name: row.name,
		url: row.url,
		status: row.status as Status,
		detail: typeof row.detail === 'string' && row.detail ? row.detail : undefined
	};
}

// createNdjsonReader is an incremental parser: push decoded chunks as they
// arrive and flush the trailing line at the end of the body.
export function createNdjsonReader(onResult: (result: ProviderResult) => void) {
	let buffer = '';
	return {
		push(text: string) {
			buffer += text;
			let index: number;
			while ((index = buffer.indexOf('\n')) !== -1) {
				const line = buffer.slice(0, index);
				buffer = buffer.slice(index + 1);
				const result = parseResult(line);
				if (result) onResult(result);
			}
		},
		flush() {
			const result = parseResult(buffer);
			buffer = '';
			if (result) onResult(result);
		}
	};
}

async function consume(body: ReadableStream<Uint8Array>, onResult: (r: ProviderResult) => void) {
	const reader = body.getReader();
	const decoder = new TextDecoder();
	const parser = createNdjsonReader(onResult);
	try {
		for (;;) {
			const { done, value } = await reader.read();
			if (done) break;
			parser.push(decoder.decode(value, { stream: true }));
		}
		parser.push(decoder.decode());
		parser.flush();
	} catch (error) {
		reader.cancel().catch(() => {});
		throw error;
	}
}

async function openCheck(label: string, signal?: AbortSignal): Promise<Response> {
	return fetch(`${apiBase()}/check?name=${encodeURIComponent(label)}`, {
		signal,
		headers: { Accept: 'application/x-ndjson' },
		referrerPolicy: 'no-referrer',
		credentials: 'omit'
	});
}

// streamFromApi consumes the backend NDJSON stream, emitting each result as it
// arrives. It rejects when the backend is unreachable or responds with an
// error. The backend is the sole source: there is no browser-direct fallback.
export async function streamFromApi(
	label: string,
	onResult: (result: ProviderResult) => void,
	signal?: AbortSignal
): Promise<void> {
	const res = await openCheck(label, signal);
	if (!res.ok) {
		res.body?.cancel().catch(() => {});
		throw new Error(`backend responded ${res.status}`);
	}
	if (!res.body) throw new Error('backend returned an empty response');
	await consume(res.body, onResult);
}
