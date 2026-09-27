import { describe, expect, it } from 'vitest';
import { parseBootstrap } from './bootstrap';

describe('parseBootstrap', () => {
	it('maps TLDs to their registry base URL, normalizing the trailing slash', () => {
		const map = parseBootstrap({
			services: [
				[['com', 'net'], ['https://rdap.verisign.com/com/v1/']],
				[['test'], ['https://rdap.test']]
			]
		});
		expect(map.get('com')).toBe('https://rdap.verisign.com/com/v1/');
		expect(map.get('net')).toBe('https://rdap.verisign.com/com/v1/');
		expect(map.get('test')).toBe('https://rdap.test/');
	});

	it('ignores malformed services', () => {
		const map = parseBootstrap({
			services: [['io', ['https://rdap.example/']], [['no-urls'], []], 'nope', null]
		});
		expect(map.size).toBe(0);
	});

	it('returns an empty map for non-object input', () => {
		expect(parseBootstrap(null).size).toBe(0);
		expect(parseBootstrap({}).size).toBe(0);
	});
});
