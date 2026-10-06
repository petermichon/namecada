import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
	bitbucket,
	bluesky,
	checkAll,
	codeberg,
	crates,
	devto,
	farcaster,
	firefox,
	github,
	gitlab,
	hex,
	homebrew,
	jsr,
	mastodon,
	npm,
	nuget,
	providers,
	pub,
	pypi,
	streamAll,
	toLabel,
	wikipedia,
	wordpress
} from './providers';
import { BOOTSTRAP_URL } from './bootstrap';
import { setCached } from './providerCache';

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

describe('pypi', () => {
	afterEach(() => vi.unstubAllGlobals());

	it('maps 200 to taken and 404 to available', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn((input: unknown) =>
				Promise.resolve(new Response(null, { status: String(input).includes('taken') ? 200 : 404 }))
			)
		);
		expect(await pypi.check('taken')).toEqual({ status: 'taken' });
		expect(await pypi.check('free')).toEqual({ status: 'available' });
	});
});

describe('crates', () => {
	afterEach(() => vi.unstubAllGlobals());

	it('maps 200 to taken and 404 to available', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn((input: unknown) =>
				Promise.resolve(new Response(null, { status: String(input).includes('taken') ? 200 : 404 }))
			)
		);
		expect(await crates.check('taken')).toEqual({ status: 'taken' });
		expect(await crates.check('free')).toEqual({ status: 'available' });
	});
});

describe('nuget', () => {
	afterEach(() => vi.unstubAllGlobals());

	it('maps 200 to taken and 404 to available', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn((input: unknown) =>
				Promise.resolve(new Response(null, { status: String(input).includes('taken') ? 200 : 404 }))
			)
		);
		expect(await nuget.check('taken')).toEqual({ status: 'taken' });
		expect(await nuget.check('free')).toEqual({ status: 'available' });
	});
});

describe('bluesky', () => {
	afterEach(() => vi.unstubAllGlobals());

	it('maps 200 to taken and 400 to available', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn((input: unknown) =>
				Promise.resolve(new Response(null, { status: String(input).includes('taken') ? 200 : 400 }))
			)
		);
		expect(await bluesky.check('taken')).toEqual({ status: 'taken' });
		expect(await bluesky.check('free')).toEqual({ status: 'available' });
	});
});

describe('mastodon', () => {
	afterEach(() => vi.unstubAllGlobals());

	it('maps 200 to taken and 404 to available', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn((input: unknown) =>
				Promise.resolve(new Response(null, { status: String(input).includes('taken') ? 200 : 404 }))
			)
		);
		expect(await mastodon.check('taken')).toEqual({ status: 'taken' });
		expect(await mastodon.check('free')).toEqual({ status: 'available' });
	});
});

describe('hex', () => {
	afterEach(() => vi.unstubAllGlobals());

	it('maps 200 to taken and 404 to available', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn((input: unknown) =>
				Promise.resolve(new Response(null, { status: String(input).includes('taken') ? 200 : 404 }))
			)
		);
		expect(await hex.check('taken')).toEqual({ status: 'taken' });
		expect(await hex.check('free')).toEqual({ status: 'available' });
	});
});

describe('pub.dev', () => {
	afterEach(() => vi.unstubAllGlobals());

	it('maps 200 to taken and 404 to available', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn((input: unknown) =>
				Promise.resolve(new Response(null, { status: String(input).includes('taken') ? 200 : 404 }))
			)
		);
		expect(await pub.check('taken')).toEqual({ status: 'taken' });
		expect(await pub.check('free')).toEqual({ status: 'available' });
	});
});

describe('homebrew', () => {
	afterEach(() => vi.unstubAllGlobals());

	it('maps 200 to taken and 404 to available', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn((input: unknown) =>
				Promise.resolve(new Response(null, { status: String(input).includes('taken') ? 200 : 404 }))
			)
		);
		expect(await homebrew.check('taken')).toEqual({ status: 'taken' });
		expect(await homebrew.check('free')).toEqual({ status: 'available' });
	});
});

describe('firefox', () => {
	afterEach(() => vi.unstubAllGlobals());

	it('maps 200 to taken and 404 to available', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn((input: unknown) =>
				Promise.resolve(new Response(null, { status: String(input).includes('taken') ? 200 : 404 }))
			)
		);
		expect(await firefox.check('taken')).toEqual({ status: 'taken' });
		expect(await firefox.check('free')).toEqual({ status: 'available' });
	});

	it('treats a 401 (existing but unlisted) as taken', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn(() => Promise.resolve(new Response(null, { status: 401 })))
		);
		expect(await firefox.check('example')).toEqual({ status: 'taken' });
	});
});

describe('gitlab', () => {
	afterEach(() => vi.unstubAllGlobals());

	it('is taken when the users array is non-empty and available when empty', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn((input: unknown) =>
				Promise.resolve(
					new Response(JSON.stringify(String(input).includes('taken') ? [{ id: 1 }] : []), {
						status: 200
					})
				)
			)
		);
		expect(await gitlab.check('taken')).toEqual({ status: 'taken' });
		expect(await gitlab.check('free')).toEqual({ status: 'available' });
	});

	it('is unknown on a non-200 response', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn(() => Promise.resolve(new Response(null, { status: 500 })))
		);
		expect(await gitlab.check('acema')).toEqual({ status: 'unknown' });
	});
});

describe('jsr', () => {
	afterEach(() => vi.unstubAllGlobals());

	it('maps 200 to taken and 404 to available', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn((input: unknown) =>
				Promise.resolve(new Response(null, { status: String(input).includes('taken') ? 200 : 404 }))
			)
		);
		expect(await jsr.check('taken')).toEqual({ status: 'taken' });
		expect(await jsr.check('free')).toEqual({ status: 'available' });
	});
});

describe('wordpress', () => {
	afterEach(() => vi.unstubAllGlobals());

	it('maps 200 to taken and 404 to available', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn((input: unknown) =>
				Promise.resolve(new Response(null, { status: String(input).includes('taken') ? 200 : 404 }))
			)
		);
		expect(await wordpress.check('taken')).toEqual({ status: 'taken' });
		expect(await wordpress.check('free')).toEqual({ status: 'available' });
	});

	it('treats a closed plugin (404 with closed:true) as taken', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn(() => Promise.resolve(new Response(JSON.stringify({ closed: true }), { status: 404 })))
		);
		expect(await wordpress.check('google')).toEqual({ status: 'taken' });
	});
});

describe('bitbucket', () => {
	afterEach(() => vi.unstubAllGlobals());

	it('maps 200 to taken and 404 to available', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn((input: unknown) =>
				Promise.resolve(new Response(null, { status: String(input).includes('taken') ? 200 : 404 }))
			)
		);
		expect(await bitbucket.check('taken')).toEqual({ status: 'taken' });
		expect(await bitbucket.check('free')).toEqual({ status: 'available' });
	});
});

describe('codeberg', () => {
	afterEach(() => vi.unstubAllGlobals());

	it('maps 200 to taken and 404 to available', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn((input: unknown) =>
				Promise.resolve(new Response(null, { status: String(input).includes('taken') ? 200 : 404 }))
			)
		);
		expect(await codeberg.check('taken')).toEqual({ status: 'taken' });
		expect(await codeberg.check('free')).toEqual({ status: 'available' });
	});
});

describe('wikipedia', () => {
	afterEach(() => vi.unstubAllGlobals());

	it('is available when the user is marked missing', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn((input: unknown) =>
				Promise.resolve(
					new Response(
						JSON.stringify(
							String(input).includes('taken')
								? { query: { users: [{ userid: 1, name: 'x' }] } }
								: { query: { users: [{ name: 'x', missing: '' }] } }
						),
						{ status: 200 }
					)
				)
			)
		);
		expect(await wikipedia.check('taken')).toEqual({ status: 'taken' });
		expect(await wikipedia.check('free')).toEqual({ status: 'available' });
	});

	it('is unknown on a non-200 response', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn(() => Promise.resolve(new Response(null, { status: 500 })))
		);
		expect(await wikipedia.check('acema')).toEqual({ status: 'unknown' });
	});
});

describe('devto', () => {
	afterEach(() => vi.unstubAllGlobals());

	it('maps 200 to taken and 404 to available', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn((input: unknown) =>
				Promise.resolve(new Response(null, { status: String(input).includes('taken') ? 200 : 404 }))
			)
		);
		expect(await devto.check('taken')).toEqual({ status: 'taken' });
		expect(await devto.check('free')).toEqual({ status: 'available' });
	});
});

describe('farcaster', () => {
	afterEach(() => vi.unstubAllGlobals());

	it('maps 200 to taken and 404 to available', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn((input: unknown) =>
				Promise.resolve(new Response(null, { status: String(input).includes('taken') ? 200 : 404 }))
			)
		);
		expect(await farcaster.check('taken')).toEqual({ status: 'taken' });
		expect(await farcaster.check('free')).toEqual({ status: 'available' });
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
				if (url.includes('gitlab.com')) {
					return Promise.resolve(new Response(JSON.stringify([]), { status: 200 }));
				}
				if (url.includes('wikipedia.org')) {
					return Promise.resolve(
						new Response(JSON.stringify({ query: { users: [{ name: 'x', missing: '' }] } }), {
							status: 200
						})
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

describe('runProvider caching', () => {
	beforeEach(() => {
		const map = new Map<string, string>();
		vi.stubGlobal('localStorage', {
			getItem: (key: string) => map.get(key) ?? null,
			setItem: (key: string, value: string) => void map.set(key, value),
			removeItem: (key: string) => void map.delete(key),
			clear: () => map.clear(),
			key: (index: number) => [...map.keys()][index] ?? null,
			get length() {
				return map.size;
			}
		});
	});
	afterEach(() => vi.unstubAllGlobals());

	it('serves a fresh cached result without hitting the network', async () => {
		setCached('npm', 'acema', 'taken');
		const calls: string[] = [];
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
		expect(rows.find((row) => row.id === 'npm')?.status).toBe('taken');
		expect(calls.some((url) => url.includes('registry.npmjs.org'))).toBe(false);
	});
});
