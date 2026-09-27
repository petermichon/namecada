import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
	domainUrl,
	lookupDomain,
	normalizeDomain,
	statusFromHttp,
	tldOf,
	UnsupportedTldError
} from './rdap';
import { BOOTSTRAP_URL } from './bootstrap';
import { isFresh } from './cache';

describe('normalizeDomain', () => {
	it('lowercases and trims a full domain', () => {
		expect(normalizeDomain('  AceMa.CoM  ')).toBe('acema.com');
	});

	it('requires an explicit TLD, including multi-label domains', () => {
		expect(normalizeDomain('acema.com')).toBe('acema.com');
		expect(normalizeDomain('acema.io')).toBe('acema.io');
		expect(normalizeDomain('acema.co.uk')).toBe('acema.co.uk');
	});

	it('rejects a bare label with no TLD', () => {
		expect(normalizeDomain('acema')).toBeNull();
		expect(normalizeDomain('.acema')).toBeNull();
		expect(normalizeDomain('acema.')).toBeNull();
	});

	it('strips stray leading/trailing dots around a full domain', () => {
		expect(normalizeDomain('.acema.com')).toBe('acema.com');
		expect(normalizeDomain('acema.com.')).toBe('acema.com');
	});

	it('accepts digits and internal hyphens', () => {
		expect(normalizeDomain('a-1-b.com')).toBe('a-1-b.com');
	});

	it('rejects empty, spaced and malformed input', () => {
		expect(normalizeDomain('')).toBeNull();
		expect(normalizeDomain('   ')).toBeNull();
		expect(normalizeDomain('a b.com')).toBeNull();
		expect(normalizeDomain('-abc.com')).toBeNull();
		expect(normalizeDomain('abc-.com')).toBeNull();
		expect(normalizeDomain('a..com')).toBeNull();
		expect(normalizeDomain(`${'a'.repeat(64)}.com`)).toBeNull();
	});

	it('accepts a syntactically valid but unknown TLD for later resolution', () => {
		expect(normalizeDomain('something.wrong')).toBe('something.wrong');
	});
});

describe('tldOf', () => {
	it('returns the last label', () => {
		expect(tldOf('acema.com')).toBe('com');
		expect(tldOf('acema.co.uk')).toBe('uk');
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

describe('domainUrl', () => {
	it('appends the RDAP domain path to the registry base', () => {
		expect(domainUrl('https://rdap.verisign.com/com/v1/', 'acema.com')).toBe(
			'https://rdap.verisign.com/com/v1/domain/acema.com'
		);
	});
});

describe('lookupDomain', () => {
	beforeEach(() => {
		vi.stubGlobal(
			'fetch',
			vi.fn((input: unknown) => {
				const url = String(input);
				if (url === BOOTSTRAP_URL) {
					return Promise.resolve(
						new Response(JSON.stringify({ services: [[['test'], ['https://rdap.test/']]] }), {
							status: 200
						})
					);
				}
				return Promise.resolve(new Response(null, { status: 404 }));
			})
		);
	});

	afterEach(() => vi.unstubAllGlobals());

	it('routes to the TLD registry and maps the response status', async () => {
		const result = await lookupDomain('acema.test');
		expect(result).toEqual({ domain: 'acema.test', status: 'unregistered' });
		expect(fetch).toHaveBeenCalledWith('https://rdap.test/domain/acema.test', expect.anything());
	});

	it('throws for a TLD with no RDAP endpoint', async () => {
		await expect(lookupDomain('acema.nonexistent')).rejects.toBeInstanceOf(UnsupportedTldError);
	});
});

describe('isFresh', () => {
	it('expires entries past the ttl', () => {
		const entry = { status: 'registered' as const, at: 1_000 };
		expect(isFresh(entry, 1_500, 1_000)).toBe(true);
		expect(isFresh(entry, 2_500, 1_000)).toBe(false);
	});
});
