import type { ProviderResult, Status } from './rows';

// The check stream is NDJSON with three event types, in order:
//   manifest → result* → done
// The manifest arrives before any provider runs, so every row renders at once.
export const CHECK_ENDPOINT = '/v1/check';

const STATUSES: Status[] = ['available', 'taken', 'unknown'];

export interface ManifestSource {
	id: string;
	platformName: string;
	url: string;
	paid: boolean;
	checked: boolean;
}

export interface ManifestEvent {
	type: 'manifest';
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

function text(value: unknown): string {
	return typeof value === 'string' ? value : '';
}

function parseSource(raw: unknown): ManifestSource | null {
	if (typeof raw !== 'object' || raw === null) return null;
	const row = raw as Record<string, unknown>;
	const id = text(row.id);
	if (!id || typeof row.url !== 'string') return null;
	return {
		id,
		platformName: text(row.platformName) || id,
		url: row.url,
		paid: row.paid === true,
		checked: row.checked === true
	};
}

function parseResult(row: Record<string, unknown>): ResultEvent | null {
	const id = text(row.id);
	const platform = text(row.platform);
	if (!id || !text(row.name) || typeof row.url !== 'string') return null;
	if (!STATUSES.includes(row.status as Status)) return null;
	return {
		type: 'result',
		id,
		platform,
		platformName: text(row.platformName) || platform,
		name: text(row.name),
		url: row.url,
		status: row.status as Status,
		detail: text(row.detail) || undefined,
		paid: row.paid === true
	};
}

// parseEvent turns one NDJSON line into a check event, or null when the line is
// blank or not a well-formed event (a malformed line is skipped rather than
// failing the whole stream).
export function parseEvent(line: string): CheckEvent | null {
	const trimmed = line.trim();
	if (!trimmed) return null;

	let row: Record<string, unknown> | null = null;
	try {
		const value: unknown = JSON.parse(trimmed);
		if (typeof value === 'object' && value !== null) row = value as Record<string, unknown>;
	} catch {
		return null;
	}
	if (!row) return null;

	switch (row.type) {
		case 'manifest':
			if (!Array.isArray(row.sources)) return null;
			return {
				type: 'manifest',
				label: text(row.label),
				sources: row.sources.flatMap((source) => parseSource(source) ?? [])
			};
		case 'result':
			return parseResult(row);
		case 'done':
			return { type: 'done' };
		default:
			return null;
	}
}

// createNdjsonReader is an incremental parser: push decoded chunks as they
// arrive and flush the trailing line at the end of the body.
export function createNdjsonReader(onEvent: (event: CheckEvent) => void) {
	let buffer = '';
	const drain = (final: boolean) => {
		let index: number;
		while ((index = buffer.indexOf('\n')) !== -1) {
			const line = buffer.slice(0, index);
			buffer = buffer.slice(index + 1);
			const event = parseEvent(line);
			if (event) onEvent(event);
		}
		if (final) {
			const event = parseEvent(buffer);
			buffer = '';
			if (event) onEvent(event);
		}
	};
	return {
		push: (text: string) => {
			buffer += text;
			drain(false);
		},
		flush: () => drain(true)
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

// In the desktop shell the engine answers over IPC with the same events.
function inTauri(): boolean {
	return typeof window !== 'undefined' && '__TAURI_INTERNALS__' in window;
}

const tauriCheckClient: CheckClient = {
	async stream(label, onEvent, signal) {
		const { Channel, invoke } = await import('@tauri-apps/api/core');
		if (signal?.aborted) return;

		const channel = new Channel<CheckEvent>();
		channel.onmessage = (event) => {
			if (!signal?.aborted) onEvent(event);
		};
		await invoke('check', { label, onEvent: channel });
	}
};

// appCheckClient picks the transport once at load: the desktop app calls the
// engine over IPC, the browser build uses the NDJSON endpoint.
export const appCheckClient: CheckClient = inTauri() ? tauriCheckClient : { stream: streamFromApi };
