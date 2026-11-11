import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import ResultRow from './ResultRow.svelte';
import type { Row } from '$lib/useLookup.svelte';

function row(status: Row['status'], url = 'https://example.com/acema', paid = false): Row {
	return {
		id: `github-${status}`,
		platform: 'github',
		platformName: 'GitHub',
		name: 'acema',
		url,
		status,
		paid,
		tags: ['social']
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

	it('uses the same rounded row treatment as the status filters', () => {
		const html = statusSlot(row('available'));
		const anchor = html.match(/<a[^>]*class="([^"]*)"/)?.[1] ?? '';

		expect(anchor).toContain('rounded-md');
		expect(anchor).toContain('px-2');
		expect(anchor).toContain('hover:bg-');
		expect(html).not.toContain('border-b');
	});

	it('marks only paid sources', () => {
		expect(statusSlot(row('available'))).not.toContain('Paid');
		expect(statusSlot(row('available', undefined, true))).toContain('Paid');
	});
});
