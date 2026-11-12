import { afterEach, describe, expect, it, vi } from 'vitest';
import { CHECK_ENDPOINT, createNdjsonReader, parseEvent, streamFromApi } from './api';
import type { CheckEvent } from './api';

function stubApi(body: string) {
	return vi.fn((input: unknown) => {
		if (String(input).includes(`${CHECK_ENDPOINT}?name=`)) {
			return Promise.resolve(new Response(body, { status: 200 }));
		}
		return Promise.resolve(new Response(null, { status: 404 }));
	});
}

function result(overrides: Record<string, unknown> = {}): string {
	return JSON.stringify({
		type: 'result',
		id: 'lemmy',
		platform: 'lemmy',
		platformName: 'Lemmy',
		name: 'acema',
		url: 'https://lemmy.world/u/acema',
		status: 'taken',
		...overrides
	});
}

const manifest = JSON.stringify({
	type: 'manifest',
	version: 1,
	label: 'acema',
	sources: [
		{
			id: 'npm',
			platformName: 'npm',
			url: 'https://npmjs.com/acema',
			paid: false,
			checked: true
		},
		{
			id: 'tiktok',
			platformName: 'TikTok',
			url: 'https://tiktok.com/@acema',
			paid: false,
			checked: false
		}
	]
});

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('parseEvent', () => {
	it('parses the manifest', () => {
		const event = parseEvent(manifest);
		expect(event).toMatchObject({ type: 'manifest', label: 'acema' });
		expect(event?.type === 'manifest' && event.sources).toHaveLength(2);
		expect(event?.type === 'manifest' && event.sources[1]).toMatchObject({
			id: 'tiktok',
			platformName: 'TikTok',
			checked: false,
			paid: false
		});
	});

	it('defaults missing manifest fields and drops malformed sources', () => {
		const event = parseEvent(
			JSON.stringify({
				type: 'manifest',
				sources: [{ id: 'x' }, { id: 'npm', url: 'u' }]
			})
		);
		expect(event).toMatchObject({ type: 'manifest', label: '' });
		expect(event?.type === 'manifest' && event.sources).toEqual([
			{ id: 'npm', platformName: 'npm', url: 'u', paid: false, checked: false }
		]);
	});

	it('parses a result event', () => {
		expect(parseEvent(result())).toMatchObject({
			type: 'result',
			id: 'lemmy',
			platformName: 'Lemmy',
			status: 'taken'
		});
	});

	it('drops an empty detail and defaults paid', () => {
		const event = parseEvent(result({ detail: '' }));
		expect(event).toMatchObject({ detail: undefined, paid: false });
	});

	it('parses the done event', () => {
		expect(parseEvent(JSON.stringify({ type: 'done' }))).toEqual({ type: 'done' });
	});

	it('rejects blank, malformed, unknown, and incomplete lines', () => {
		expect(parseEvent('')).toBeNull();
		expect(parseEvent('  ')).toBeNull();
		expect(parseEvent('not json')).toBeNull();
		expect(parseEvent(JSON.stringify({ type: 'nope' }))).toBeNull();
		expect(parseEvent(JSON.stringify({ type: 'result', id: 'x' }))).toBeNull();
		expect(parseEvent(result({ status: 'weird' }))).toBeNull();
	});
});

describe('createNdjsonReader', () => {
	it('parses complete lines and flushes a trailing partial one', () => {
		const seen: CheckEvent[] = [];
		const reader = createNdjsonReader((event) => seen.push(event));
		reader.push(manifest + '\n\n');
		reader.push('{"type":"result","id":"b"');
		reader.push(',"name":"n","url":"u","status":"unknown"}');
		reader.flush();
		expect(seen.map((event) => event.type)).toEqual(['manifest', 'result']);
	});
});

describe('streamFromApi', () => {
	it('streams the manifest, results, and done in order', async () => {
		const body =
			[
				manifest,
				result({ id: 'lemmy', status: 'taken' }),
				result({ id: 'peertube', platformName: 'PeerTube', status: 'available' }),
				JSON.stringify({
					type: 'done',
					counts: { available: 1, taken: 1, unknown: 0, unsupported: 1 }
				})
			].join('\n') + '\n';
		vi.stubGlobal('fetch', stubApi(body));

		const seen: CheckEvent[] = [];
		await streamFromApi('acema', (event) => seen.push(event));

		expect(seen.map((event) => event.type)).toEqual(['manifest', 'result', 'result', 'done']);
		expect(seen[1].type === 'result' && seen[1].platformName).toBe('Lemmy');
	});

	it('requests the versioned endpoint', async () => {
		const fetchMock = stubApi(manifest + '\n');
		vi.stubGlobal('fetch', fetchMock);
		await streamFromApi('acema', () => {});
		expect(fetchMock).toHaveBeenCalledWith(
			`${CHECK_ENDPOINT}?name=acema`,
			expect.objectContaining({ credentials: 'omit' })
		);
	});

	it('rejects on an error response', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn(() => Promise.resolve(new Response(null, { status: 500 })))
		);
		await expect(streamFromApi('acema', () => {})).rejects.toThrow('500');
	});

	it('rejects when the backend is unreachable', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn(() => Promise.reject(new TypeError('Failed to fetch')))
		);
		await expect(streamFromApi('acema', () => {})).rejects.toThrow('Failed to fetch');
	});

	it('propagates an abort', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn(() => Promise.reject(new DOMException('aborted', 'AbortError')))
		);
		await expect(streamFromApi('acema', () => {})).rejects.toBeInstanceOf(DOMException);
	});
});
