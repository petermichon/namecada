import { describe, expect, it } from 'vitest';
import { extractDetails, isNotableStatus, nameserverHint } from './details';

const raw = {
	handle: '2138514_DOMAIN_COM-VRSN',
	status: ['client transfer prohibited', 'server delete prohibited'],
	events: [
		{ eventAction: 'registration', eventDate: '1997-09-15T04:00:00Z' },
		{ eventAction: 'expiration', eventDate: '2028-09-14T04:00:00Z' }
	],
	nameservers: [{ ldhName: 'NS1.GOOGLE.COM' }, { ldhName: 'NS2.GOOGLE.COM' }],
	secureDNS: { delegationSigned: true },
	links: [
		{ rel: 'self', href: 'https://rdap.verisign.com/com/v1/domain/google.com' },
		{ rel: 'related', href: 'https://rdap.markmonitor.com/domain/GOOGLE.COM' }
	],
	entities: [
		{
			roles: ['registrar'],
			vcardArray: [
				'vcard',
				[
					['fn', {}, 'text', 'MarkMonitor Inc.'],
					['version', {}, 'text', '4.0']
				]
			],
			publicIds: [{ type: 'IANA Registrar ID', identifier: '292' }],
			entities: [
				{
					roles: ['abuse'],
					vcardArray: ['vcard', [['email', {}, 'text', 'abusecomplaints@markmonitor.com']]]
				}
			]
		}
	]
};

describe('extractDetails', () => {
	it('extracts every field from a domain RDAP object', () => {
		const details = extractDetails(raw);
		expect(details).toEqual({
			statuses: ['client transfer prohibited', 'server delete prohibited'],
			events: [
				{ action: 'registration', date: '1997-09-15T04:00:00Z' },
				{ action: 'expiration', date: '2028-09-14T04:00:00Z' }
			],
			registrar: {
				name: 'MarkMonitor Inc.',
				id: '292',
				abuseEmail: 'abusecomplaints@markmonitor.com'
			},
			nameservers: ['NS1.GOOGLE.COM', 'NS2.GOOGLE.COM'],
			dnssec: true,
			related: 'https://rdap.markmonitor.com/domain/GOOGLE.COM'
		});
	});

	it('returns empty collections for a minimal object', () => {
		expect(extractDetails({})).toEqual({ statuses: [], events: [], nameservers: [] });
	});

	it('returns null for non-object input', () => {
		expect(extractDetails(null)).toBeNull();
		expect(extractDetails('nope')).toBeNull();
	});
});

describe('isNotableStatus', () => {
	it('flags imminent-availability and problem statuses', () => {
		expect(isNotableStatus('pending delete')).toBe(true);
		expect(isNotableStatus('redemptionPeriod')).toBe(true);
		expect(isNotableStatus('autoRenewPeriod')).toBe(true);
		expect(isNotableStatus('clientHold')).toBe(true);
	});

	it('treats routine locks as not notable', () => {
		expect(isNotableStatus('client transfer prohibited')).toBe(false);
		expect(isNotableStatus('client delete prohibited')).toBe(false);
		expect(isNotableStatus('ok')).toBe(false);
	});
});

describe('nameserverHint', () => {
	it('flags no nameservers as not delegated', () => {
		expect(nameserverHint([])).toEqual({ kind: 'none' });
	});

	it('detects known parking providers', () => {
		expect(nameserverHint(['ns1.sedoparking.com'])).toEqual({ kind: 'parked', provider: 'Sedo' });
		expect(nameserverHint(['ns1.google.com', 'ns2.dan.com'])).toEqual({
			kind: 'parked',
			provider: 'Dan.com'
		});
	});

	it('returns null for ordinary delegated nameservers', () => {
		expect(nameserverHint(['ns1.google.com', 'ns2.google.com'])).toBeNull();
	});
});
