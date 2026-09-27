import { describe, expect, it, vi } from 'vitest';
import { lookupDomain, normalizeDomain, rdapUrl, statusFromHttp } from './rdap';
import { isFresh } from './cache';

describe('normalizeDomain', () => {
	it('lowercases and trims', () => {
		expect(normalizeDomain('  AceMa  ')).toBe('acema.com');
	});

	it('strips a trailing .com, with or without a leading dot', () => {
		expect(normalizeDomain('acema.com')).toBe('acema.com');
		expect(normalizeDomain('.acema')).toBe('acema.com');
	});

	it('accepts digits and internal hyphens', () => {
		expect(normalizeDomain('a-1-b')).toBe('a-1-b.com');
	});

	it('rejects empty, spaced and malformed labels', () => {
		expect(normalizeDomain('')).toBeNull();
		expect(normalizeDomain('   ')).toBeNull();
		expect(normalizeDomain('a b')).toBeNull();
		expect(normalizeDomain('-abc')).toBeNull();
		expect(normalizeDomain('abc-')).toBeNull();
		expect(normalizeDomain('a'.repeat(64))).toBeNull();
	});
});

describe('statusFromHttp', () => {
	it('maps 200 and 404, anything else is unknown', () => {
		expect(statusFromHttp(200)).toBe('registered');
		expect(statusFromHttp(404)).toBe('unregistered');
		expect(statusFromHttp(429)).toBe('unknown');
		expect(statusFromHttp(500)).toBe('unknown');
	});
});

describe('rdapUrl', () => {
	it('points at the Verisign .com endpoint', () => {
		expect(rdapUrl('acema.com')).toBe('https://rdap.verisign.com/com/v1/domain/acema.com');
	});
});

describe('lookupDomain', () => {
	it('queries RDAP and maps the response status', async () => {
		const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 404 }));
		vi.stubGlobal('fetch', fetchMock);

		const result = await lookupDomain('acema.com');

		expect(fetchMock).toHaveBeenCalledWith(
			'https://rdap.verisign.com/com/v1/domain/acema.com',
			expect.anything()
		);
		expect(result).toEqual({ domain: 'acema.com', status: 'unregistered' });

		vi.unstubAllGlobals();
	});

	it('returns unknown on a server error', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(null, { status: 503 })));

		expect(await lookupDomain('acema.com')).toEqual({ domain: 'acema.com', status: 'unknown' });

		vi.unstubAllGlobals();
	});
});

describe('isFresh', () => {
	it('expires entries past the ttl', () => {
		const entry = { status: 'registered' as const, at: 1_000 };
		expect(isFresh(entry, 1_500, 1_000)).toBe(true);
		expect(isFresh(entry, 2_500, 1_000)).toBe(false);
	});
});
