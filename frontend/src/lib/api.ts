import { streamAll, type PlatformId, type ProviderResult, type Status } from './providers';

// The backend `serve` command defaults to :8080 (see backend/internal/api).
export const DEFAULT_API_BASE = 'http://localhost:8080';

export type Source = 'api' | 'direct';

const STATUSES: Status[] = ['available', 'taken', 'unknown'];

// apiBase resolves the backend origin. VITE_API_BASE overrides the default so a
// deployment can point the app at a remote `serve` without a rebuild of the
// fallback path.
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
	return {
		id: row.id,
		platform: row.platform as PlatformId,
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
// error, leaving the caller free to fall back to browser-direct checks.
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

// streamResults prefers the backend API and falls back to browser-direct checks
// when it is unreachable. `onSource` is called once the source is known, before
// any result, so the caller can seed its UI accordingly. An abort propagates
// instead of triggering the fallback.
export async function streamResults(
	label: string,
	onResult: (result: ProviderResult) => void,
	signal?: AbortSignal,
	onSource?: (source: Source) => void
): Promise<Source> {
	let res: Response;
	try {
		res = await openCheck(label, signal);
	} catch (error) {
		if (isAbort(error)) throw error;
		onSource?.('direct');
		await streamAll(label, onResult, signal);
		return 'direct';
	}

	if (!res.ok || !res.body) {
		res.body?.cancel().catch(() => {});
		onSource?.('direct');
		await streamAll(label, onResult, signal);
		return 'direct';
	}

	onSource?.('api');
	await consume(res.body, onResult);
	return 'api';
}
