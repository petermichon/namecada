import { afterEach, describe, expect, it, vi } from 'vitest';
import { parseEvent, type CheckClient } from './api';
import { ALL_STATUSES, byPlatformName, DEFAULT_COST, DEFAULT_FILTER, type Row } from './rows';
import { Lookup } from './useLookup.svelte';
import type { Status } from './rows';

function row(
	platformName: string,
	status: Row['status'],
	name = 'acema',
	paid = false,
	url = `https://example.com/${platformName.toLowerCase()}/${name}`
): Row {
	return {
		id: `${platformName.toLowerCase()}-${name}`,
		platform: platformName.toLowerCase(),
		platformName,
		name,
		url,
		status,
		paid
	};
}

function manifest(
	sources: Array<{
		id: string;
		platformName: string;
		url: string;
		paid?: boolean;
		checked?: boolean;
	}>,
	label = 'acema'
): string {
	return JSON.stringify({
		type: 'manifest',
		version: 1,
		label,
		sources: sources.map((source) => ({
			paid: false,
			checked: true,
			...source
		}))
	});
}

function result(id: string, platformName: string, status: Status, overrides = {}): string {
	return JSON.stringify({
		type: 'result',
		id,
		platform: id,
		platformName,
		name: 'acema',
		url: `https://example.com/${id}/acema`,
		status,
		paid: false,
		...overrides
	});
}

function stubApi(body: string) {
	return vi.fn((input: unknown) => {
		if (String(input).includes('/v1/check?name=')) {
			return Promise.resolve(new Response(body, { status: 200 }));
		}
		return Promise.resolve(new Response(null, { status: 404 }));
	});
}

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('byPlatformName', () => {
	it('compares display names', () => {
		expect(byPlatformName(row('GitHub', 'taken'), row('Lemmy', 'taken'))).toBeLessThan(0);
		expect(byPlatformName(row('PeerTube', 'taken'), row('Lemmy', 'taken'))).toBeGreaterThan(0);
		expect(byPlatformName(row('Lemmy', 'taken'), row('Lemmy', 'available'))).toBe(0);
	});

	it('breaks display-name ties by URL, not arrival order', () => {
		const withUrl = (url: string): Row => ({ ...row('Domains', 'taken'), url });
		const tld = ['.org', '.ai', '.com', '.app', '.io', '.dev'].map((s) =>
			withUrl(`https://example${s}`)
		);
		expect(tld.toSorted(byPlatformName).map((r) => r.url)).toEqual([
			'https://example.ai',
			'https://example.app',
			'https://example.com',
			'https://example.dev',
			'https://example.io',
			'https://example.org'
		]);

		const hex = [
			{ ...row('Hex', 'taken'), url: 'https://hex.pm/users/acema' },
			{ ...row('Hex', 'taken'), url: 'https://hex.pm/packages/acema' }
		];
		expect(hex.toSorted(byPlatformName)[0].url).toBe('https://hex.pm/packages/acema');
	});
});

describe('visibleRows', () => {
	it('orders alphabetically by platformName regardless of status', () => {
		const lookup = new Lookup();
		lookup.rows = [row('PeerTube', 'pending'), row('GitHub', 'taken'), row('Lemmy', 'available')];
		expect(lookup.visibleRows.map((r) => r.platformName)).toEqual(['GitHub', 'Lemmy', 'PeerTube']);
	});

	it('keeps the order when filtered', () => {
		const lookup = new Lookup();
		lookup.rows = [row('PeerTube', 'taken'), row('GitHub', 'taken'), row('Lemmy', 'available')];
		lookup.filter = ['taken'];
		expect(lookup.visibleRows.map((r) => r.platformName)).toEqual(['GitHub', 'PeerTube']);
	});

	it('unions the selected statuses', () => {
		const lookup = new Lookup();
		lookup.rows = [row('PeerTube', 'unknown'), row('GitHub', 'taken'), row('Lemmy', 'available')];
		lookup.filter = ['available', 'unknown'];
		expect(lookup.visibleRows.map((r) => r.platformName)).toEqual(['Lemmy', 'PeerTube']);
	});

	it('shows nothing when no status is selected', () => {
		const lookup = new Lookup();
		lookup.rows = [row('GitHub', 'taken'), row('Lemmy', 'available'), row('TikTok', 'unsupported')];
		lookup.filter = [];
		expect(lookup.visibleRows).toEqual([]);
	});

	it('hides not-checked sources by default', () => {
		const lookup = new Lookup();
		lookup.rows = [row('GitHub', 'taken'), row('TikTok', 'unsupported')];
		expect(lookup.filter).toEqual(DEFAULT_FILTER);
		expect(lookup.visibleRows.map((r) => r.platformName)).toEqual(['GitHub']);
	});

	it('keeps pending rows visible while a verdict is selected', () => {
		const lookup = new Lookup();
		lookup.rows = [row('GitHub', 'pending'), row('Lemmy', 'taken')];
		expect(lookup.visibleRows.map((r) => r.platformName)).toEqual(['GitHub', 'Lemmy']);
	});

	it('hides pending rows when only not-checked is selected', () => {
		const lookup = new Lookup();
		lookup.rows = [row('GitHub', 'pending'), row('TikTok', 'unsupported')];
		lookup.filter = ['unsupported'];
		expect(lookup.visibleRows.map((r) => r.platformName)).toEqual(['TikTok']);
	});

	it('filters to not-checked sources', () => {
		const lookup = new Lookup();
		lookup.rows = [row('GitHub', 'taken'), row('TikTok', 'unsupported')];
		lookup.filter = ['unsupported'];
		expect(lookup.visibleRows.map((r) => r.platformName)).toEqual(['TikTok']);
	});

	it('breaks platformName ties by URL, not arrival order', () => {
		const lookup = new Lookup();
		lookup.rows = [row('Lemmy', 'taken', 'b'), row('Lemmy', 'available', 'a')];
		expect(lookup.visibleRows.map((r) => r.name)).toEqual(['a', 'b']);
	});

	it('keeps alphabetical order across paid and free rows', () => {
		const lookup = new Lookup();
		lookup.rows = [
			row('npm', 'taken'),
			row('Domains', 'taken', 'acema', true),
			row('GitHub', 'taken')
		];
		expect(lookup.visibleRows.map((r) => r.platformName)).toEqual(['Domains', 'GitHub', 'npm']);
	});

	it('filters by cost', () => {
		const lookup = new Lookup();
		lookup.rows = [
			row('Domains', 'taken', 'acema', true),
			row('GitHub', 'taken'),
			row('Heroku', 'taken', 'acema', true)
		];
		lookup.cost = ['paid'];
		expect(lookup.visibleRows.map((r) => r.platformName)).toEqual(['Domains', 'Heroku']);
	});

	it('selects both cost groups by default', () => {
		const lookup = new Lookup();
		lookup.rows = [row('Domains', 'taken', 'acema', true), row('GitHub', 'taken')];

		expect(lookup.cost).toEqual(DEFAULT_COST);
		expect(lookup.visibleRows.map((r) => r.platformName)).toEqual(['Domains', 'GitHub']);

		lookup.cost = [];
		expect(lookup.visibleRows).toEqual([]);
	});

	it('combines status and cost filters', () => {
		const lookup = new Lookup();
		lookup.rows = [
			row('GitHub', 'taken'),
			row('Lemmy', 'available'),
			row('npm', 'available'),
			row('Domains', 'taken', 'acema', true)
		];

		lookup.filter = ['available'];
		lookup.cost = ['free'];
		expect(lookup.visibleRows.map((r) => r.platformName)).toEqual(['Lemmy', 'npm']);

		lookup.filter = ['available', 'taken'];
		lookup.cost = ['paid'];
		expect(lookup.visibleRows.map((r) => r.platformName)).toEqual(['Domains']);
	});

	it('counts costs across every row', () => {
		const lookup = new Lookup();
		lookup.rows = [
			row('GitHub', 'taken'),
			row('npm', 'taken'),
			row('Domains', 'taken', 'acema', true)
		];

		expect(lookup.costCounts).toEqual({ free: 2, paid: 1 });
	});

	it('counts only the rows that are visible', () => {
		const lookup = new Lookup();
		lookup.rows = [
			row('GitHub', 'taken'),
			row('Lemmy', 'available'),
			row('npm', 'available'),
			row('Domains', 'taken', 'acema', true)
		];
		lookup.cost = ['paid'];
		expect(lookup.visibleCounts).toEqual({
			available: 0,
			taken: 1,
			unknown: 0,
			unsupported: 0
		});
	});
});

describe('counts', () => {
	it('counts verdicts and not-checked rows separately', () => {
		const lookup = new Lookup();
		lookup.rows = [
			row('GitHub', 'taken'),
			row('Lemmy', 'available'),
			row('PeerTube', 'unknown'),
			row('TikTok', 'unsupported'),
			row('npm', 'pending')
		];

		expect(lookup.counts).toEqual({ available: 1, taken: 1, unknown: 1, unsupported: 1 });
	});

	it('narrows status counts to the selected cost', () => {
		const lookup = new Lookup();
		lookup.rows = [
			row('Domains', 'taken', 'acema', true),
			row('GitHub', 'available'),
			row('Lemmy', 'taken')
		];
		lookup.cost = ['paid'];

		expect(lookup.counts).toEqual({ available: 0, taken: 1, unknown: 0, unsupported: 0 });
		expect(lookup.statusTotal).toBe(1);
	});

	it('narrows cost counts to the status filter', () => {
		const lookup = new Lookup();
		lookup.rows = [
			row('Domains', 'taken', 'acema', true),
			row('GitHub', 'available'),
			row('Lemmy', 'taken')
		];
		lookup.filter = ['taken'];

		expect(lookup.costCounts).toEqual({ free: 1, paid: 1 });
	});
});

describe('announcement', () => {
	it('summarises the selected counts', () => {
		const lookup = new Lookup();
		lookup.rows = [row('GitHub', 'taken'), row('Lemmy', 'available'), row('PeerTube', 'unknown')];
		expect(lookup.announcement).toBe('3 shown. 1 available, 1 unavailable, 1 unknown');
	});

	it('reports when no status is selected', () => {
		const lookup = new Lookup();
		lookup.rows = [row('GitHub', 'taken')];
		lookup.filter = [];
		expect(lookup.announcement).toBe('No statuses selected.');
	});

	it('prefixes the shown count when a filter is active', () => {
		const lookup = new Lookup();
		lookup.rows = [row('GitHub', 'taken'), row('Lemmy', 'available'), row('PeerTube', 'unknown')];
		lookup.filter = ['available', 'unknown'];
		expect(lookup.announcement).toBe('2 shown. 1 available, 0 unavailable, 1 unknown');
	});

	it('summarises only the rows left by the cost filter', () => {
		const lookup = new Lookup();
		lookup.rows = [
			row('GitHub', 'taken'),
			row('Lemmy', 'available'),
			row('Domains', 'taken', 'acema', true)
		];
		lookup.cost = ['paid'];
		expect(lookup.announcement).toBe('1 shown. 0 available, 1 unavailable, 0 unknown');
	});

	it('mentions visible not-checked sources', () => {
		const lookup = new Lookup();
		lookup.rows = [row('GitHub', 'taken'), row('TikTok', 'unsupported')];
		lookup.filter = ['taken', 'unsupported'];
		expect(lookup.announcement).toBe(
			'2 shown. 0 available, 1 unavailable, 0 unknown, 1 not checked'
		);
	});
});

describe('check', () => {
	it('aborts the in-flight scan when a new check starts', () => {
		const signals: AbortSignal[] = [];
		const client: CheckClient = {
			stream(_label, _onEvent, signal) {
				return new Promise((_resolve, reject) => {
					if (!signal) return;
					signals.push(signal);
					signal.addEventListener('abort', () => reject(new DOMException('aborted', 'AbortError')));
				});
			}
		};

		const lookup = new Lookup(client);
		lookup.input = 'first';
		lookup.check({ sync: false });
		lookup.input = 'second';
		lookup.check({ sync: false });

		expect(signals).toHaveLength(2);
		expect(signals[0].aborted).toBe(true);
		expect(signals[1].aborted).toBe(false);

		lookup.goHome();
		expect(signals[1].aborted).toBe(true);
	});

	it('keeps the previous row slots pending until the manifest arrives', async () => {
		vi.stubGlobal('fetch', stubApi(''));
		const lookup = new Lookup();
		lookup.input = 'lumen';
		lookup.lastLabel = 'acema';
		lookup.rows = [row('GitHub', 'taken'), row('Lemmy', 'available'), row('TikTok', 'unsupported')];

		lookup.check({ sync: false });

		expect(lookup.rows.map((r) => r.id)).toEqual(['github-acema', 'lemmy-acema', 'tiktok-acema']);
		expect(lookup.rows[0].status).toBe('pending');
		expect(lookup.rows[0].url).toBe('');
		expect(lookup.rows[0].name).toBe('');
		expect(lookup.rows[2].status).toBe('unsupported');

		await vi.waitFor(() => expect(lookup.loading).toBe(false));
	});

	it('seeds every source from the manifest and fills results by id', async () => {
		const body =
			[
				manifest([
					{ id: 'npm', platformName: 'npm', url: 'https://npmjs.com/acema' },
					{ id: 'lemmy', platformName: 'Lemmy', url: 'https://lemmy.world/u/acema' },
					{
						id: 'tiktok',
						platformName: 'TikTok',
						url: 'https://tiktok.com/@acema',
						checked: false
					}
				]),
				result('npm', 'npm', 'available'),
				result('lemmy', 'Lemmy', 'taken'),
				JSON.stringify({
					type: 'done',
					counts: { available: 1, taken: 1, unknown: 0, unsupported: 1 }
				})
			].join('\n') + '\n';
		vi.stubGlobal('fetch', stubApi(body));

		const lookup = new Lookup();
		lookup.input = 'acema';
		lookup.check({ sync: false });

		await vi.waitFor(() => expect(lookup.loading).toBe(false));

		expect(lookup.rows).toHaveLength(3);
		expect(lookup.rows.map((r) => [r.id, r.status])).toEqual([
			['npm', 'available'],
			['lemmy', 'taken'],
			['tiktok', 'unsupported']
		]);
		expect(lookup.rows.every((r) => r.name === 'acema')).toBe(true);
		expect(lookup.rows[2].url).toBe('https://tiktok.com/@acema');
		expect(lookup.counts).toEqual({ available: 1, taken: 1, unknown: 0, unsupported: 1 });
	});

	it('streams from an injected client, independent of HTTP', async () => {
		const lines = [
			manifest([{ id: 'npm', platformName: 'npm', url: 'https://npmjs.com/acema' }]),
			result('npm', 'npm', 'available')
		];
		const client: CheckClient = {
			async stream(_label, onEvent) {
				for (const line of lines) {
					const event = parseEvent(line);
					if (event) onEvent(event);
				}
			}
		};

		const lookup = new Lookup(client);
		lookup.input = 'acema';
		lookup.check({ sync: false });
		await vi.waitFor(() => expect(lookup.loading).toBe(false));

		expect(lookup.rows).toHaveLength(1);
		expect(lookup.rows[0].status).toBe('available');
		expect(lookup.counts.available).toBe(1);
	});

	it('uses the label from the manifest for row names', async () => {
		vi.stubGlobal(
			'fetch',
			stubApi(
				manifest([{ id: 'npm', platformName: 'npm', url: 'https://npmjs.com/acema' }], 'acema') +
					'\n' +
					result('npm', 'npm', 'available')
			)
		);

		const lookup = new Lookup();
		lookup.input = 'acema';
		lookup.check({ sync: false });
		await vi.waitFor(() => expect(lookup.loading).toBe(false));

		expect(lookup.rows[0].name).toBe('acema');
	});

	it('marks unresolved rows unreachable when the stream ends early', async () => {
		vi.stubGlobal(
			'fetch',
			stubApi(manifest([{ id: 'npm', platformName: 'npm', url: 'https://npmjs.com/acema' }]) + '\n')
		);

		const lookup = new Lookup();
		lookup.input = 'acema';
		lookup.check({ sync: false });
		await vi.waitFor(() => expect(lookup.loading).toBe(false));

		expect(lookup.rows[0].status).toBe('unknown');
		expect(lookup.rows[0].detail).toBe('unreachable');
	});

	it('clears retained slots and surfaces the error when the stream fails before the manifest', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn(() => Promise.reject(new TypeError('Failed to fetch')))
		);

		const lookup = new Lookup();
		lookup.input = 'acema';
		lookup.lastLabel = 'quartz';
		lookup.rows = [row('GitHub', 'taken')];
		lookup.check({ sync: false });
		await vi.waitFor(() => expect(lookup.loading).toBe(false));

		expect(lookup.rows).toEqual([]);
		expect(lookup.error).toContain('Could not reach the backend');
	});
});

describe('invalid input', () => {
	it('suggests the bare label for a dotted input', () => {
		const lookup = new Lookup();
		lookup.input = 'Acema.com';
		lookup.check({ sync: false });

		expect(lookup.error).toContain('without the dot');
		expect(lookup.suggestion).toBe('acema');
	});

	it('offers no suggestion when nothing safe can be extracted', () => {
		const lookup = new Lookup();
		lookup.input = 'a b';
		lookup.check({ sync: false });

		expect(lookup.error).toContain('letters, numbers and hyphens');
		expect(lookup.suggestion).toBeNull();
	});

	it('clears the error and suggestion when the user edits', () => {
		const lookup = new Lookup();
		lookup.input = 'acema.com';
		lookup.check({ sync: false });

		lookup.clearError();

		expect(lookup.error).toBe('');
		expect(lookup.suggestion).toBeNull();
	});

	it('accepts the suggestion and starts the corrected scan', async () => {
		vi.stubGlobal('fetch', stubApi(''));
		const lookup = new Lookup();
		lookup.input = 'acema.com';
		lookup.check({ sync: false });

		lookup.acceptSuggestion({ sync: false });

		expect(lookup.input).toBe('acema');
		expect(lookup.suggestion).toBeNull();
		expect(lookup.lastLabel).toBe('acema');
		await vi.waitFor(() => expect(lookup.loading).toBe(false));
	});
});

describe('showAll', () => {
	it('clears the cost filter and includes deferred sources', () => {
		const lookup = new Lookup();
		lookup.filter = ['taken'];
		lookup.cost = ['paid'];

		lookup.showAll();

		expect(lookup.filter).toEqual(ALL_STATUSES);
		expect(lookup.cost).toEqual(DEFAULT_COST);
	});
});

describe('goHome', () => {
	it('clears the report and returns to the landing state', () => {
		const lookup = new Lookup();
		lookup.input = 'acema';
		lookup.lastLabel = 'acema';
		lookup.rows = [row('GitHub', 'taken')];
		lookup.error = 'unreachable';
		lookup.suggestion = 'acema';
		lookup.filter = ['taken'];
		lookup.cost = ['paid'];
		lookup.loading = true;

		lookup.goHome();

		expect(lookup.input).toBe('');
		expect(lookup.lastLabel).toBeNull();
		expect(lookup.rows).toEqual([]);
		expect(lookup.error).toBe('');
		expect(lookup.suggestion).toBeNull();
		expect(lookup.filter).toEqual(DEFAULT_FILTER);
		expect(lookup.cost).toEqual(DEFAULT_COST);
		expect(lookup.loading).toBe(false);
		expect(lookup.visibleRows).toEqual([]);
	});
});
