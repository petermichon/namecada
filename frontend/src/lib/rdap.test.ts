import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
	domainUrl,
	lookupDomain,
	normalizeDomain,
	parseDomain,
	registrationDate,
	statusFromHttp,
	tldOf,
	UnsupportedTldError
} from './rdap';
import { BOOTSTRAP_URL } from './bootstrap';

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

	it('reduces a subdomain to its registrable domain', () => {
		expect(normalizeDomain('abc.def.com')).toBe('def.com');
		expect(normalizeDomain('www.bbc.co.uk')).toBe('bbc.co.uk');
	});
});

describe('parseDomain', () => {
	it('splits the registrable domain from the subdomain', () => {
		expect(parseDomain('abc.google.com')).toEqual({ domain: 'google.com', subdomain: 'abc' });
		expect(parseDomain('google.com')).toEqual({ domain: 'google.com', subdomain: '' });
	});

	it('handles multi-label public suffixes', () => {
		expect(parseDomain('abc.google.co.uk')).toEqual({
			domain: 'google.co.uk',
			subdomain: 'abc'
		});
	});

	it('rejects a bare public suffix or missing TLD', () => {
		expect(parseDomain('co.uk')).toBeNull();
		expect(parseDomain('acema')).toBeNull();
		expect(parseDomain('a..com')).toBeNull();
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

describe('registrationDate', () => {
	it('finds the registration event date', () => {
		expect(
			registrationDate({
				events: [
					{ eventAction: 'expiration', eventDate: '2030-01-01T00:00:00Z' },
					{ eventAction: 'registration', eventDate: '1997-09-15T04:00:00Z' }
				]
			})
		).toBe('1997-09-15T04:00:00Z');
	});

	it('returns null when there is no registration event', () => {
		expect(registrationDate({ events: [] })).toBeNull();
		expect(registrationDate({})).toBeNull();
		expect(registrationDate(null)).toBeNull();
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
				if (url.endsWith('/domain/google.test')) {
					return Promise.resolve(
						new Response(
							JSON.stringify({
								events: [{ eventAction: 'registration', eventDate: '1997-09-15T04:00:00Z' }]
							}),
							{ status: 200 }
						)
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

	it('extracts the registration date and raw payload for a registered domain', async () => {
		const result = await lookupDomain('google.test');
		expect(result).toEqual({
			domain: 'google.test',
			status: 'registered',
			registeredOn: '1997-09-15T04:00:00Z',
			raw: { events: [{ eventAction: 'registration', eventDate: '1997-09-15T04:00:00Z' }] }
		});
	});

	it('throws for a TLD with no RDAP endpoint', async () => {
		await expect(lookupDomain('acema.nonexistent')).rejects.toBeInstanceOf(UnsupportedTldError);
	});
});
