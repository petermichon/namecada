import type { ProviderResult, Status } from './platforms';

// The check stream is NDJSON with three event types, in order:
//   manifest → result* → done
// The manifest is sent before any provider runs, so the client can render every
// row (checked or deferred) immediately and fill it in by id.
export const CHECK_ENDPOINT = '/v1/check';

const STATUSES: Status[] = ['available', 'taken', 'unknown'];

export interface ManifestSource {
	id: string;
	platformName: string;
	url: string;
	paid: boolean;
	checked: boolean;
	tags: string[];
}

export interface ManifestEvent {
	type: 'manifest';
	version: number;
	label: string;
	sources: ManifestSource[];
}

export interface ResultEvent extends ProviderResult {
	type: 'result';
}

export interface DoneEvent {
	type: 'done';
}

export type CheckEvent = ManifestEvent | ResultEvent | DoneEvent;

// CheckClient is the transport seam: the app shell (HTTP today, native IPC
// later) provides the events; the UI and the lookup state do not care which.
export interface CheckClient {
	stream(label: string, onEvent: (event: CheckEvent) => void, signal?: AbortSignal): Promise<void>;
}

export function isAbort(error: unknown): boolean {
	return error instanceof DOMException && error.name === 'AbortError';
}

function object(value: unknown): Record<string, unknown> | null {
	return typeof value === 'object' && value !== null ? (value as Record<string, unknown>) : null;
}

function parseManifest(row: Record<string, unknown>): ManifestEvent | null {
	if (!Array.isArray(row.sources)) return null;
	const sources: ManifestSource[] = [];
	for (const raw of row.sources) {
		const source = object(raw);
		if (!source || typeof source.id !== 'string' || !source.id) continue;
		if (typeof source.url !== 'string') continue;
		const platform = typeof source.platformName === 'string' ? source.platformName : source.id;
		const tags = Array.isArray(source.tags)
			? source.tags.filter((tag): tag is string => typeof tag === 'string' && tag.length > 0)
			: [];
		sources.push({
			id: source.id,
			platformName: platform || source.id,
			url: source.url,
			paid: source.paid === true,
			checked: source.checked === true,
			tags
		});
	}
	return {
		type: 'manifest',
		version: typeof row.version === 'number' ? row.version : 0,
		label: typeof row.label === 'string' ? row.label : '',
		sources
	};
}

function parseResultEvent(row: Record<string, unknown>): ResultEvent | null {
	if (typeof row.id !== 'string' || !row.id) return null;
	if (typeof row.name !== 'string' || typeof row.url !== 'string') return null;
	if (typeof row.status !== 'string' || !STATUSES.includes(row.status as Status)) return null;
	const platform = typeof row.platform === 'string' ? row.platform : '';
	return {
		type: 'result',
		id: row.id,
		platform,
		platformName:
			typeof row.platformName === 'string' && row.platformName ? row.platformName : platform,
		name: row.name,
		url: row.url,
		status: row.status as Status,
		detail: typeof row.detail === 'string' && row.detail ? row.detail : undefined,
		paid: row.paid === true
	};
}

function parseDone(row: Record<string, unknown>): DoneEvent | null {
	return row.type === 'done' ? { type: 'done' } : null;
}

// parseEvent turns one NDJSON line into a check event, or null when the line is
// blank or not a well-formed event (a malformed line is skipped rather than
// failing the whole stream).
export function parseEvent(line: string): CheckEvent | null {
	const trimmed = line.trim();
	if (!trimmed) return null;
	let data: unknown;
	try {
		data = JSON.parse(trimmed);
	} catch {
		return null;
	}
	const row = object(data);
	if (!row) return null;
	switch (row.type) {
		case 'manifest':
			return parseManifest(row);
		case 'result':
			return parseResultEvent(row);
		case 'done':
			return parseDone(row);
		default:
			return null;
	}
}

// createNdjsonReader is an incremental parser: push decoded chunks as they
// arrive and flush the trailing line at the end of the body.
export function createNdjsonReader(onEvent: (event: CheckEvent) => void) {
	let buffer = '';
	return {
		push(text: string) {
			buffer += text;
			let index: number;
			while ((index = buffer.indexOf('\n')) !== -1) {
				const line = buffer.slice(0, index);
				buffer = buffer.slice(index + 1);
				const event = parseEvent(line);
				if (event) onEvent(event);
			}
		},
		flush() {
			const event = parseEvent(buffer);
			buffer = '';
			if (event) onEvent(event);
		}
	};
}

async function consume(body: ReadableStream<Uint8Array>, onEvent: (e: CheckEvent) => void) {
	const reader = body.getReader();
	const decoder = new TextDecoder();
	const parser = createNdjsonReader(onEvent);
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

// streamFromApi consumes the backend event stream. It rejects when the backend
// is unreachable or responds with an error.
export async function streamFromApi(
	label: string,
	onEvent: (event: CheckEvent) => void,
	signal?: AbortSignal
): Promise<void> {
	const res = await fetch(`${CHECK_ENDPOINT}?name=${encodeURIComponent(label)}`, {
		signal,
		headers: { Accept: 'application/x-ndjson' },
		referrerPolicy: 'no-referrer',
		credentials: 'omit'
	});
	if (!res.ok) {
		res.body?.cancel().catch(() => {});
		throw new Error(`backend responded ${res.status}`);
	}
	if (!res.body) throw new Error('backend returned an empty response');
	await consume(res.body, onEvent);
}

export const httpCheckClient: CheckClient = { stream: streamFromApi };
