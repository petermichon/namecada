import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import ResultsList from './ResultsList.svelte';
import type { Row } from '$lib/rows';

function row(id: string, status: Row['status'] = 'available', paid = false): Row {
	return {
		id,
		platform: 'github',
		platformName: 'GitHub',
		name: id,
		url: `https://example.com/${id}`,
		status,
		paid
	};
}

function firstUlClass(html: string): string {
	return html.match(/<ul[^>]*class="([^"]*)"/)?.[1] ?? '';
}

function liCount(html: string): number {
	return [...html.matchAll(/<li[\s>]/g)].length;
}

describe('ResultsList', () => {
	it('renders every row in one list without paid or free sections', () => {
		const { body } = render(ResultsList, {
			props: { rows: [row('a', 'available', true), row('b', 'taken'), row('c')] }
		});

		expect(liCount(body)).toBe(3);
		expect(body).not.toContain('Paid platforms');
		expect(body).not.toContain('Free platforms');
	});

	it('keeps the responsive two-column layout and rounded rows', () => {
		const { body } = render(ResultsList, { props: { rows: [row('a'), row('b')] } });
		const ul = firstUlClass(body);

		expect(ul).toContain('grid-cols-1');
		expect(ul).toContain('lg:grid-cols-2');
		expect(ul).toContain('gap-y-0.5');
		expect(body).toContain('rounded-md');
		expect(body).not.toContain('border-b');
	});

	it('renders an empty list when nothing matches', () => {
		const { body } = render(ResultsList, { props: { rows: [] } });

		expect(liCount(body)).toBe(0);
		expect(body).toContain('<ul');
	});
});
