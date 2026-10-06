import { afterEach, describe, expect, it, vi } from 'vitest';
import { checkAll, github, npm, providers, streamAll, toLabel } from './providers';
import { BOOTSTRAP_URL } from './bootstrap';

describe('toLabel', () => {
	it('accepts a bare label', () => {
		expect(toLabel('  AceMa ')).toBe('acema');
		expect(toLabel('a-1-b')).toBe('a-1-b');
	});

	it('extracts the label from a full domain, including subdomains', () => {
		expect(toLabel('acema.com')).toBe('acema');
		expect(toLabel('www.bbc.co.uk')).toBe('bbc');
	});

	it('rejects empty and malformed input', () => {
		expect(toLabel('')).toBeNull();
		expect(toLabel('   ')).toBeNull();
		expect(toLabel('a b')).toBeNull();
		expect(toLabel('-abc.com')).toBeNull();
	});
});

describe('github', () => {
	afterEach(() => vi.unstubAllGlobals());

	it('maps 200 to taken and 404 to available from a single users lookup', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn((input: unknown) =>
				Promise.resolve(
					new Response(null, { status: String(input).includes('/users/taken') ? 200 : 404 })
				)
			)
		);
		expect(await github.check('taken')).toEqual({ status: 'taken' });
		expect(await github.check('free')).toEqual({ status: 'available' });
	});
});

describe('npm', () => {
	afterEach(() => vi.unstubAllGlobals());

	it('maps 200 to taken and 404 to available', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn((input: unknown) =>
				Promise.resolve(new Response(null, { status: String(input).includes('taken') ? 200 : 404 }))
			)
		);
		expect(await npm.check('taken')).toEqual({ status: 'taken' });
		expect(await npm.check('free')).toEqual({ status: 'available' });
	});
});

describe('checkAll', () => {
	afterEach(() => vi.unstubAllGlobals());

	it('returns one row per provider', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn((input: unknown) => {
				const url = String(input);
				if (url === BOOTSTRAP_URL) {
					return Promise.resolve(
						new Response(
							JSON.stringify({
								services: [[['com', 'org', 'net', 'dev'], ['https://rdap.test/']]]
							}),
							{ status: 200 }
						)
					);
				}
				return Promise.resolve(new Response(null, { status: 404 }));
			})
		);
		const rows = await checkAll('acema');
		expect(rows).toHaveLength(providers.length);
		expect(rows.every((row) => row.status === 'available')).toBe(true);
		expect(rows.map((row) => row.name)).toContain('acema.com');
		expect(rows.map((row) => row.name)).toContain('acema');
	});

	it('keeps a failing provider as unknown instead of rejecting', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn((input: unknown) => {
				const url = String(input);
				if (url === BOOTSTRAP_URL) {
					return Promise.resolve(
						new Response(
							JSON.stringify({
								services: [[['com', 'org', 'net', 'dev'], ['https://rdap.test/']]]
							}),
							{ status: 200 }
						)
					);
				}
				if (url.includes('api.github.com')) return Promise.reject(new Error('boom'));
				return Promise.resolve(new Response(null, { status: 404 }));
			})
		);
		const rows = await checkAll('acema');
		const githubRow = rows.find((row) => row.id === 'github');
		expect(githubRow?.status).toBe('unknown');
		expect(githubRow?.detail).toBe('unreachable');
	});
});

describe('streamAll', () => {
	afterEach(() => vi.unstubAllGlobals());

	it('emits one result per provider as each settles', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn((input: unknown) => {
				const url = String(input);
				if (url === BOOTSTRAP_URL) {
					return Promise.resolve(
						new Response(
							JSON.stringify({
								services: [[['com', 'org', 'net', 'dev'], ['https://rdap.test/']]]
							}),
							{ status: 200 }
						)
					);
				}
				return Promise.resolve(new Response(null, { status: 404 }));
			})
		);
		const seen: string[] = [];
		await streamAll('acema', (result) => seen.push(result.id));
		expect(seen.sort()).toEqual(providers.map((provider) => provider.id).sort());
	});
});

describe('providers registry', () => {
	it('has a unique id per provider', () => {
		const ids = providers.map((provider) => provider.id);
		expect(new Set(ids).size).toBe(ids.length);
	});
});
