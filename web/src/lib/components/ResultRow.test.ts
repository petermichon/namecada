import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import ResultRow from './ResultRow.svelte';
import type { Row } from '$lib/useLookup.svelte';

function row(status: Row['status'], url = 'https://example.com/acema'): Row {
	return {
		id: `github-${status}`,
		platform: 'github',
		platformName: 'GitHub',
		name: 'acema',
		url,
		status,
		paid: false
	};
}

function statusSlot(row: Row): string {
	return render(ResultRow, { props: { row } }).body;
}

describe('ResultRow', () => {
	it('labels unsupported rows as not checked with an icon, not a verdict', () => {
		const html = statusSlot(row('unsupported'));

		expect(html).toContain('not checked');
		expect(html).not.toContain('>available<');
		expect(html).not.toContain('unavailable');
		expect(html).not.toContain('unknown');
		expect(html).toContain('<svg');
	});

	it('still labels the verdict statuses', () => {
		expect(statusSlot(row('available'))).toContain('available');
		expect(statusSlot(row('taken'))).toContain('unavailable');
		expect(statusSlot(row('unknown'))).toContain('unknown');
	});

	it('carries the grid separators as borders on the row', () => {
		const html = statusSlot(row('available'));
		const li = html.match(/<li[^>]*class="([^"]*)"/)?.[1] ?? '';

		expect(li).toContain('bg-card');
		expect(li).toContain('border-b');
		expect(li).toContain('border-border');
		expect(li).toContain('lg:border-l');
	});
});
