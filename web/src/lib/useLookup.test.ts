import { afterEach, describe, expect, it, vi } from 'vitest';
import { byPaidThenPlatformName, byPlatformName, Lookup, type Row } from './useLookup.svelte';

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
		).toBe(0);
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
