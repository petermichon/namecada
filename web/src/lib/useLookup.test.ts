import { afterEach, describe, expect, it, vi } from 'vitest';
import { byPaidThenPlatformName, byPlatformName, Lookup, type Row } from './useLookup.svelte';
import type { Status } from './platforms';

function row(platformName: string, status: Row['status'], name = 'acema', paid = false): Row {
	return {
		id: `${platformName.toLowerCase()}-${name}`,
		platform: platformName.toLowerCase(),
		platformName,
		name,
		url: `https://example.com/${platformName}/${name}`,
		status,
		paid
	};
}

// providerRow mirrors the backend: one row per provider, keyed by the stable
// provider id rather than the label being looked up.
function providerRow(id: string, platformName: string, status: Status, name = 'acema'): Row {
	return {
		id,
		platform: id,
		platformName,
		name,
		url: `https://example.com/${id}/${name}`,
		status,
		paid: false
	};
}

function providerResult(id: string, platformName: string, status: Status, name = 'acema'): string {
	return JSON.stringify({
		id,
		platform: id,
		platformName,
		name,
		url: `https://example.com/${id}/${name}`,
		status,
		paid: false
	});
}

function stubApi(body: string) {
	return vi.fn((input: unknown) => {
		if (String(input).includes('/check?name=')) {
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

describe('byPaidThenPlatformName', () => {
	it('puts paid rows before free ones', () => {
		expect(
			byPaidThenPlatformName(row('Domains', 'taken', 'acema', true), row('GitHub', 'taken'))
		).toBeLessThan(0);
		expect(
			byPaidThenPlatformName(row('GitHub', 'taken'), row('Domains', 'taken', 'acema', true))
		).toBeGreaterThan(0);
	});

	it('falls back to the display name within a group', () => {
		expect(byPaidThenPlatformName(row('GitHub', 'taken'), row('Lemmy', 'taken'))).toBeLessThan(0);
		expect(
			byPaidThenPlatformName(row('Domains', 'taken', 'a', true), row('Domains', 'taken', 'b', true))
		).toBeLessThan(0);
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
		lookup.filter = 'taken';
		expect(lookup.visibleRows.map((r) => r.platformName)).toEqual(['GitHub', 'PeerTube']);
	});

	it('breaks platformName ties by URL, not arrival order', () => {
		const lookup = new Lookup();
		lookup.rows = [row('Lemmy', 'taken', 'b'), row('Lemmy', 'available', 'a')];
		expect(lookup.visibleRows.map((r) => r.name)).toEqual(['a', 'b']);
	});

	it('puts paid rows before free rows', () => {
		const lookup = new Lookup();
		lookup.rows = [
			row('npm', 'taken'),
			row('Domains', 'taken', 'acema', true),
			row('GitHub', 'taken')
		];
		expect(lookup.visibleRows.map((r) => r.platformName)).toEqual(['Domains', 'GitHub', 'npm']);
	});

	it('groups rows into paid and free', () => {
		const lookup = new Lookup();
		lookup.rows = [
			row('npm', 'taken'),
			row('Domains', 'taken', 'acema', true),
			row('GitHub', 'taken')
		];
		expect(lookup.paidRows.map((r) => r.platformName)).toEqual(['Domains']);
		expect(lookup.freeRows.map((r) => r.platformName)).toEqual(['GitHub', 'npm']);
	});

	it('applies the status filter within the groups', () => {
		const lookup = new Lookup();
		lookup.rows = [
			row('npm', 'available'),
			row('Domains', 'taken', 'acema', true),
			row('GitHub', 'taken')
		];
		lookup.filter = 'taken';
		expect(lookup.paidRows.map((r) => r.platformName)).toEqual(['Domains']);
		expect(lookup.freeRows.map((r) => r.platformName)).toEqual(['GitHub']);
	});
});

describe('stale', () => {
	it('is false while results are still pending', () => {
		const lookup = new Lookup();
		lookup.input = 'lumen';
		lookup.lastLabel = 'acema';
		lookup.rows = [row('GitHub', 'pending')];
		expect(lookup.stale).toBe(false);
	});

	it('is false with no rows', () => {
		const lookup = new Lookup();
		lookup.input = 'lumen';
		lookup.lastLabel = 'acema';
		expect(lookup.stale).toBe(false);
	});

	it('is true once a renamed input settles', () => {
		const lookup = new Lookup();
		lookup.input = 'lumen';
		lookup.lastLabel = 'acema';
		lookup.rows = [row('GitHub', 'taken')];
		expect(lookup.stale).toBe(true);
	});

	it('is false when the input matches the last check', () => {
		const lookup = new Lookup();
		lookup.input = 'acema';
		lookup.lastLabel = 'acema';
		lookup.rows = [row('GitHub', 'taken')];
		expect(lookup.stale).toBe(false);
	});
});

describe('check', () => {
	it('keeps the previous rows mounted and flags busy on a re-scan', async () => {
		vi.stubGlobal('fetch', stubApi(''));
		const lookup = new Lookup();
		lookup.input = 'lumen';
		lookup.lastLabel = 'acema';
		lookup.rows = [
			providerRow('github', 'GitHub', 'taken'),
			providerRow('lemmy', 'Lemmy', 'available')
		];

		lookup.check({ sync: false });

		// The old list stays on screen (height-stable, no scroll clamp) and is
		// dimmed/aria-busy until the first fresh result arrives.
		expect(lookup.rows.map((r) => r.id)).toEqual(['github', 'lemmy']);
		expect(lookup.busy).toBe(true);
		expect(lookup.searching).toBe(true);

		await vi.waitFor(() => expect(lookup.loading).toBe(false));
	});

	it('overwrites stale rows by id and clears busy once results arrive', async () => {
		const body =
			[
				providerResult('github', 'GitHub', 'available'),
				providerResult('lemmy', 'Lemmy', 'taken')
			].join('\n') + '\n';
		vi.stubGlobal('fetch', stubApi(body));
		const lookup = new Lookup();
		lookup.input = 'lumen';
		lookup.lastLabel = 'acema';
		lookup.rows = [
			providerRow('github', 'GitHub', 'taken'),
			providerRow('lemmy', 'Lemmy', 'available')
		];

		lookup.check({ sync: false });
		await vi.waitFor(() => expect(lookup.loading).toBe(false));

		expect(lookup.rows.map((r) => r.id)).toEqual(['github', 'lemmy']);
		expect(lookup.rows.map((r) => r.status)).toEqual(['available', 'taken']);
		expect(lookup.busy).toBe(false);
		expect(lookup.checked).toBe(2);
	});

	it('drops stale rows and surfaces the error when a re-scan fails', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn(() => Promise.reject(new TypeError('Failed to fetch')))
		);
		const lookup = new Lookup();
		lookup.input = 'lumen';
		lookup.lastLabel = 'acema';
		lookup.rows = [providerRow('github', 'GitHub', 'taken')];

		lookup.check({ sync: false });
		await vi.waitFor(() => expect(lookup.loading).toBe(false));

		expect(lookup.rows).toEqual([]);
		expect(lookup.error).toContain('Could not reach the backend');
	});
});

describe('copyResults', () => {
	it('copies rows ordered alphabetically by platformName', async () => {
		const written: string[] = [];
		vi.stubGlobal('navigator', {
			clipboard: {
				writeText: async (text: string) => {
					written.push(text);
				}
			}
		});

		const lookup = new Lookup();
		lookup.lastLabel = 'acema';
		lookup.rows = [row('PeerTube', 'available'), row('GitHub', 'taken'), row('Lemmy', 'taken')];
		await lookup.copyResults();

		expect(written[0]).toBe(
			'acema\nGitHub\tacema\ttaken\nLemmy\tacema\ttaken\nPeerTube\tacema\tavailable'
		);
	});
});
