import { afterEach, describe, expect, it, vi } from 'vitest';
import { apiBase, createNdjsonReader, DEFAULT_API_BASE, parseResult, streamFromApi } from './api';
import type { ProviderResult } from './providers';

function line(result: Partial<ProviderResult> & { id: string }): string {
	return JSON.stringify(result);
}

function stubApi(body: string, init?: ResponseInit) {
	return vi.fn((input: unknown) => {
		if (String(input).includes('/check?name=')) {
			return Promise.resolve(new Response(body, { status: 200, ...init }));
		}
		return Promise.resolve(new Response(null, { status: 404 }));
	});
}

afterEach(() => {
	vi.unstubAllGlobals();
	vi.unstubAllEnvs();
});

describe('apiBase', () => {
	it('defaults to the local backend', () => {
		vi.stubEnv('VITE_API_BASE', '');
		expect(apiBase()).toBe(DEFAULT_API_BASE);
	});

	it('honors VITE_API_BASE and strips trailing slashes', () => {
		vi.stubEnv('VITE_API_BASE', 'https://api.example.com/');
		expect(apiBase()).toBe('https://api.example.com');
	});
});

describe('parseResult', () => {
	it('parses a backend result line', () => {
		expect(
			parseResult(
				line({
					id: 'lemmy',
					platform: 'lemmy',
					name: 'acema',
					url: 'https://lemmy.world/u/acema',
					status: 'taken'
				})
			)
		).toMatchObject({ id: 'lemmy', status: 'taken' });
	});

	it('drops an empty detail', () => {
		expect(
			parseResult(line({ id: 'x', name: 'n', url: 'u', status: 'available', detail: '' }))
		).toMatchObject({ detail: undefined });
	});

	it('rejects blank, malformed, and incomplete lines', () => {
		expect(parseResult('')).toBeNull();
		expect(parseResult('  ')).toBeNull();
		expect(parseResult('not json')).toBeNull();
		expect(
			parseResult(line({ id: 'x', name: 'n', url: 'u', status: 'weird' as never }))
		).toBeNull();
		expect(parseResult(JSON.stringify({ id: 'x' }))).toBeNull();
	});
});

describe('createNdjsonReader', () => {
	it('parses complete lines and flushes a trailing partial one', () => {
		const seen: ProviderResult[] = [];
		const reader = createNdjsonReader((result) => seen.push(result));
		reader.push(line({ id: 'a', name: 'n', url: 'u', status: 'taken' }) + '\n\n');
		reader.push('{"id":"b"');
		reader.push(',"name":"n","url":"u","status":"unknown"}');
		reader.flush();
		expect(seen.map((result) => result.id)).toEqual(['a', 'b']);
	});
});

describe('streamFromApi', () => {
	it('streams one result per NDJSON line', async () => {
		const body =
			[
				line({
					id: 'lemmy',
					platform: 'lemmy',
					name: 'acema',
					url: 'https://lemmy.world/u/acema',
					status: 'taken'
				}),
				line({
					id: 'peertube',
					platform: 'peertube',
					name: 'acema',
					url: 'https://framatube.org/a/acema',
					status: 'available',
					detail: ''
				})
			].join('\n') + '\n';
		vi.stubGlobal('fetch', stubApi(body));

		const seen: ProviderResult[] = [];
		await streamFromApi('acema', (result) => seen.push(result));

		expect(seen.map((result) => result.id)).toEqual(['lemmy', 'peertube']);
		expect(seen[1].detail).toBeUndefined();
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
