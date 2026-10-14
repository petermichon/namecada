import { afterEach, describe, expect, it, vi } from 'vitest';
import { byPlatformName, Lookup, type Row } from './useLookup.svelte';

function row(platformName: string, status: Row['status'], name = 'acema'): Row {
	return {
		id: `${platformName.toLowerCase()}-${name}`,
		platform: platformName.toLowerCase(),
		platformName,
		name,
		url: `https://example.com/${platformName}/${name}`,
		status
	};
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

	it('is stable for rows sharing a platformName', () => {
		const lookup = new Lookup();
		lookup.rows = [row('Lemmy', 'taken', 'b'), row('Lemmy', 'available', 'a')];
		expect(lookup.visibleRows.map((r) => r.name)).toEqual(['b', 'a']);
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
